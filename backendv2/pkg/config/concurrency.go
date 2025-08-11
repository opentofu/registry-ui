package config

import "fmt"

type ConcurrencyConfig struct {
	Provider int `koanf:"provider"`
	Version  int `koanf:"version"`
	Upload   int `koanf:"upload"`
}

func (c *ConcurrencyConfig) Validate() error {
	if c.Provider < 0 {
		return fmt.Errorf("provider concurrency must be greater than 0")
	}
	if c.Version <= 0 {
		return fmt.Errorf("version concurrency must be greater than 0")
	}
	if c.Upload <= 0 {
		return fmt.Errorf("upload concurrency must be greater than 0")
	}

	if c.Provider == 0 {
		c.Provider = 10
	}

	if c.Version == 0 {
		c.Version = 10
	}

	if c.Upload == 0 {
		c.Upload = 100
	}
	return nil
}
