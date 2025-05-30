package license

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
)

type Detector interface {
	Detect(ctx context.Context, repository fs.ReadDirFS, detectOptions ...DetectOpt) (List, error)
}

// License describes a license found in a repository. Note: the license detection is best effort. When displaying the
// license to the user, always show a link to the actual license and warn users that they have to inspect the license
// themselves.
type License struct {
	// SPDX is the SPDX identifier for the license.
	SPDX string `json:"spdx"`
	// Confidence indicates how accurate the license detection is.
	Confidence float32 `json:"confidence"`
	// IsCompatible signals if the license is compatible with the OpenTofu project.
	IsCompatible bool `json:"is_compatible"`
	// File holds the file in the repository where the license was detected.
	File string `json:"file"`
	// Link may contain a link to the license file for humans to view. This may be empty.
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

func WithConfidenceOverrideThreshold(threshold float32) Opt {
	return func(config *Config) error {
		if threshold < 0 || threshold > 1 {
			return fmt.Errorf("invalid threshold: %f", threshold)
		}
		config.ConfidenceOverrideThreshold = threshold
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
	// ConfidenceOverrideThreshold is the limit at which a detected license overrides all other detected licenses.
	// Defaults to 98%.
	ConfidenceOverrideThreshold float32
}

func (c *Config) ApplyDefaults() error {
	if c.ConfidenceThreshold == 0.0 {
		c.ConfidenceThreshold = 0.85
	}
	if c.ConfidenceOverrideThreshold == 0 {
		c.ConfidenceOverrideThreshold = 0.98
	}
	return nil
}

func (c *Config) Validate() error {
	if len(c.CompatibleLicenses) == 0 {
		return fmt.Errorf("no licenses configured")
	}
	if c.ConfidenceOverrideThreshold < c.ConfidenceThreshold {
		return fmt.Errorf("the confidence override threshold (%f) is lower than the confidence threshold (%f)", c.ConfidenceOverrideThreshold, c.ConfidenceThreshold)
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

	filesWithLicenses := make(map[string][]License)

	for license, match := range licenses {
		// Skip deprecated licenses (this is license names, not filenames)
		if strings.HasPrefix(license, "deprecated_") {
			continue
		}

		_, isCompatible := d.licenseMap[strings.ToLower(license)]

		for file, confidence := range match.Files {
			if confidence >= d.config.ConfidenceThreshold {
				filesWithLicenses[file] = append(filesWithLicenses[file], License{
					SPDX:         license,
					Confidence:   confidence,
					IsCompatible: isCompatible,
					File:         file,
				})
			}
		}
	}

	var licenseFiles []string
	for file := range filesWithLicenses {
		if shouldIgnore, reason := shouldIgnoreLicenseFile(file); shouldIgnore {
			log.Printf("Ignoring license file: %s (reason: %s)", file, reason)
			delete(filesWithLicenses, file)
			continue
		}
		licenseFiles = append(licenseFiles, file)

		// Sort licenses within each file by confidence
		slices.SortFunc(filesWithLicenses[file], func(a, b License) int {
			if a.Confidence > b.Confidence {
				return -1
			}
			if a.Confidence < b.Confidence {
				return 1
			}
			return strings.Compare(strings.ToLower(a.SPDX), strings.ToLower(b.SPDX))
		})
	}

	// Sort license files: docs first, then path depth, then alphabetical
	slices.SortFunc(licenseFiles, func(a, b string) int {
		aIsDoc := isDocumentationDirectory(a)
		bIsDoc := isDocumentationDirectory(b)

		if aIsDoc != bIsDoc {
			if aIsDoc {
				return -1
			}
			return 1
		}

		aDepth := pathDepth(a)
		bDepth := pathDepth(b)
		if aDepth != bDepth {
			return aDepth - bDepth
		}

		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	var result []License

	// Iterate through sorted list of potential license files
	for _, file := range licenseFiles {
		for _, l := range filesWithLicenses[file] {
			if err := detectCfg.LinkFetcher(&l); err != nil {
				return nil, err
			}

			// Exit early (keeping in mind the sort order above)
			if l.Confidence >= d.config.ConfidenceOverrideThreshold {
				return []License{
					l,
				}, nil
			}
			result = append(result, l)
		}
	}
	return result, nil
}

func shouldIgnoreLicenseFile(filePath string) (bool, string) {
	fileName := filepath.Base(filePath)
	dirPath := filepath.Dir(filePath)

	// Ignore specific filenames
	ignoredFiles := []string{
		"THIRD_PARTY_LICENSES.txt", "THIRD_PARTY_LICENSE", "3RD_PARTY_LICENSES",
		"PATENTS", "NOTICE",
	}
	for _, ignored := range ignoredFiles {
		if strings.EqualFold(fileName, ignored) {
			return true, "ignored filename"
		}
	}

	ignoredDirs := []string{"vendor", "node_modules"}
	for _, dir := range ignoredDirs {
		if strings.Contains(dirPath, dir+"/") || strings.HasPrefix(dirPath, dir) {
			return true, "dependency directory"
		}
	}

	if strings.Contains(filePath, "examples/") || strings.Contains(filePath, "test") {
		return true, "examples/test directory"
	}

	return false, ""
}

func isDocumentationDirectory(filePath string) bool {
	docDirs := []string{"docs/", "doc/", "website/docs/", "documentation/"}
	for _, dir := range docDirs {
		if strings.HasPrefix(filePath, dir) {
			return true
		}
	}
	return false
}

func pathDepth(filePath string) int {
	if filePath == "." || filePath == "" {
		return 0
	}
	return strings.Count(filePath, "/")
}
