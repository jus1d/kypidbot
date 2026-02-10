package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PullRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

func (c *Client) PullModel() error {
	reqBody := PullRequest{
		Model:  c.model,
		Stream: false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal pull request: %w", err)
	}

	resp, err := http.Post(c.url+"/api/pull", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var status StatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		return fmt.Errorf("unmarshal pull status: %w", err)
	}

	if status.Status != "success" {
		return fmt.Errorf("pull failed with status: %s", status.Status)
	}

	return nil
}
