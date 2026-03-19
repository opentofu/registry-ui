package license

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-enry/go-license-detector/v4/licensedb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/repository"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

type Detector struct {
	config       config.LicenseConfig
	licenseMap   map[string]struct{}
	githubClient *repository.Client
}

func New(licenseConfig config.LicenseConfig, githubClient *repository.Client) (*Detector, error) {
	licenseMap := map[string]struct{}{}
	for _, license := range licenseConfig.CompatibleLicenses {
		licenseMap[strings.ToLower(license)] = struct{}{}
	}
	return &Detector{
		licenseMap:   licenseMap,
		config:       licenseConfig,
		githubClient: githubClient,
	}, nil
}

func (d *Detector) Detect(ctx context.Context, directory string, repoURL string) (List, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "license.detect")
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

	matches, err := d.detectLicenseInDirectory(ctx, directory)
	if err != nil {
		return nil, err
	}

	var result []License
	if len(matches) > 0 {
		filesWithLicenses := d.buildLicenseFileMap(matches, repoURL)
		licenseFiles := d.filterAndSortLicenseFiles(ctx, filesWithLicenses)
		result = d.collectResults(ctx, span, licenseFiles, filesWithLicenses)
	}

	if len(result) > 0 {
		span.SetAttributes(
			attribute.Bool("license.used_github_fallback", false),
			attribute.Int("license.final_result_count", len(result)),
		)
		return result, nil
	}

	// No licenses found locally, try GitHub API fallback
	slog.InfoContext(ctx, "No licenses detected locally, trying GitHub API fallback", "repo_url", repoURL)
	githubLicense, err := d.detectLicenseFromGitHub(ctx, repoURL)
	if err != nil {
		slog.WarnContext(ctx, "GitHub API license detection failed", "error", err)
		return nil, nil
	}
	if githubLicense != nil {
		slog.InfoContext(ctx, "GitHub API fallback succeeded", "license", githubLicense.SPDX)
		span.SetAttributes(
			attribute.Bool("license.used_github_fallback", true),
			attribute.Int("license.github_result_count", 1),
		)
		return []License{*githubLicense}, nil
	}

	return nil, nil
}

func (d *Detector) detectLicenseInDirectory(ctx context.Context, directory string) ([]licensedb.Match, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "license.detect_in_directory")
	defer span.End()

	// Use licensedb.Analyse to analyze the directory
	results := licensedb.Analyse(directory)
	if len(results) == 0 {
		slog.DebugContext(ctx, "No results from license analysis")
		span.SetAttributes(attribute.Int("license.detected_count", 0))
		return nil, nil
	}

	// Check if there were any errors (but treat "no license found" as not an error)
	for _, result := range results {
		if result.ErrStr != "" {
			// "no license file was found" is not an error - it's expected for repos without licenses
			// We'll return nil,nil and let the GitHub API fallback handle it
			if result.ErrStr == "no license file was found" {
				slog.DebugContext(ctx, "No license file found in directory")
				span.SetAttributes(
					attribute.Bool("license.no_license_found", true),
					attribute.String("license.detection_status", "no_license"),
				)
				return nil, nil
			}
			// Other errors are actual problems
			err := fmt.Errorf("license analysis error: %s", result.ErrStr)
			span.RecordError(err)
			slog.ErrorContext(ctx, "Error analyzing licenses", "error", err)
			return nil, err
		}
	}

	// Collect all matches from all results
	var allMatches []licensedb.Match
	for _, result := range results {
		allMatches = append(allMatches, result.Matches...)
	}

	span.SetAttributes(attribute.Int("license.detected_count", len(allMatches)))

	slog.DebugContext(ctx, "License detection completed", "detected_licenses_count", len(allMatches))
	return allMatches, nil
}

func (d *Detector) buildLicenseFileMap(matches []licensedb.Match, repoURL string) map[string][]License {
	filesWithLicenses := make(map[string][]License)

	for _, match := range matches {
		// Skip deprecated licenses (this is license names, not filenames)
		if strings.HasPrefix(match.License, "deprecated_") {
			continue
		}

		_, isCompatible := d.licenseMap[strings.ToLower(match.License)]

		if match.Confidence >= d.config.ConfidenceThreshold {
			filesWithLicenses[match.File] = append(filesWithLicenses[match.File], License{
				SPDX:         match.License,
				Confidence:   match.Confidence,
				IsCompatible: isCompatible,
				File:         match.File,
				Link:         generateGitHubLink(repoURL, match.File),
			})
		}
	}

	return filesWithLicenses
}

func (d *Detector) filterAndSortLicenseFiles(ctx context.Context, filesWithLicenses map[string][]License) []string {
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

func (d *Detector) collectResults(ctx context.Context, span trace.Span, licenseFiles []string, filesWithLicenses map[string][]License) []License {
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

	// Ignore README and other documentation files that might contain license references
	upperFileName := strings.ToUpper(fileName)
	docFiles := []string{"README", "CHANGELOG", "CHANGES", "HISTORY", "NEWS"}
	for _, docFile := range docFiles {
		if strings.HasPrefix(upperFileName, docFile) {
			return true, "documentation file"
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
	return strings.Count(filePath, "/")
}
