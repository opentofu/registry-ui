package moduleschema

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/tofudl"
)

// ExternalTofuExtractorConfig is the configuration for the Extractor using the tofu binary.
type ExternalTofuExtractorConfig struct {
	// TofuPath holds the path to the tofu binary. If empty, path/path.exe will be used via the OS path lookup.
	TofuPath string `json:"tofu_path"`
}

// NewExternalTofuExtractor creates an Extractor that launches the tofu binary with the "metadata dump" command
// to extract metadata. If you pass a downloader, the extractor will attempt to download an appropriate tofu version
// if needed.
func NewExternalTofuExtractor(
	config ExternalTofuExtractorConfig,
	log logger.Logger,
	downloader tofudl.Downloader,
) (Extractor, error) {
	return &externalTofuExtractor{
		config:     config,
		logger:     log.WithName("module-extractor"),
		downloader: downloader,
	}, nil
}

type externalTofuExtractor struct {
	config     ExternalTofuExtractorConfig
	logger     logger.Logger
	downloader tofudl.Downloader
}

func (e *externalTofuExtractor) Extract(ctx context.Context, moduleDirectory string) (Schema, error) {
	command := e.config.TofuPath
	if command == "" {
		command = tofuName
	}

	absName, err := exec.LookPath(command)
	needsDownload := false
	if err != nil {
		needsDownload = true
	} else {
		absName, err = filepath.Abs(absName)
		if err != nil {
			needsDownload = true
		}
	}
	command = absName

	if needsDownload {
		if e.downloader == nil {
			return Schema{}, &SchemaExtractionFailedError{
				nil,
				fmt.Errorf("cannot find working tofu binary and no downloader is configured"),
			}
		}
		binary, err := e.downloader.Download(ctx)
		if err != nil {
			return Schema{}, &SchemaExtractionFailedError{
				nil,
				err,
			}
		}
		if e.config.TofuPath == "" {
			e.config.TofuPath = path.Join(os.TempDir(), tofuName)
		}
		if err := os.WriteFile(e.config.TofuPath, binary, 0755); err != nil {
			return Schema{}, &SchemaExtractionFailedError{
				nil,
				fmt.Errorf("cannot write tofu binary to %s (%w)", e.config.TofuPath, err),
			}
		}
		absName, err = filepath.Abs(e.config.TofuPath)
		if err != nil {
			return Schema{}, fmt.Errorf("cannot determine absolute path for %s", e.config.TofuPath)
		}
		command = absName
	}

	cmd := exec.Command(command, "metadata", "dump", "-json")
	output := &bytes.Buffer{}
	cmd.Stdout = output
	stderr := &stderrLogger{
		"tofu metadata dump: ",
		ctx,
		e.logger,
		nil,
		true,
	}
	// TODO: why is this writing to stderr?
	cmd.Stderr = output
	defer func() {
		_ = stderr.Close()
	}()
	cmd.Dir, err = filepath.Abs(moduleDirectory)
	if err != nil {
		return Schema{}, &SchemaExtractionFailedError{
			nil,
			fmt.Errorf("cannot determine absolute module directory (%w)", err),
		}
	}
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return Schema{}, &SchemaExtractionFailedError{
				output.Bytes(),
				err,
			}
		}
		if exitErr.ExitCode() != 0 {
			e.logger.Debug(ctx, "tofu metadata dump exited with a non-zero exit code (%d). Output:\n%s", exitErr.ExitCode(), output.Bytes())
			return Schema{}, &SchemaExtractionFailedError{
				output.Bytes(),
				err,
			}
		}
	}
	var result Schema
	outputBytes := output.Bytes()
	if err := json.Unmarshal(outputBytes, &result); err != nil {
		return Schema{}, &SchemaExtractionFailedError{
			outputBytes,
			fmt.Errorf("metadata dump JSON decoding failed (%w)", err),
		}
	}

	return result, nil
}

type stderrLogger struct {
	prefix        string
	ctx           context.Context
	loggerBackend logger.Logger
	buf           []byte
	error         bool
}

func (s *stderrLogger) Write(p []byte) (n int, err error) {
	s.buf = append(s.buf, p...)
	lastNewline := 0
	for i, b := range s.buf {
		if b == '\n' {
			s.writeLine(string(s.buf[lastNewline:i]))
			lastNewline = i + 1
		}
	}
	s.buf = s.buf[lastNewline:]
	return len(p), nil
}

func (s *stderrLogger) Close() error {
	if len(s.buf) > 0 {
		s.writeLine(string(s.buf))
	}
	return nil
}

func (s *stderrLogger) writeLine(line string) {
	s.loggerBackend.Error(s.ctx, s.prefix+line)
}
