package gemini

import "github.com/jus1d/kypidbot/internal/config"

type Client struct {
	apiKey    string
	model     string
	maxLength int
}

func New(c *config.Gemini) *Client {
	return &Client{
		apiKey:    c.APIKey,
		model:     c.Model,
		maxLength: c.MaxLength,
	}
}
