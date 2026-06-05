package config

type LicenseConfig struct {
	CompatibleLicenses []string `koanf:"compatiblelicenses"`

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

// defaultCompatibleLicenses is the set of SPDX licenses considered compatible
// when no list is provided via config file or environment. Please keep in sync with
// config.example.yaml.
var defaultCompatibleLicenses = []string{
	"AFL-1.1", "AFL-1.2", "AFL-2.0", "AFL-2.1", "AFL-3.0",
	"Apache-1.1", "Apache-2.0",
	"Artistic-1.0", "Artistic-1.0-Perl", "Artistic-1.0-cl8", "Artistic-2.0",
	"0BSD", "BSD-1-Clause", "BSD-2-Clause", "BSD-2-Clause-Patent",
	"BSD-3-Clause", "BSD-3-Clause-LBNL",
	"CDDL-1.0",
	"EPL-1.0", "EPL-2.0",
	"ICU", "ISC",
	"MIT", "MIT-0", "MIT-Modern-Variant", "MIT-feh",
	"MPL-1.0", "MPL-1.1", "MPL-2.0", "MPL-2.0-no-copyleft-exception",
	"Unlicense", "Xnet", "Zlib",
}

func (c *LicenseConfig) Validate() error {
	// Fall back to the built-in default list when none is provided.
	if len(c.CompatibleLicenses) == 0 {
		c.CompatibleLicenses = defaultCompatibleLicenses
	}

	if c.ConfidenceThreshold == 0.0 {
		c.ConfidenceThreshold = 0.85
	}
	if c.ConfidenceOverrideThreshold == 0 {
		c.ConfidenceOverrideThreshold = 0.98
	}

	return nil
}
