package tofu

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// Show executes tofu show -json -module=DIR and returns the parsed Config.
// Returns (config, stderr, error) — stderr is returned for callers to store in SchemaError if needed.
func Show(ctx context.Context, moduleDir string) (*Config, string, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "tofu.show")
	defer span.End()

	span.SetAttributes(
		attribute.String("module.dir", moduleDir),
	)

	cwd, err := os.Getwd()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Execute `tofu show -json -module=moduleDir`
	cmd := exec.CommandContext(ctx, path.Join(cwd, BinaryName), "show", "-json", "-module="+moduleDir)
	cmd.Dir = moduleDir

	slog.DebugContext(ctx, "Executing tofu show", "cmd", cmd.String(), "dir", moduleDir)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// Return stderr to caller for storage in SchemaError (don't log it to avoid large traces)
	stderrStr := stderr.String()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, stderrStr, fmt.Errorf("tofu show failed: %w", err)
	}

	output := stdout.String()
	span.SetAttributes(
		attribute.Int("tofu.output_size", len(output)),
	)

	// Parse the JSON output
	var config *Config
	if err := json.Unmarshal([]byte(output), &config); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, stderrStr, fmt.Errorf("failed to parse tofu JSON output: %w", err)
	}

	slog.DebugContext(ctx, "Successfully executed tofu show",
		"dir", moduleDir,
		"outputSize", len(output))

	return config, stderrStr, nil
}
