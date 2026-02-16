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

// TODO: Investigate if we should bring in tofudl here?

// copyFile copies a file from src to dst, preserving file permissions.
// This is used instead of os.Rename to support cross-filesystem moves.
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info to preserve permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Create destination file with same permissions
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

func DownloadLatestNightly(ctx context.Context, destination string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "DownloadLatestNightly")
	defer span.End()

	slog.DebugContext(ctx, "Downloading latest nightly build of Tofu", "destination", destination, "latestURL", latestURL)

	req, err := http.NewRequestWithContext(ctx, "GET", latestURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// deserialize JSON response
	var response struct {
		Version   string   `json:"version"`
		Date      string   `json:"date"`
		Commit    string   `json:"commit"`
		Path      string   `json:"path"`
		Artifacts []string `json:"artifacts"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Determine the platform string based on runtime
	platformString := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	// Find the artifact for the detected platform
	var artifactPath string
	for _, artifact := range response.Artifacts {
		if strings.Contains(artifact, platformString) && strings.HasSuffix(artifact, ".zip") {
			artifactPath = artifact
			break
		}
	}
	if artifactPath == "" {
		return fmt.Errorf("no %s artifact found", platformString)
	}

	slog.InfoContext(ctx, "Found latest nightly build of OpenTofu", "version", response.Version, "date", response.Date, "commit", response.Commit, "artifact", artifactPath, "platform", platformString)
	artifactURL := fmt.Sprintf("%s%s%s", nightliesBaseURL, response.Path, artifactPath)
	slog.DebugContext(ctx, "Downloading artifact", "url", artifactURL)
	req, err = http.NewRequestWithContext(ctx, "GET", artifactURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create artifact request: %w", err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform artifact request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected artifact status code: %d", resp.StatusCode)
	}

	slog.DebugContext(ctx, "Downloading and extracting artifact", "url", artifactURL)
	tmpDir, err := os.MkdirTemp("", "tofu-nightly-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir")
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read artifact body: %w", err)
	}

	slog.DebugContext(ctx, "Extracting zip", "size", len(body), "tempDir", tmpDir)
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	for _, file := range zipReader.File {
		if file.Name != "tofu" && !strings.HasSuffix(file.Name, "/tofu") {
			continue
		}

		fpath := fmt.Sprintf("%s/%s", tmpDir, file.Name)
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}
		dstFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to open file for writing: %w", err)
		}
		fileInArchive, err := file.Open()
		if err != nil {
			dstFile.Close()
			return fmt.Errorf("failed to open file in archive: %w", err)
		}
		_, err = io.Copy(dstFile, fileInArchive)
		dstFile.Close()
		fileInArchive.Close()
		if err != nil {
			return fmt.Errorf("failed to copy file content: %w", err)
		}

		slog.DebugContext(ctx, "Extracted file from zip", "file", fpath)
	}

	srcPath := fmt.Sprintf("%s/tofu", tmpDir)
	err = copyFile(srcPath, destination)
	if err != nil {
		return fmt.Errorf("failed to copy tofu binary to destination: %w", err)
	}

	slog.InfoContext(ctx, "Successfully downloaded latest nightly build of Tofu", "destination", destination)

	return nil
}
