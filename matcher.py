#!/usr/bin/env python3


import json
import logging
import os
import re
import sys
from dataclasses import dataclass
from typing import TypeAlias

os.environ["TQDM_DISABLE"] = "1"

from sentence_transformers import SentenceTransformer
from sklearn.metrics.pairwise import cosine_similarity

from config import HF_TOKEN

logging.getLogger("sentence_transformers").setLevel(logging.ERROR)
logging.getLogger("transformers").setLevel(logging.ERROR)


Pair: TypeAlias = tuple[int, int, float]


@dataclass
class User:
    username: str
    sex: str
    interests: str


@dataclass
class Settings:
    input_path: str
    output_path: str


def parse_options() -> Settings:
    args = sys.argv[1:]

    if not args:
        print("Usage: match input.json [-o output.json]", file=sys.stderr)
        sys.exit(1)

    output_path = "output.json"
    input_path = None

    i = 0
    while i < len(args):
        arg = args[i]

        if arg == "-o":
            if i + 1 >= len(args):
                print("Error: -o requires an argument", file=sys.stderr)
                sys.exit(1)
            output_path = args[i + 1]
            i += 2
        else:
            if input_path is not None:
                print("Error: multiple input files specified", file=sys.stderr)
                sys.exit(1)
            input_path = arg
            i += 1

    if input_path is None:
        print("Error: input file not specified", file=sys.stderr)
        sys.exit(1)

    return Settings(input_path=input_path, output_path=output_path)


def parse_users_from_json(path: str) -> list[User]:
    with open(path) as f:
        data = json.load(f)

    return [User(**user_data) for user_data in data["users"]]


def write_pairs_as_json(pairs: list[Pair], path: str):
    output = [
        {
            "a": {
                "username": users[i].username,
                "sex": users[i].sex,
                "interests": users[i].interests,
            },
            "b": {
                "username": users[j].username,
                "sex": users[j].sex,
                "interests": users[j].interests,
            },
            "score": score,
        }
        for i, j, score in pairs
    ]

    with open(path, "w") as f:
        json.dump(output, f, ensure_ascii=False, indent=4)


def extract_preferences(users: list[User]) -> dict[int, set[str]]:
    """Extract user preferences from interests."""
    preferences = {}
    pattern = r"@(\w+)"

    for i, user in enumerate(users):
        mentions = re.findall(pattern, user.interests)
        if mentions:
            preferences[i] = set(mentions)

    return preferences


def match_people(interests: list[str], users: list[User]) -> list[Pair]:
    model = SentenceTransformer(
        "paraphrase-multilingual-MiniLM-L12-v2",
        token=HF_TOKEN,
    )
    vectors = model.encode(interests)

    sim_matrix = cosine_similarity(vectors)

    preferences = extract_preferences(users)

    n = len(interests)
    scores = []
    for i in range(n):
        for j in range(i + 1, n):
            a, b = users[i], users[j]

            # TODO: match `sex` by preferences. Introduce `looking_for` field for `User` dataclass
            if a.sex == b.sex:
                continue

            base_score = float(sim_matrix[i, j])

            a_wants_b = i in preferences and b.username in preferences[i]
            b_wants_a = j in preferences and a.username in preferences[j]

            if a_wants_b and b_wants_a:
                base_score += 0.5
            elif a_wants_b or b_wants_a:
                base_score += 0.3

            scores.append((base_score, i, j))

    scores.sort(reverse=True)

    used = set()
    pairs = []

    for score, i, j in scores:
        if i not in used and j not in used:
            pairs.append((i, j, round(score, 3)))
            used.add(i)
            used.add(j)

    return pairs


if __name__ == "__main__":
    opts = parse_options()

    users = parse_users_from_json(opts.input_path)
    interests = [u.interests for u in users]

    pairs = match_people(interests, users)
    write_pairs_as_json(pairs, opts.output_path)
