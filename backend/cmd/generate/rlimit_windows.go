//go:build windows

package main

import (
	"context"
	"github.com/opentofu/libregistry/logger"
)

func setRLimit(_ context.Context, _ logger.Logger) error {
	return nil
}
