package main

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	testTofuDir := path.Join(wd, "testtofu")
	repoDir := path.Join(testTofuDir, "repo")
	binaryName := "tofu"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := path.Join(testTofuDir, binaryName)

	if err := os.MkdirAll(testTofuDir, 0755); err != nil {
		panic(err)
	}

	buildTofu(repoDir, testTofuDir, binaryPath)
}

func buildTofu(repoDir string, testTofuDir string, binaryPath string) {
	if _, err := os.Stat(binaryPath); err == nil {
		return
	}
	if _, err := os.Stat(repoDir); err != nil {
		runCommand(testTofuDir, "git", "clone", "https://github.com/opentofu/opentofu.git", repoDir)
	}

	runCommand(repoDir, "git", "pull")
	runCommand(repoDir, "git", "checkout", "experiment/json_config_dump_rebased")
	runCommand(repoDir, "go", "build", "-o", filepath.ToSlash(binaryPath), "github.com/opentofu/opentofu/cmd/tofu")
}

func runCommand(wd string, argv ...string) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = wd
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 0 {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}
