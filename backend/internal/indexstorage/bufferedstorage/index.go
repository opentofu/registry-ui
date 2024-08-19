package bufferedstorage

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/opentofu/libregistry/logger"

	"github.com/opentofu/registry-ui/internal/indexstorage"
)

func newLocalIndex(localDir string, logger logger.Logger) *localIndex {
	return &localIndex{
		CommitStarted: false,
		Root: &directory{
			IsWiped:        false,
			Subdirectories: map[string]*directory{},
			Files:          map[string]*file{},
		},
		backingFile:   path.Join(localDir, ".index.json"),
		lock:          &sync.Mutex{},
		lastCommitted: time.Time{},

		logger: logger,
	}
}

// localIndex keeps track of files and directories that exist locally. It also keeps track of
// directories that have been wiped locally so the commit can wipe them remotely before
// uploading any dirty files.
type localIndex struct {
	Root *directory `json:"root"`

	backingFile   string
	lock          *sync.Mutex
	CommitStarted bool

	lastCommitted time.Time

	logger logger.Logger
}

type fileStatus int

const (
	// fileStatusUnknown indicates that the file status is locally unknown and a fetch should be attempted from
	// the backing storage.
	fileStatusUnknown fileStatus = 0
	// fileStatusPresent indicates that the file is present and exists locally.
	fileStatusPresent fileStatus = 1
	// fileStatusAbsent indicates that the file is not present and was wiped, a fetch should not be attempted.
	fileStatusAbsent fileStatus = 2
)

func (l *localIndex) MarkCommitStarted(ctx context.Context) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.CommitStarted = true
	return l.save(ctx)
}

func (l *localIndex) FileStatus(ctx context.Context, filepath indexstorage.Path) fileStatus {
	l.lock.Lock()
	defer l.lock.Unlock()
	_, status := l.resolveFile(ctx, filepath)
	return status
}

func (l *localIndex) UnmarkDirWiped(ctx context.Context, dir *directory) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	dir.IsWiped = false
	return l.trySave(ctx)
}

func (l *localIndex) MarkDirWiped(ctx context.Context, filepath indexstorage.Path) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	parts := strings.Split(string(filepath), "/")
	currentDir := l.Root
	for _, part := range parts[:len(parts)-1] {
		newCurrentDir, ok := currentDir.Subdirectories[part]
		if !ok {
			newCurrentDir = &directory{
				IsWiped:        false,
				Subdirectories: map[string]*directory{},
				Files:          map[string]*file{},
			}
			currentDir.Subdirectories[part] = newCurrentDir
		}
		currentDir = newCurrentDir
	}

	lastPart := parts[len(parts)-1]
	currentDir.Subdirectories[lastPart] = &directory{
		IsWiped:        true,
		Subdirectories: map[string]*directory{},
		Files:          map[string]*file{},
	}
	return l.trySave(ctx)
}

func (l *localIndex) RegisterCachedFile(_ context.Context, filepath indexstorage.Path) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.createFile(filepath, &file{
		IsDirty: false,
	})
}

func (l *localIndex) UnmarkFileDirty(ctx context.Context, f *file) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	f.IsDirty = false
	return l.trySave(ctx)
}

func (l *localIndex) MarkFileDirty(ctx context.Context, filepath indexstorage.Path) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	f, _ := l.resolveFile(ctx, filepath)
	if f == nil {
		return l.createFile(filepath, &file{
			IsDirty: true,
		})
	}
	f.IsDirty = true
	return l.trySave(ctx)
}

func (l *localIndex) createFile(filepath indexstorage.Path, data *file) error {
	parts := strings.Split(string(filepath), "/")
	currentDir := l.Root
	for _, part := range parts[:len(parts)-1] {
		newCurrentDir, ok := currentDir.Subdirectories[part]
		if !ok {
			newCurrentDir = &directory{
				IsWiped:        false,
				Subdirectories: map[string]*directory{},
				Files:          map[string]*file{},
			}
			currentDir.Subdirectories[part] = newCurrentDir
		}
		currentDir = newCurrentDir
	}
	currentDir.Files[parts[len(parts)-1]] = data
	return nil
}

func (l *localIndex) resolveFile(_ context.Context, filepath indexstorage.Path) (*file, fileStatus) {
	parts := strings.Split(string(filepath), "/")
	currentDir := l.Root
	wiped := false
	for _, part := range parts[:len(parts)-1] {
		if currentDir.IsWiped {
			wiped = true
		}
		var ok bool
		currentDir, ok = currentDir.Subdirectories[part]
		if !ok {
			if wiped {
				return nil, fileStatusAbsent
			}
			return nil, fileStatusUnknown
		}
	}
	if currentDir.IsWiped {
		wiped = true
	}
	f, ok := currentDir.Files[parts[len(parts)-1]]
	if !ok {
		if wiped {
			return nil, fileStatusAbsent
		}
		return nil, fileStatusUnknown
	}
	return f, fileStatusPresent
}

func (l *localIndex) Load() error {
	l.lock.Lock()
	defer l.lock.Unlock()

	marshalled, err := os.ReadFile(l.backingFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshalled, &l)
}

func (l *localIndex) Close() error {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.save(context.Background())
}

func (l *localIndex) trySave(ctx context.Context) error {
	if time.Now().Sub(l.lastCommitted) < 30*time.Second {
		return nil
	}

	l.lastCommitted = time.Now()

	return l.save(ctx)
}

func (l *localIndex) save(ctx context.Context) error {
	l.logger.Debug(ctx, "Saving local index")
	marshalled, err := json.Marshal(l)
	if err != nil {
		return err
	}
	if err := os.WriteFile(l.backingFile, marshalled, 0644); err != nil {
		return err
	}
	return nil
}

type directory struct {
	// IsWiped indicates that the directory is wiped and should be fully resynced from the local directory.
	IsWiped bool `json:"is_wiped"`
	// Subdirectories holds a set of Subdirectories of the current directory.
	Subdirectories map[string]*directory `json:"subdirectories"`
	// Files contains a map of files that are present locally. This list may not be exhaustive unless IsWiped is true
	// and files may be added on reading.
	Files map[string]*file `json:"files"`
}

func (d *directory) UnmarshalJSON(data []byte) error {
	tmp := struct {
		IsWiped        bool                  `json:"is_wiped"`
		Subdirectories map[string]*directory `json:"subdirectories"`
		Files          map[string]*file      `json:"files"`
	}{
		false,
		map[string]*directory{},
		map[string]*file{},
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	d.IsWiped = tmp.IsWiped
	d.Subdirectories = tmp.Subdirectories
	d.Files = tmp.Files
	return nil
}

type file struct {
	// IsDirty indicates that the file should be re-uploaded.
	IsDirty bool `json:"is_dirty"`
}
