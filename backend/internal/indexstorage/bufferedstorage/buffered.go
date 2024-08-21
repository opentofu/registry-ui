package bufferedstorage

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/opentofu/libregistry/logger"
	"golang.org/x/sync/errgroup"

	"github.com/opentofu/registry-ui/internal/indexstorage"
)

type BufferedStorage interface {
	indexstorage.TransactionalAPI

	UncommittedFiles() int
}

func New(log logger.Logger, localDir string, backingStorage indexstorage.API, parallelism int) (BufferedStorage, error) {
	log = log.WithName("Transactional Storage")

	index := newLocalIndex(localDir, log)

	stat, err := os.Stat(localDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(localDir, 0755); err != nil {
				return nil, fmt.Errorf("local directory %s does not exist and creation failed (%w)", localDir, err)
			}
		}
	} else {
		if !stat.IsDir() {
			return nil, fmt.Errorf("local directory %s is not a directory ", localDir)
		}
	}

	if err := index.Load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	if parallelism < 1 {
		return nil, fmt.Errorf("parallelism cannot be lower than 1")
	}

	return &buffered{
		logger:         log,
		localDir:       localDir,
		index:          index,
		backingStorage: backingStorage,
		lock:           &sync.Mutex{},
		parallelism:    parallelism,
	}, nil
}

type buffered struct {
	logger         logger.Logger
	backingStorage indexstorage.API
	index          *localIndex
	localDir       string
	lock           *sync.Mutex
	parallelism    int
}

func (b *buffered) UncommittedFiles() int {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.index.Root.uncommittedFiles()
}

func (b *buffered) Recover(ctx context.Context) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.index.CommitStarted {
		return b.commit(ctx)
	}
	return b.rollback(ctx)
}

func (b *buffered) ReadFile(ctx context.Context, filePath indexstorage.Path) ([]byte, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if err := filePath.Validate(); err != nil {
		return nil, err
	}
	status := b.index.FileStatus(ctx, filePath)
	localPath := path.Join(b.localDir, string(filePath))
	switch status {
	case fileStatusUnknown:
		data, err := b.backingStorage.ReadFile(ctx, filePath)
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(path.Dir(localPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for local cache file %s (%w)", localPath, err)
		}
		if err := os.WriteFile(localPath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write local cache file (%w)", err)
		}
		if err := b.index.RegisterCachedFile(ctx, filePath); err != nil {
			return nil, err
		}
		return data, nil
	case fileStatusPresent:
		return os.ReadFile(localPath)
	case fileStatusAbsent:
		return nil, fs.ErrNotExist
	default:
		return nil, fs.ErrNotExist
	}
}

func (b *buffered) WriteFile(ctx context.Context, filePath indexstorage.Path, contents []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if err := filePath.Validate(); err != nil {
		return err
	}

	localPath := path.Join(b.localDir, string(filePath))
	if err := os.MkdirAll(path.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s (%w)", localPath, err)
	}
	if err := os.WriteFile(localPath, contents, 0644); err != nil {
		return err
	}
	return b.index.MarkFileDirty(ctx, filePath)
}

func (b *buffered) RemoveAll(ctx context.Context, filePath indexstorage.Path) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if err := filePath.Validate(); err != nil {
		return err
	}
	localPath := path.Join(b.localDir, string(filePath))
	if err := os.RemoveAll(localPath); err != nil {
		return err
	}
	return b.index.MarkDirWiped(ctx, filePath)
}

func (b *buffered) Subdirectory(_ context.Context, dir indexstorage.Path) (indexstorage.API, error) {
	return &subdir{
		dir,
		b,
	}, nil
}

func (b *buffered) Rollback(ctx context.Context) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.rollback(ctx)
}

func (b *buffered) rollback(ctx context.Context) error {
	b.logger.Info(ctx, "Rolling back changes...")

	return b.reset(ctx)
}

func (b *buffered) Commit(ctx context.Context) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	err := b.commit(ctx)
	if err != nil {
		return err
	}

	err = b.index.Close()
	if err != nil {
		return err
	}
	return nil
}

func (b *buffered) commit(ctx context.Context) error {
	b.logger.Info(ctx, "Committing changes to persistent storage...")

	if err := b.index.MarkCommitStarted(ctx); err != nil {
		return err
	}

	eg := &errgroup.Group{}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg.SetLimit(b.parallelism)
	if err := b.commitWalk(ctx, cancel, b.index.Root, "", eg); err != nil {
		b.logger.Info(ctx, "Commit failed (%v).", err)
		return err
	}
	if err := eg.Wait(); err != nil {
		b.logger.Info(ctx, "Commit failed (%v).", err)
		return err
	}
	if err := b.reset(ctx); err != nil {
		b.logger.Info(ctx, "Commit failed (%v).", err)
		return err
	}
	b.logger.Info(ctx, "Commit complete.")
	return nil
}

