package tofu

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/opentofu/registry-ui/pkg/telemetry"
)

const (
	nightliesBaseURL = "https://nightlies.opentofu.org"
	latestURL        = nightliesBaseURL + "/nightlies/latest.json"
)

// extractTofuBinary extracts the tofu binary from a zip archive into destination.
func extractTofuBinary(zipData []byte, destination string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	for _, file := range zipReader.File {
		if file.Name != "tofu" && !strings.HasSuffix(file.Name, "/tofu") {
			continue
		}
		if file.FileInfo().IsDir() {
			continue
		}

		fileInArchive, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in archive: %w", err)
		}
		dstFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			fileInArchive.Close()
			return fmt.Errorf("failed to open destination file for writing: %w", err)
		}
		_, err = io.Copy(dstFile, fileInArchive)
		dstFile.Close()
		fileInArchive.Close()
		if err != nil {
			return fmt.Errorf("failed to copy file content: %w", err)
		}
		return nil
	}
	return fmt.Errorf("tofu binary not found in archive")
}

func httpGet(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request for %s: %w", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, url)
	}

	return resp.Body, nil
}

// resolveNightlyArtifactURL fetches the latest nightly metadata and returns the
// download URL for the zip artifact matching the current platform.
func resolveNightlyArtifactURL(ctx context.Context) (string, error) {
	reader, err := httpGet(ctx, latestURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest nightly metadata: %w", err)
	}
	defer reader.Close()

	var response struct {
		Version   string   `json:"version"`
		Date      string   `json:"date"`
		Commit    string   `json:"commit"`
		Path      string   `json:"path"`
		Artifacts []string `json:"artifacts"`
	}
	if err := json.NewDecoder(reader).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode nightly metadata: %w", err)
	}

	platformString := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	var artifactPath string
	for _, artifact := range response.Artifacts {
		if strings.Contains(artifact, platformString) && strings.HasSuffix(artifact, ".zip") {
			artifactPath = artifact
			break
		}
	}
	if artifactPath == "" {
		return "", fmt.Errorf("no %s artifact found", platformString)
	}

	slog.InfoContext(ctx, "Found latest nightly build of OpenTofu", "version", response.Version, "date", response.Date, "commit", response.Commit, "artifact", artifactPath, "platform", platformString)
	return fmt.Sprintf("%s%s%s", nightliesBaseURL, response.Path, artifactPath), nil
}

// DownloadLatestNightly downloads the latest nightly build of OpenTofu.
// We use nightlies rather than stable releases because:
//   - It ensures we always have access to the latest `tofu show -json` module configuration output,
//     which is required for module indexing.
//   - It lets us ship fixes to OpenTofu quickly without maintaining a separate branch.
func DownloadLatestNightly(ctx context.Context, destination string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "tofu.download_latest_nightly")
	defer span.End()

	slog.DebugContext(ctx, "Downloading latest nightly build of Tofu", "destination", destination)

	artifactURL, err := resolveNightlyArtifactURL(ctx)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "Downloading artifact", "url", artifactURL)
	reader, err := httpGet(ctx, artifactURL)
	if err != nil {
		return fmt.Errorf("failed to download artifact: %w", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read artifact body: %w", err)
	}

	if err := extractTofuBinary(body, destination); err != nil {
		return err
	}

	slog.InfoContext(ctx, "Successfully downloaded latest nightly build of Tofu", "destination", destination)
	return nil
}
