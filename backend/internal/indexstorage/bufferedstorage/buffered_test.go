package bufferedstorage_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/indexstorage/bufferedstorage"
	"github.com/opentofu/registry-ui/internal/indexstorage/filesystemstorage"
	"github.com/opentofu/tofutestutils"
)

func TestSimpleCommit(t *testing.T) {
	const testContent = "Hello world!"
	const testFile1 = "test.txt"

	backingDir := t.TempDir()
	backingStorage := tofutestutils.Must2(filesystemstorage.New(backingDir))

	buffer1 := tofutestutils.Must2(bufferedstorage.New(logger.NewTestLogger(t), t.TempDir(), backingStorage, 25))

	ctx := tofutestutils.Context(t)

	tofutestutils.Must(buffer1.WriteFile(ctx, testFile1, []byte(testContent)))

	assertFileDoesNotExist(t, ctx, backingStorage, testFile1)
	assertFileExists(t, ctx, buffer1, testFile1)

	tofutestutils.Must(buffer1.Commit(ctx))

	assertFileExists(t, ctx, buffer1, testFile1)
	contents := assertFileExists(t, ctx, backingStorage, testFile1)

	if string(contents) != testContent {
		t.Fatalf("Incorrect contents returned: %s", contents)
	}
}

func TestDirectoryWipe(t *testing.T) {
	const testContent = "Hello world!"
	const testDir = "test"
	var testFile1 = indexstorage.Path(path.Join(testDir, "test.txt"))
	var testFile2 = indexstorage.Path(path.Join(testDir, "test2.txt"))

	backingDir := t.TempDir()
	backingStorage := tofutestutils.Must2(filesystemstorage.New(backingDir))
	buffer := tofutestutils.Must2(bufferedstorage.New(logger.NewTestLogger(t), t.TempDir(), backingStorage, 25))

	ctx := tofutestutils.Context(t)

	// Create a file in the test directory
	tofutestutils.Must(backingStorage.WriteFile(ctx, testFile1, []byte(testContent)))

	// Remove the directory of the test file
	tofutestutils.Must(buffer.RemoveAll(ctx, testDir))

	// Check if it still exists on the backing storage, but not on the local one.
	assertFileExists(t, ctx, backingStorage, testFile1)
	assertFileDoesNotExist(t, ctx, backingStorage, testFile2)
	assertFileDoesNotExist(t, ctx, buffer, testFile1)
	assertFileDoesNotExist(t, ctx, buffer, testFile2)

	// Create a new file on the buffer
	tofutestutils.Must(buffer.WriteFile(ctx, testFile2, []byte(testContent)))
	assertFileExists(t, ctx, backingStorage, testFile1)
	assertFileDoesNotExist(t, ctx, backingStorage, testFile2)
	assertFileDoesNotExist(t, ctx, buffer, testFile1)
	assertFileExists(t, ctx, buffer, testFile2)

	// Commit the changes
	tofutestutils.Must(buffer.Commit(ctx))

	// Check the changes
	assertFileDoesNotExist(t, ctx, backingStorage, testFile1)
	assertFileExists(t, ctx, backingStorage, testFile2)
	assertFileDoesNotExist(t, ctx, buffer, testFile1)
	assertFileExists(t, ctx, buffer, testFile2)
}

func TestDeepDirectories(t *testing.T) {
	const testContent = "Hello world!"
	var testFile1 = indexstorage.Path(path.Join("test1", "test2", "test3", "test.txt"))
	var testFile2 = indexstorage.Path(path.Join("test3", "test2", "test1", "test2.txt"))
	backingDir := t.TempDir()
	backingStorage := tofutestutils.Must2(filesystemstorage.New(backingDir))
	buffer := tofutestutils.Must2(bufferedstorage.New(logger.NewTestLogger(t), t.TempDir(), backingStorage, 25))

	ctx := tofutestutils.Context(t)

	tofutestutils.Must(buffer.WriteFile(ctx, testFile1, []byte(testContent)))
	tofutestutils.Must(buffer.WriteFile(ctx, testFile2, []byte(testContent)))
	tofutestutils.Must(buffer.Commit(ctx))
}

func TestSubdir(t *testing.T) {
	const testContent = "Hello world!"
	backingDir := t.TempDir()
	backingStorage := tofutestutils.Must2(filesystemstorage.New(backingDir))
	buffer := tofutestutils.Must2(bufferedstorage.New(logger.NewTestLogger(t), t.TempDir(), backingStorage, 25))

	ctx := tofutestutils.Context(t)

	subdir := tofutestutils.Must2(buffer.Subdirectory(ctx, "test"))

	tofutestutils.Must(subdir.WriteFile(ctx, "test.txt", []byte(testContent)))

	storedContents := assertFileExists(t, ctx, buffer, "test/test.txt")
	if string(storedContents) != testContent {
		t.Fatalf("Incorrect file contents found: %s", storedContents)
	}
	tofutestutils.Must(buffer.Commit(ctx))
	assertFileExists(t, ctx, backingStorage, "test/test.txt")

	subdir2 := tofutestutils.Must2(subdir.Subdirectory(ctx, "test"))
	tofutestutils.Must(subdir2.WriteFile(ctx, "test.txt", []byte(testContent)))
	storedContents = assertFileExists(t, ctx, buffer, "test/test/test.txt")
	if string(storedContents) != testContent {
		t.Fatalf("Incorrect file contents found: %s", storedContents)
	}
	tofutestutils.Must(buffer.Commit(ctx))
	assertFileExists(t, ctx, backingStorage, "test/test/test.txt")

	tofutestutils.Must(subdir2.RemoveAll(ctx, ""))

	assertFileDoesNotExist(t, ctx, buffer, "test/test/test.txt")
	assertFileExists(t, ctx, buffer, "test/test.txt")
}

func assertFileDoesNotExist(t *testing.T, ctx context.Context, storage indexstorage.API, file indexstorage.Path) {
	t.Helper()
	_, err := storage.ReadFile(ctx, file)
	if err == nil {
		t.Fatalf("reading a non-existent file %s did not fail", file)
	}
	if !os.IsNotExist(err) {
		t.Fatalf("unexpected error type %T returned from reading a non-existent file %s (%v)", err, file, err)
	}
}

func assertFileExists(t *testing.T, ctx context.Context, storage indexstorage.API, file indexstorage.Path) []byte {
	t.Helper()
	content, err := storage.ReadFile(ctx, file)
	if err != nil {
		t.Fatalf("unexpected error file reading %s (%v)", file, err)
	}
	return content
}
