package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func (c *Client) GetEmbedding(text string) ([]float64, error) {
	req := EmbeddingRequest{
		Model:  c.model,
		Prompt: text,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	resp, err := http.Post(c.url+"/api/embeddings", "application/json", bytes.NewBuffer(data))
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

	var embResp EmbeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("unmarshal embedding response: %w", err)
	}

	return embResp.Embedding, nil
}

func (c *Client) GetEmbeddings(texts []string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))

	for i, text := range texts {
		e, err := c.GetEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("get embedding for text %d: %w", i, err)
		}
		embeddings[i] = e
	}

	return embeddings, nil
}
