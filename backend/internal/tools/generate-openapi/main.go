package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "run", "github.com/go-swagger/go-swagger/cmd/swagger@v0.31.0", "generate", "spec", "-o", "server/openapi.yml", "-m")
	cmd.Env = append(os.Environ(), "SWAGGER_GENERATE_EXTENSION=false")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		log.Print(err)
		os.Exit(1)
	}
}
