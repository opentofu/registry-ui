// Package git provides utilities for managing git repositories with worktree support.
//
// This package uses a hybrid approach for git operations:
//
// # go-git Library (github.com/go-git/go-git/v5)
//
// Used for core git operations:
//   - Repository cloning (EnsureCloned)
//   - Fetching tags (FetchTags)
//
// # exec.Command (git CLI)
//
// Used for worktree operations:
//   - Adding worktrees (AddWorktree)
//   - Removing worktrees (RemoveWorktree)
//   - Pruning stale worktrees
//
// The reason for using this dual-approach is that go-git does not implement git worktree commands.
//
// # Why Not Pure go-git?
//
// We chose to use go-git wherever possible to maintain
// type safety and avoid parsing command output, but fall back to exec.Command
// only where go-git lacks functionality.
package git
