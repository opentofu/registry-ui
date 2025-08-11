package config

import "fmt"

type LicenseConfig struct {
	CompatibleLicenses  []string `koanf:"compatibleLicenses"`
	ConfidenceThreshold float32  `koanf:"confidenceThreshold"`
	// ConfidenceOverrideThreshold is the limit at which a detected license overrides all other detected licenses.
	ConfidenceOverrideThreshold float32 `koanf:"confidenceOverrideThreshold"`
}

func (c *LicenseConfig) Validate() error {
	// compatible licenses shouldnt be empty
	if c.CompatibleLicenses == nil || len(c.CompatibleLicenses) == 0 {
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
