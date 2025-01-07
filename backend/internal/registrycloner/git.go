package registrycloner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// runGitCommand runs the provided git command in the specified directory.
func (c cloner) runGitCommand(ctx context.Context, args []string, dir string) error {
	stderr := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir

	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git command failed (%w; %s)", err, stderr.String())
	}

	return nil
}

// CloneGitRepo clones a git repository with the option for a shallow clone.
func (c cloner) cloneGitRepo(ctx context.Context) error {
	return c.runGitCommand(ctx, []string{"git", "clone", c.cfg.Repo, "--depth", "1", c.cfg.Directory}, "")
}

// FetchTags fetches the tags for the repository in the specified directory.
func (c cloner) fetchTags(ctx context.Context) error { //nolint:unused
	return c.runGitCommand(ctx, []string{"git", "fetch", "--tags", "--force"}, c.cfg.Directory)
}

// PullLatestMain pulls the latest changes from the main branch in the specified directory.
func (c cloner) pullLatest(ctx context.Context) error {
	return c.runGitCommand(ctx, []string{"git", "pull", "origin", c.cfg.Ref}, c.cfg.Directory)
}

// Switch switches to the specified reference in the repository.
func (c cloner) switchRef(ctx context.Context) error {
	return c.runGitCommand(ctx, []string{"git", "switch", "-d", c.cfg.Ref}, c.cfg.Directory)
}

func (c cloner) clean(ctx context.Context) error {
	return c.runGitCommand(ctx, []string{"git", "clean", "-fd"}, c.cfg.Directory)
}

func (c cloner) reset(ctx context.Context) error {
	return c.runGitCommand(ctx, []string{"git", "reset", "--hard"}, c.cfg.Directory)
}
