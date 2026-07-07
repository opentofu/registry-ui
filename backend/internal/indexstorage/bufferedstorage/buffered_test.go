package bufferedstorage_test

import (
	"bytes"
	"context"
	"os"
	"path"
	"testing"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/indexstorage/bufferedstorage"
	"github.com/opentofu/registry-ui/internal/indexstorage/filesystemstorage"
)

func TestSimpleCommit(t *testing.T) {
	const testContent = "Hello world!"
	const testFile1 = "test.txt"

	backingDir := t.TempDir()
	backingStorage, err := filesystemstorage.New(backingDir)
	if err != nil {
		t.Fatalf("failed to create backing storage: %v", err)
	}

	buffer1, err := bufferedstorage.New(logger.NewTestLogger(t), t.TempDir(), backingStorage, 25)
	if err != nil {
		t.Fatalf("failed to create buffered storage: %v", err)
	}

	ctx := t.Context()

	if err := buffer1.WriteFile(ctx, testFile1, []byte(testContent)); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	assertFileDoesNotExist(t, ctx, backingStorage, testFile1)
	assertFileExists(t, ctx, buffer1, testFile1)

	if err := buffer1.Commit(ctx); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

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
	backingStorage, err := filesystemstorage.New(backingDir)
	if err != nil {
		t.Fatalf("failed to create backing storage: %v", err)
	}
	buffer, err := bufferedstorage.New(logger.NewTestLogger(t), t.TempDir(), backingStorage, 25)
	if err != nil {
		t.Fatalf("failed to create buffered storage: %v", err)
	}

	ctx := t.Context()

	// Create a file in the test directory
	if err := backingStorage.WriteFile(ctx, testFile1, []byte(testContent)); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Remove the directory of the test file
	if err := buffer.RemoveAll(ctx, testDir); err != nil {
		t.Fatalf("failed to remove directory: %v", err)
	}

	// Check if it still exists on the backing storage, but not on the local one.
	assertFileExists(t, ctx, backingStorage, testFile1)
	assertFileDoesNotExist(t, ctx, backingStorage, testFile2)
	assertFileDoesNotExist(t, ctx, buffer, testFile1)
	assertFileDoesNotExist(t, ctx, buffer, testFile2)

	// Create a new file on the buffer
	if err := buffer.WriteFile(ctx, testFile2, []byte(testContent)); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	assertFileExists(t, ctx, backingStorage, testFile1)
	assertFileDoesNotExist(t, ctx, backingStorage, testFile2)
	assertFileDoesNotExist(t, ctx, buffer, testFile1)
	assertFileExists(t, ctx, buffer, testFile2)

	// Commit the changes
	if err := buffer.Commit(ctx); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

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
	backingStorage, err := filesystemstorage.New(backingDir)
	if err != nil {
		t.Fatalf("failed to create backing storage: %v", err)
	}
	buffer, err := bufferedstorage.New(logger.NewTestLogger(t), t.TempDir(), backingStorage, 25)
	if err != nil {
		t.Fatalf("failed to create buffered storage: %v", err)
	}

	ctx := t.Context()

	if err := buffer.WriteFile(ctx, testFile1, []byte(testContent)); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := buffer.WriteFile(ctx, testFile2, []byte(testContent)); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := buffer.Commit(ctx); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

func TestSubdir(t *testing.T) {
	const testContent = "Hello world!"
	backingDir := t.TempDir()
	backingStorage, err := filesystemstorage.New(backingDir)
	if err != nil {
		t.Fatalf("failed to create backing storage: %v", err)
	}
	buffer, err := bufferedstorage.New(logger.NewTestLogger(t), t.TempDir(), backingStorage, 25)
	if err != nil {
		t.Fatalf("failed to create buffered storage: %v", err)
	}

	ctx := t.Context()

	subdir, err := buffer.Subdirectory(ctx, "test")
	if err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	if err := subdir.WriteFile(ctx, "test.txt", []byte(testContent)); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	storedContents := assertFileExists(t, ctx, buffer, "test/test.txt")
	if string(storedContents) != testContent {
		t.Fatalf("Incorrect file contents found: %s", storedContents)
	}
	if err := buffer.Commit(ctx); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
	assertFileExists(t, ctx, backingStorage, "test/test.txt")

	subdir2, err := subdir.Subdirectory(ctx, "test")
	if err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}
	if err := subdir2.WriteFile(ctx, "test.txt", []byte(testContent)); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	storedContents = assertFileExists(t, ctx, buffer, "test/test/test.txt")
	if string(storedContents) != testContent {
		t.Fatalf("Incorrect file contents found: %s", storedContents)
	}
	if err := buffer.Commit(ctx); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
	assertFileExists(t, ctx, backingStorage, "test/test/test.txt")

	if err := subdir2.RemoveAll(ctx, ""); err != nil {
		t.Fatalf("failed to remove all: %v", err)
	}

	assertFileDoesNotExist(t, ctx, buffer, "test/test/test.txt")
	assertFileExists(t, ctx, buffer, "test/test.txt")
}

func TestSameContent(t *testing.T) {
	const testContent = "Hello world!"
	backingDir := t.TempDir()
	backingStorage, err := filesystemstorage.New(backingDir)
	if err != nil {
		t.Fatalf("failed to create backing storage: %v", err)
	}

	ctx := t.Context()

	if err := backingStorage.WriteFile(ctx, "test.txt", []byte(testContent)); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	localDir := t.TempDir()
	buffer, err := bufferedstorage.New(logger.NewTestLogger(t), localDir, backingStorage, 25)
	if err != nil {
		t.Fatalf("failed to create buffered storage: %v", err)
	}
	if _, err := os.Stat(path.Join(localDir, "test.txt")); err == nil {
		t.Fatalf("Test file was present before fetched.")
	}
	readData, err := buffer.ReadFile(ctx, "test.txt")
	if err != nil {
		t.Fatalf("Could not read test file: %v", err)
	}
	if !bytes.Equal(readData, []byte(testContent)) {
		t.Fatalf("The test file has incorrect contents: %s", readData)
	}

	if _, err = os.Stat(path.Join(localDir, "test.txt")); err != nil {
		t.Fatalf("Test file was not present before after fetching: %v", err)
	}
	if err := buffer.WriteFile(ctx, "test.txt", []byte(testContent)); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if buffer.UncommittedFiles() != 0 {
		t.Fatalf("Uncommitted files despite same content.")
	}
	if err := buffer.WriteFile(ctx, "test.txt", []byte(testContent+"!")); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if buffer.UncommittedFiles() != 1 {
		t.Fatalf("No or multiple uncommitted files despite different content!")
	}
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
