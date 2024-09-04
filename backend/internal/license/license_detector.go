package license

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
)

type Detector interface {
	Detect(ctx context.Context, repository fs.ReadDirFS, detectOptions ...DetectOpt) (List, error)
}

// List is a list of licenses found in a repository.
//
// swagger:model LicenseList
type List []License

func (l List) HasCompatible() bool {
	for _, license := range l {
		if license.IsCompatible {
			return true
		}
	}
	return false
}

func (l List) HasIncompatible() bool {
	for _, license := range l {
		if !license.IsCompatible {
			return true
		}
	}
	return false
}

func (l List) IsRedistributable() bool {
	// We check for incompatible licenses to avoid mistaking a license in a subdirectory for the main license
	// of the project.
	return len(l) > 0 && !l.HasIncompatible()
}

func (l List) Explain() string {
	if l.IsRedistributable() {
		licenses := map[string]struct{}{}
		for _, license := range l {
			licenses[license.SPDX] = struct{}{}
		}
		var licenseList []string
		for license := range licenses {
			licenseList = append(licenseList, license)
		}
		return "This project is redistributable because it contains the following licenses: " + strings.Join(licenseList, ", ")
	}
	if len(l) == 0 {
		return "This project is not redistributable because it contains no licenses."
	}
	incompatibleLicenses := map[string]struct{}{}
	compatibleLicenses := map[string]struct{}{}
	for _, license := range l {
		if !license.IsCompatible {
			incompatibleLicenses[license.SPDX] = struct{}{}
		} else {
			compatibleLicenses[license.SPDX] = struct{}{}
		}
	}
	var incompatibleLicenseList []string
	for license := range incompatibleLicenses {
		incompatibleLicenseList = append(incompatibleLicenseList, license)
	}
	var compatibleLicenseList []string
	for license := range compatibleLicenses {
		compatibleLicenseList = append(compatibleLicenseList, license)
	}
	if len(compatibleLicenses) > 0 {
		return "This project is not redistributable because it contains the following incompatible licenses: " + strings.Join(incompatibleLicenseList, ",") + ". It also contains the following compatible licenses: " + strings.Join(compatibleLicenseList, ", ")
	}
	return "This project is not redistributable because it contains the following incompatible licenses: " + strings.Join(incompatibleLicenseList, ",")

}

func (l List) String() string {
	str := make([]string, len(l))
	for i, license := range l {
		str[i] = license.SPDX
	}
	return strings.Join(str, ", ")
}

// License describes a license found in a repository.
//
// swagger:model License
type License struct {
	// SPDX is the SPDX identifier for the license.
	// required: true
	SPDX string `json:"spdx"`
	// Confidence indicates how accurate the license detection is.
	// required: true
	Confidence float32 `json:"confidence"`
	// IsCompatible signals if the license is compatible with the OpenTofu project.
	// required: true
	IsCompatible bool `json:"is_compatible"`
	// File holds the file in the repository where the license was detected.
	// required: true
	File string `json:"file"`
	// Link may contain a link to the license file for humans to view. This may be empty.
	// required: false
	Link string `json:"link,omitempty"`
}

type Opt func(config *Config) error

func WithConfidenceThreshold(threshold float32) Opt {
	return func(config *Config) error {
		if threshold < 0 || threshold > 1 {
			return fmt.Errorf("invalid threshold: %f", threshold)
		}
		config.ConfidenceThreshold = threshold
		return nil
	}
}

func WithCompatibleLicenses(licenses ...string) Opt {
	return func(config *Config) error {
		config.CompatibleLicenses = append(config.CompatibleLicenses, licenses...)
		return nil
	}
}

type Config struct {
	CompatibleLicenses  []string
	ConfidenceThreshold float32
}

func (c *Config) ApplyDefaults() error {
	if c.ConfidenceThreshold == 0.0 {
		c.ConfidenceThreshold = 0.9
	}
	return nil
}

func (c *Config) Validate() error {
	if len(c.CompatibleLicenses) == 0 {
		return fmt.Errorf("no licenses configured")
	}
	return nil
}

type DetectOpt func(config *DetectConfig) error

type LinkFetcher func(license *License) error

type DetectConfig struct {
	LinkFetcher LinkFetcher
}

func (d *DetectConfig) applyDefaults() {
	if d.LinkFetcher == nil {
		d.LinkFetcher = func(license *License) error {
			return nil
		}
	}
}

func WithLinkFetcher(linkFetcher LinkFetcher) DetectOpt {
	return func(config *DetectConfig) error {
		config.LinkFetcher = linkFetcher
		return nil
	}
}

func New(
	opts ...Opt,
) (Detector, error) {
	config := Config{}
	for _, opt := range opts {
		if err := opt(&config); err != nil {
			return nil, err
		}
	}
	if err := config.ApplyDefaults(); err != nil {
		return nil, err
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	licenseMap := map[string]struct{}{}
	for _, license := range config.CompatibleLicenses {
		licenseMap[strings.ToLower(license)] = struct{}{}
	}
	return &detector{
		licenseMap: licenseMap,
		config:     config,
	}, nil
}

type detector struct {
	config     Config
	licenseMap map[string]struct{}
}

func (d detector) Detect(_ context.Context, repository fs.ReadDirFS, detectOptions ...DetectOpt) (List, error) {
	detectCfg := DetectConfig{}
	for _, opt := range detectOptions {
		if err := opt(&detectCfg); err != nil {
			return nil, err
		}
	}
	detectCfg.applyDefaults()

	licenses, err := licensedb.Detect(filer.FromFS(repository))
	if err != nil {
		if errors.Is(err, licensedb.ErrNoLicenseFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error detecting licenses: %w", err)
	}
	var result []License
	for license, match := range licenses {
		if strings.HasPrefix(license, "deprecated_") {
			continue
		}
		if match.Confidence >= d.config.ConfidenceThreshold {
			_, isCompatible := d.licenseMap[strings.ToLower(license)]
			l := License{
				SPDX:         license,
				Confidence:   match.Confidence,
				IsCompatible: isCompatible,
				File:         match.File,
			}
			if err := detectCfg.LinkFetcher(&l); err != nil {
				return nil, err
			}
			result = append(result, l)
		}
	}
	return result, nil
}
