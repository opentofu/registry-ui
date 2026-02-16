// Package dltofunightly implements a CLI command to download the latest nightly build of Tofu.
package dltofunightly

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/opentofu/registry-ui/pkg/telemetry"
	"github.com/opentofu/registry-ui/pkg/tofu"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "dl-tofu-nightly",
		Usage: "Download the latest nightly build of Tofu",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx)
		},
	}
}

func run(ctx context.Context) error {
	ctx, span := telemetry.Tracer().Start(ctx, "download-tofu-nightly")
	defer span.End()

	// remove the file if it already exists
	stat, err := os.Stat("./tofu")
	if err == nil && !stat.IsDir() {
		err = os.Remove("./tofu")
		if err != nil {
			return fmt.Errorf("failed to remove existing tofu file: %w", err)
		}
		slog.InfoContext(ctx, "Removed existing tofu file")
	}

	err = tofu.DownloadLatestNightly(ctx, "./tofu")
	if err != nil {
		return fmt.Errorf("failed to download latest nightly build of Tofu: %w", err)
	}
	return nil
}
