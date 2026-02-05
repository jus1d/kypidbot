package config

import (
	"fmt"
	"os"
)

type Config struct {
	TelegramToken string
	OllamaURL     string
	DatabaseURL   string
}

func MustLoad() Config {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		panic(fmt.Sprintf("TELEGRAM_TOKEN is required"))
	}

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/kypidbot?sslmode=disable"
	}

	return Config{
		TelegramToken: token,
		OllamaURL:     ollamaURL,
		DatabaseURL:   databaseURL,
	}
}