func (b *buffered) commitWalk(ctx context.Context, cancel context.CancelFunc, dir *directory, dirPath indexstorage.Path, eg *errgroup.Group) error {
	// TODO this is super ugly and definitely violates the index boundary. Make this nicer.
	if dir.IsWiped {
		b.logger.Debug(ctx, "Removing /%s/* ...", dirPath)
		removeStart := time.Now()
		if err := b.backingStorage.RemoveAll(ctx, dirPath); err != nil {
			b.logger.Trace(ctx, "Failed to remove /%s/* (%v).", dirPath, err)
			return err
		}
		// Remove the directory remotely and remove the wipe mark so it behaves like a normal directory in a later run
		// in case the commit fails at a later stage.
		if err := b.index.UnmarkDirWiped(ctx, dir); err != nil {
			b.logger.Trace(ctx, "Failed to remove /%s/* (%v).", dirPath, err)
			return err
		}
		removeEnd := time.Now()
		b.logger.Debug(ctx, "Completed removing /%s/* in %f seconds.", dirPath, removeEnd.Sub(removeStart).Seconds())
	}

	b.logger.Info(ctx, "Committing /%s ...", dirPath)

	// Since the directory is now wiped if needed, it is safe to process all subdirectories and files in parallel.
	for name, subdir := range dir.Subdirectories {
		name := name
		subdir := subdir
		// TODO parallelize this to make directory removals non-sequential. However, they need to happen before the
		//      file uploads are queued.
		select {
		case <-ctx.Done():
			return fmt.Errorf("commit aborted (%v)", ctx.Err())
		default:
		}
		newPath := indexstorage.Path(path.Join(string(dirPath), name))
		b.logger.Trace(ctx, "Committing directory /%s...", newPath)
		if err := b.commitWalk(ctx, cancel, subdir, newPath, eg); err != nil {
			b.logger.Trace(ctx, "Committing directory /%s failed (%v).", newPath, err)
			return fmt.Errorf("failed to commit %s (%w)", name, err)
		}
		b.logger.Trace(ctx, "Committing directory /%s completed.", newPath)
	}
	for name, f := range dir.Files {
		if !f.IsDirty {
			continue
		}
		name := name
		f := f
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return fmt.Errorf("commit aborted (%v)", ctx.Err())
			default:
			}
			uploadStart := time.Now()
			uploadPath := path.Join(string(dirPath), name)
			fullPath := path.Join(b.localDir, uploadPath)
			b.logger.Debug(ctx, "Storing file /%s ...", uploadPath)
			contents, err := os.ReadFile(fullPath)
			if err != nil {
				b.logger.Trace(ctx, "Storing file /%s failed (%v)...", uploadPath, err)
				cancel()
				return fmt.Errorf("failed to read local file")
			}
			b.logger.Trace(ctx, "Read file /%s (%d bytes).", uploadPath, len(contents))
			if err := b.backingStorage.WriteFile(ctx, indexstorage.Path(uploadPath), contents); err != nil {
				b.logger.Trace(ctx, "Storing file /%s failed (%v)...", uploadPath, err)
				cancel()
				return fmt.Errorf("failed to sync %s (%w)", uploadPath, err)
			}
			b.logger.Trace(ctx, "Stored file /%s (%d bytes).", uploadPath, len(contents))
			// Remove the dirty mark so a later run doesn't try to upload it again if this run fails later down the
			// line.
			if err := b.index.UnmarkFileDirty(ctx, f); err != nil {
				b.logger.Trace(ctx, "Storing /%s failed (%v)...", uploadPath, err)
				cancel()
				return fmt.Errorf("failed to remove the dirty mark from %s (%w)", uploadPath, err)
			}
			uploadEnd := time.Now()
			b.logger.Debug(ctx, "Storing file /%s completed in %f seconds.", uploadPath, uploadEnd.Sub(uploadStart).Seconds())
			return nil
		})
	}
	return nil
}

func (b *buffered) reset(_ context.Context) error {
	if runtime.GOOS == "windows" {
		// Make sure no file locks are left open.
		runtime.GC()
	}
	if err := os.RemoveAll(b.localDir); err != nil {
		return fmt.Errorf("cannot clear local directory %s (%w)", b.localDir, err)
	}
	if err := os.MkdirAll(b.localDir, 0755); err != nil {
		return fmt.Errorf("cannot recreate local directory %s (%w)", b.localDir, err)
	}
	b.index = newLocalIndex(b.localDir, b.logger)
	return nil
}

type subdir struct {
	prefix  indexstorage.Path
	backing indexstorage.API
}

func (s *subdir) ReadFile(ctx context.Context, filepath indexstorage.Path) ([]byte, error) {
	return s.backing.ReadFile(ctx, indexstorage.Path(path.Join(string(s.prefix), string(filepath))))
}

func (s *subdir) WriteFile(ctx context.Context, filepath indexstorage.Path, contents []byte) error {
	return s.backing.WriteFile(ctx, indexstorage.Path(path.Join(string(s.prefix), string(filepath))), contents)
}

func (s *subdir) RemoveAll(ctx context.Context, dirPath indexstorage.Path) error {
	return s.backing.RemoveAll(ctx, indexstorage.Path(path.Join(string(s.prefix), string(dirPath))))
}

func (s *subdir) Subdirectory(_ context.Context, dir indexstorage.Path) (indexstorage.API, error) {
	return &subdir{
		prefix:  indexstorage.Path(path.Join(string(s.prefix), string(dir))),
		backing: s.backing,
	}, nil
}
