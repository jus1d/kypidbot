package ollama

import (
	"fmt"

	"github.com/jus1d/kypidbot/internal/config"
)

type Client struct {
	host      string
	port      string
	model     string
	url       string
	maxLength int
}

func New(c *config.Ollama) *Client {
	return &Client{
		host:      c.Host,
		port:      c.Port,
		model:     c.Model,
		url:       fmt.Sprintf("%s:%s", c.Host, c.Port),
		maxLength: c.MaxLength,
	}
}
