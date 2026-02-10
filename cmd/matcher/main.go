package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jus1d/kypidbot/internal/config"
	"github.com/jus1d/kypidbot/internal/infrastructure/ollama"
	"github.com/jus1d/kypidbot/internal/matcher"
)

type outputPair struct {
	A     string  `json:"a"`
	B     string  `json:"b"`
	Score float64 `json:"score"`
}

func main() {
	ollamaHost := flag.String("ollama-host", "http://localhost", "Ollama host")
	ollamaPort := flag.String("ollama-port", "11434", "Ollama port")
	ollamaModel := flag.String("ollama-model", "nomic-embed-text", "Ollama model")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: matcher [flags] <input.json> <output.json>\n")
		os.Exit(1)
	}

	inputPath := args[0]
	outputPath := args[1]

	data, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("read input file: %v", err)
	}

	var abouts []string
	if err := json.Unmarshal(data, &abouts); err != nil {
		log.Fatalf("parse input file: %v", err)
	}

	ollama := ollama.New(&config.Ollama{
		Host:  *ollamaHost,
		Port:  *ollamaPort,
		Model: *ollamaModel,
	})

	scores, err := matcher.MatchByScore(abouts, ollama)
	if err != nil {
		log.Fatalf("match by score: %v", err)
	}

	output := make([]outputPair, len(scores))
	for i, s := range scores {
		output[i] = outputPair{
			A:     s.A,
			B:     s.B,
			Score: s.Score,
		}
	}

	result, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		log.Fatalf("marshal output: %v", err)
	}

	if err := os.WriteFile(outputPath, result, 0644); err != nil {
		log.Fatalf("write output file: %v", err)
	}

	fmt.Printf("matched %d pairs, written to %s\n", len(scores), outputPath)
}
