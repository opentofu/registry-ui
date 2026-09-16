package config

type LicenseConfig struct {

	// ConfidenceThreshold is the minimum confidence score (0.0–1.0) a license match
	// must reach to be included in the results. Scores are produced by the go-license-detector
	// library, which compares repository files against known SPDX license texts.
	// Matches below this threshold are discarded.
	// Default: 0.85.
	ConfidenceThreshold float32 `koanf:"confidencethreshold"`

	// ConfidenceOverrideThreshold is the confidence score (0.0–1.0) at which a single
	// license match is considered authoritative. When a match meets or exceeds this
	// threshold, it is returned as the sole license and all lower-confidence matches
	// are discarded. This avoids noisy results when a clear license file exists.
	// Default: 0.98.
	ConfidenceOverrideThreshold float32 `koanf:"confidenceoverridethreshold"`
}

func (c *LicenseConfig) Validate() error {
	if c.ConfidenceThreshold == 0.0 {
		c.ConfidenceThreshold = 0.85
	}
	if c.ConfidenceOverrideThreshold == 0 {
		c.ConfidenceOverrideThreshold = 0.98
	}

	return nil
}
