package view

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var messages map[string]any

func LoadMessages(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read messages file: %w", err)
	}

	if err := yaml.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("unmarshal messages: %w", err)
	}

	return nil
}

func Msg(keys ...string) string {
	var current any = messages

	for _, key := range keys {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = m[key]
	}

	s, ok := current.(string)
	if !ok {
		return ""
	}

	return strings.TrimRight(s, "\n")
}

func Msgf(replacements map[string]string, keys ...string) string {
	s := Msg(keys...)
	for k, v := range replacements {
		s = strings.ReplaceAll(s, "{"+k+"}", v)
	}
	return s
}
