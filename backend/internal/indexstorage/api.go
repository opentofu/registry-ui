package indexstorage

import (
	"context"
	"fmt"
	"regexp"
)

type Path string

const MaxPathLength = 255

var pathPartRe = regexp.MustCompile("^[a-zA-Z0-9_.-]+(/[a-zA-Z0-9_.-]+)*$")

func (p Path) Validate() error {
	if len(p) > MaxPathLength {
		return fmt.Errorf("path too long: %s", p)
	}
	if !pathPartRe.MatchString(string(p)) {
		return fmt.Errorf("invalid path: %s", p)
	}
	return nil
}

type API interface {
	// GetFileSHA256 returns the SHA256 hash of a file hex-encoded
	GetFileSHA256(ctx context.Context, path Path) (string, error)

	// ReadFile reads the given path if found, otherwise returns a not found error.
	ReadFile(ctx context.Context, path Path) ([]byte, error)

	// WriteFile writes a given file at a path, creating all directories in the path.
	WriteFile(ctx context.Context, path Path, contents []byte) error

	// RemoveAll removes all files under the given path.
	RemoveAll(ctx context.Context, path Path) error

	// Subdirectory returns a storage API that is restricted to a subdirectory.
	Subdirectory(ctx context.Context, dir Path) (API, error)
}
