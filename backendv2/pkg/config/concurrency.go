package config

import "fmt"

type ConcurrencyConfig struct {
	Module    int `koanf:"module"`
	Provider  int `koanf:"provider"`
	Version   int `koanf:"version"`
	Upload    int `koanf:"upload"`
	Submodule int `koanf:"submodule"`
	Example   int `koanf:"example"`
}

func (c *ConcurrencyConfig) Validate() error {
	// Validate all fields - reject negative values
	if c.Provider < 0 {
		return fmt.Errorf("provider concurrency must be greater than or equal to 0")
	}
	if c.Version < 0 {
		return fmt.Errorf("version concurrency must be greater than or equal to 0")
	}
	if c.Upload < 0 {
		return fmt.Errorf("upload concurrency must be greater than or equal to 0")
	}
	if c.Module < 0 {
		return fmt.Errorf("module concurrency must be greater than or equal to 0")
	}
	if c.Submodule < 0 {
		return fmt.Errorf("submodule concurrency must be greater than or equal to 0")
	}
	if c.Example < 0 {
		return fmt.Errorf("example concurrency must be greater than or equal to 0")
	}

	// Set defaults for zero values
	if c.Provider == 0 {
		c.Provider = 10
	}

	if c.Version == 0 {
		c.Version = 10
	}

	if c.Upload == 0 {
		c.Upload = 100
	}

	if c.Module == 0 {
		// The default for module concurrency is slightly lower than providers because there is a considerably
		// higher amount of work done to scrape modules compared to providers.
		c.Module = 5
	}

	if c.Submodule == 0 {
		// Submodules run tofu show which is expensive, so use conservative default
		c.Submodule = 5
	}

	if c.Example == 0 {
		// Examples run tofu show which is expensive, so use conservative default
		c.Example = 5
	}

	return nil
}
