package ollama

import (
	"fmt"
	"strings"

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
	host := c.Host
	if !strings.HasPrefix(host, "http://") {
		host = fmt.Sprintf("http://%s", host)
	}

	return &Client{
		host:      host,
		port:      c.Port,
		model:     c.Model,
		url:       fmt.Sprintf("%s:%s", host, c.Port),
		maxLength: c.MaxLength,
	}
}
