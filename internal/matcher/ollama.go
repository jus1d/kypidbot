package matcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type embeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func getEmbedding(text string, ollamaURL string) ([]float64, error) {
	reqBody := embeddingRequest{
		Model:  "paraphrase-multilingual",
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	resp, err := http.Post(ollamaURL+"/api/embeddings", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("unmarshal embedding response: %w", err)
	}

	return embResp.Embedding, nil
}

func getEmbeddings(texts []string, ollamaURL string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))

	for i, text := range texts {
		emb, err := getEmbedding(text, ollamaURL)
		if err != nil {
			return nil, fmt.Errorf("get embedding for text %d: %w", i, err)
		}
		embeddings[i] = emb
	}

	return embeddings, nil
}
