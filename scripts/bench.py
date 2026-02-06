#!/usr/bin/env python3

import subprocess
import os

BENCH_FOLDER = "./bench"
MODELS = [
    'paraphrase-multilingual',
    'mxbai-embed-large',
    'all-minilm',
    'nomic-embed-text-v2-moe',
]

os.makedirs(BENCH_FOLDER, exist_ok=True)

for model in MODELS:
    print(f"pulling `{model}`")
    cmd = ["docker", "exec", "ollama", "ollama", "pull", model]
    subprocess.run(cmd)

    print(f"testing `{model}`")
    input_path = './data/input.clean.json'
    output_path = f'{BENCH_FOLDER}/output.{model}.json'
    cmd = ["go", "run", "./cmd/matcher", f"-ollama-model={model}", input_path, output_path]
    subprocess.run(cmd)

print(f'tested {len(MODELS)} models')
