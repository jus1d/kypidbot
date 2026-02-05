# kypidbot

## Quick Start

> You can build image locally: `docker build -t ghcr.io/jus1d/kypidbot:latest .` or pull it from `ghcr.io`: `docker pull ghcr.io/jus1d/kypidbot:latest`

Fill up the `.env` file and run the bot

Pull model for  vectorizing abouts

```bash
$ docker exec ollama ollama pull paraphrase-multilingual
```

Export `POSTGRES_PASSWORD` env variable

```bash
$ export POSTGRES_PASSWORD=kypidbot
```

Run compose

```bash
$ docker compose up -d
```
