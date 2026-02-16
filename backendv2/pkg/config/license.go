package config

import "fmt"

type LicenseConfig struct {
	CompatibleLicenses  []string `koanf:"compatiblelicenses"`
	ConfidenceThreshold float32  `koanf:"confidencethreshold"`
	// ConfidenceOverrideThreshold is the limit at which a detected license overrides all other detected licenses.
	ConfidenceOverrideThreshold float32 `koanf:"confidenceoverridethreshold"`
}

func (c *LicenseConfig) Validate() error {
	// compatible licenses shouldn't be empty
	if len(c.CompatibleLicenses) == 0 {
		return fmt.Errorf("compatible license list was empty or not provided")
	}

	if c.ConfidenceThreshold == 0.0 {
		c.ConfidenceThreshold = 0.85
	}
	if c.ConfidenceOverrideThreshold == 0 {
		c.ConfidenceOverrideThreshold = 0.98
	}

	return nil
}
