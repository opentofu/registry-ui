package filesystemstorage

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/opentofu/registry-ui/internal/indexstorage"
)

func New(directory string) (indexstorage.API, error) {
	stat, err := os.Stat(directory)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(directory, 0755); err != nil {
				return nil, fmt.Errorf("storage directory %s does not exist and cannot be created (%w)", directory, err)
			}
		} else {
			return nil, fmt.Errorf("storage directory %s is inaccessible (%w)", directory, err)
		}
	} else {
		if !stat.IsDir() {
			return nil, fmt.Errorf("storage location %s exists, but is not a directory", directory)
		}
	}
	return &storageAPI{
		directory: directory,
	}, nil
}

type storageAPI struct {
	directory string
}

func (s storageAPI) Subdirectory(_ context.Context, storagePath indexstorage.Path) (indexstorage.API, error) {
	if err := storagePath.Validate(); err != nil {
		return nil, err
	}
	absPath := path.Join(s.directory, string(storagePath))
	return &storageAPI{
		absPath,
	}, nil
}

func (s storageAPI) RemoveAll(_ context.Context, storagePath indexstorage.Path) error {
	if err := storagePath.Validate(); err != nil {
		return err
	}
	absPath := path.Join(s.directory, string(storagePath))
	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("failed to remove %s (%w)", absPath, err)
	}
	return nil
}

func (s storageAPI) ReadFile(_ context.Context, targetPath indexstorage.Path) ([]byte, error) {
	if err := targetPath.Validate(); err != nil {
		return nil, err
	}
	absPath := path.Join(s.directory, string(targetPath))
	return os.ReadFile(absPath)
}

func (s storageAPI) WriteFile(_ context.Context, targetPath indexstorage.Path, contents []byte) error {
	if err := targetPath.Validate(); err != nil {
		return err
	}
	absPath := path.Join(s.directory, string(targetPath))

	dir := path.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory %s for file %s (%w)", dir, absPath, err)
	}
	return os.WriteFile(absPath, contents, 0644)
}
