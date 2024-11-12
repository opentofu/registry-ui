//go:build !windows

package main

import (
	"context"
	"syscall"

	"github.com/opentofu/libregistry/logger"
)

func setRLimit(ctx context.Context, log logger.Logger) error {
	log.Info(ctx, "Setting maximum number of file descriptors to 50000...")
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{
		Cur: 50000,
		Max: 50000,
	}); err != nil {
		log.Warn(ctx, "Failed to set rlimit, generation may fail on larger repositories (%v)", err)
	}
	return nil
}
