package license

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/api"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/opentofu/registry-ui/pkg/config"
)

type Detector interface {
	Detect(ctx context.Context, repository fs.ReadDirFS, repoURL string) (List, error)
}

type detector struct {
	config     config.LicenseConfig
	licenseMap map[string]struct{}
}

func New(licenseConfig config.LicenseConfig) (Detector, error) {
	licenseMap := map[string]struct{}{}
	for _, license := range licenseConfig.CompatibleLicenses {
		licenseMap[strings.ToLower(license)] = struct{}{}
	}
	return &detector{
		licenseMap: licenseMap,
		config:     licenseConfig,
	}, nil
}

func (d detector) Detect(ctx context.Context, repository fs.ReadDirFS, repoURL string) (List, error) {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "license.detect")
	defer span.End()

	span.SetAttributes(
		attribute.Float64("license.confidence_threshold", float64(d.config.ConfidenceThreshold)),
		attribute.Float64("license.confidence_override_threshold", float64(d.config.ConfidenceOverrideThreshold)),
		attribute.Int("license.compatible_licenses_count", len(d.config.CompatibleLicenses)),
		attribute.String("license.repo_url", repoURL),
	)

	slog.DebugContext(ctx, "Starting license detection",
		"confidence_threshold", d.config.ConfidenceThreshold,
		"confidence_override_threshold", d.config.ConfidenceOverrideThreshold,
		"compatible_licenses_count", len(d.config.CompatibleLicenses))

	licenses, err := d.detectLicenseInDirectory(ctx, span, repository)
	if err != nil {
		return nil, err
	}
	if licenses == nil {
		return nil, nil // No licenses found
	}

	filesWithLicenses := d.buildLicenseFileMap(ctx, licenses, repoURL)
	licenseFiles := d.filterAndSortLicenseFiles(ctx, filesWithLicenses)
	result := d.collectResults(ctx, span, licenseFiles, filesWithLicenses)

	span.SetAttributes(attribute.Int("license.final_result_count", len(result)))
	return result, nil
}

func (d detector) detectLicenseInDirectory(ctx context.Context, span trace.Span, repository fs.ReadDirFS) (map[string]api.Match, error) {
	licenses, err := licensedb.Detect(filer.FromFS(repository))
	if err != nil {
		if errors.Is(err, licensedb.ErrNoLicenseFound) {
			slog.DebugContext(ctx, "No licenses found in repository")
			span.SetAttributes(attribute.Int("license.detected_count", 0))
			return nil, nil
		}
		span.RecordError(err)
		slog.ErrorContext(ctx, "Error detecting licenses", "error", err)
		return nil, fmt.Errorf("error detecting licenses: %w", err)
	}

	span.SetAttributes(attribute.Int("license.detected_count", len(licenses)))
	slog.DebugContext(ctx, "License detection completed", "detected_licenses_count", len(licenses))
	return licenses, nil
}

func (d detector) buildLicenseFileMap(ctx context.Context, licenses map[string]api.Match, repoURL string) map[string][]License {
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
					Link:         generateGitHubLink(repoURL, file),
				})
			}
		}
	}

	return filesWithLicenses
}

func (d detector) filterAndSortLicenseFiles(ctx context.Context, filesWithLicenses map[string][]License) []string {
	var licenseFiles []string
	for file := range filesWithLicenses {
		if shouldIgnore, reason := shouldIgnoreLicenseFile(file); shouldIgnore {
			slog.DebugContext(ctx, "Ignoring license file", "file", file, "reason", reason)
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

	return licenseFiles
}

func (d detector) collectResults(ctx context.Context, span trace.Span, licenseFiles []string, filesWithLicenses map[string][]License) []License {
	var result []License

	// Iterate through sorted list of potential license files
	for _, file := range licenseFiles {
		for _, l := range filesWithLicenses[file] {
			// Exit early (keeping in mind the sort order above)
			if l.Confidence >= d.config.ConfidenceOverrideThreshold {
				span.SetAttributes(
					attribute.String("license.early_return_spdx", l.SPDX),
					attribute.Float64("license.early_return_confidence", float64(l.Confidence)),
					attribute.String("license.early_return_file", l.File),
				)
				slog.InfoContext(ctx, "High-confidence license found, returning early",
					"license", l.SPDX, "confidence", l.Confidence, "file", l.File)
				return []License{l}
			}
			result = append(result, l)
		}
	}

	return result
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
