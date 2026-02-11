package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://generativelanguage.googleapis.com/v1beta"

type embedRequest struct {
	Requests []embedContentRequest `json:"requests"`
}

type embedContentRequest struct {
	Model    string  `json:"model"`
	Content  content `json:"content"`
	TaskType string  `json:"taskType"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type batchEmbedResponse struct {
	Embeddings []embedding `json:"embeddings"`
}

type embedding struct {
	Values []float64 `json:"values"`
}

func (c *Client) GetEmbeddings(texts []string) ([][]float64, error) {
	requests := make([]embedContentRequest, len(texts))
	for i, text := range texts {
		if c.maxLength > 0 {
			runes := []rune(text)
			if len(runes) > c.maxLength {
				text = string(runes[:c.maxLength])
			}
		}

		requests[i] = embedContentRequest{
			Model:    "models/" + c.model,
			Content:  content{Parts: []part{{Text: text}}},
			TaskType: "SEMANTIC_SIMILARITY",
		}
	}

	body, err := json.Marshal(embedRequest{Requests: requests})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:batchEmbedContents", baseURL, c.model)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result batchEmbedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	embeddings := make([][]float64, len(result.Embeddings))
	for i, e := range result.Embeddings {
		embeddings[i] = e.Values
	}

	return embeddings, nil
}
