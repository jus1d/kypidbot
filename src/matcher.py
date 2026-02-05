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
from database import UserModel

logging.getLogger("sentence_transformers").setLevel(logging.ERROR)
logging.getLogger("transformers").setLevel(logging.ERROR)


Pair: TypeAlias = tuple[int, int, float, str]
FullMatch: TypeAlias = tuple[int, int, float]
Preferences: TypeAlias = dict[int, set[str]]


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


def parse_users_from_json(path: str) -> list[UserModel]:
    with open(path) as f:
        data = json.load(f)

    users = []
    for i, u in enumerate(data["users"]):
        user = UserModel(
            id=i,
            telegram_id=0,
            username=u.get("username"),
            first_name=None,
            last_name=None,
            is_bot=False,
            language_code=None,
            is_premium=False,
            sex=u.get("sex"),
            about=u.get("about") or u.get("interests", ""),
            state="completed",
            time_ranges=u.get("time_ranges") or u.get("free_time", "000000"),
            is_admin=False,
        )
        users.append(user)
    return users


def write_pairs_as_json(pairs: list[Pair], users: list[UserModel], path: str):
    output = [
        {
            "a": {
                "username": users[i].username,
                "sex": users[i].sex,
                "about": users[i].about,
                "time_ranges": users[i].time_ranges,
            },
            "b": {
                "username": users[j].username,
                "sex": users[j].sex,
                "about": users[j].about,
                "time_ranges": users[j].time_ranges,
            },
            "score": score,
            "time_intersection": pair_time,
        }
        for i, j, score, pair_time in pairs
    ]

    with open(path, "w") as f:
        json.dump(output, f, ensure_ascii=False, indent=4)


def extract_preferences(users: list[UserModel]) -> Preferences:
    preferences: Preferences = {}
    pattern = r"@(\w+)"

    for i, user in enumerate(users):
        mentions = re.findall(pattern, user.about or "")
        if mentions:
            preferences[i] = set(mentions)

    return preferences


def match_people(users: list[UserModel]) -> tuple[list[Pair], list[FullMatch]]:
    model = SentenceTransformer(
        "paraphrase-multilingual-MiniLM-L12-v2",
        token=HF_TOKEN,
    )

    abouts = [u.about or "" for u in users]
    vectors = model.encode(abouts)

    sim_matrix = cosine_similarity(vectors)

    preferences = extract_preferences(users)

    # firstly -- match people with mutual sympathy
    used = set()
    pairs = []
    full_matches = []

    n = len(abouts)
    for i in range(n):
        if i in used:
            continue
        for j in range(i + 1, n):
            if j in used:
                continue

            a, b = users[i], users[j]

            if a.sex == b.sex:
                continue

            a_wants_b = i in preferences and b.username in preferences[i]
            b_wants_a = j in preferences and a.username in preferences[j]

            if a_wants_b and b_wants_a:  # full match
                pair_time = "".join(
                    "1"
                    if users[i].time_ranges[k] == "1" and users[j].time_ranges[k] == "1"
                    else "0"
                    for k in range(6)
                )

                score = float(sim_matrix[i, j])

                if "1" in pair_time:
                    pairs.append((i, j, round(score, 3), pair_time))
                else:
                    full_matches.append((i, j, round(score, 3)))

                used.add(i)
                used.add(j)
                break

    # match the remaining pairs greedily
    scores = []
    for i in range(n):
        if i in used:
            continue
        for j in range(i + 1, n):
            if j in used:
                continue

            a, b = users[i], users[j]

            if a.sex == b.sex:
                continue

            score = float(sim_matrix[i, j])

            a_wants_b = i in preferences and b.username in preferences[i]
            b_wants_a = j in preferences and a.username in preferences[j]

            if a_wants_b or b_wants_a:  # semi-match
                score += 0.3

            pair_time = "".join(
                "1" if a.time_ranges[k] == "1" and b.time_ranges[k] == "1" else "0"
                for k in range(6)
            )

            scores.append((score, i, j, pair_time))

    scores.sort(reverse=True)

    for score, i, j, pair_time in scores:
        if i not in used and j not in used and "1" in pair_time:
            pairs.append((i, j, round(score, 3), pair_time))
            used.add(i)
            used.add(j)

    return pairs, full_matches


if __name__ == "__main__":
    opts = parse_options()

    users = parse_users_from_json(opts.input_path)

    pairs, full_matches = match_people(users)
    write_pairs_as_json(pairs, users, opts.output_path)
