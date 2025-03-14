//go:build !windows

package main

import (
	"context"
	"log/slog"
	"syscall"
)

func setRLimit(ctx context.Context, log *slog.Logger) error {
	log.InfoContext(ctx, "Setting maximum number of file descriptors to 50000...")
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{
		Cur: 50000,
		Max: 50000,
	}); err != nil {
		log.WarnContext(ctx, "Failed to set rlimit, generation may fail on larger repositories (%v)", err)
	}
	return nil
}
