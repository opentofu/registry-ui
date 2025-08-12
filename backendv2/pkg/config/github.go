package config

import (
	"fmt"
)

type GitHubConfig struct {
	Token string `koanf:"token"`
}

func (c *GitHubConfig) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("github.token is required")
	}
	return nil
}