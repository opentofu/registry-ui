package moduleschema_test

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/registry-ui/internal/moduleindex/moduleschema"
	"github.com/opentofu/tofudl"
	"github.com/opentofu/tofudl/mockmirror"
)

func getTofuDL(t *testing.T) tofudl.Downloader {
	contents, err := os.ReadFile("./testtofu/tofu")
	if err != nil {
		t.Fatalf("Failed to read tofu binary (did you run go generate?)")
	}
	mirror := mockmirror.NewFromBinary(t, contents)
	downloader, err := tofudl.New(
		tofudl.ConfigGPGKey(mirror.GPGKey()),
		tofudl.ConfigAPIURL(mirror.APIURL()),
		tofudl.ConfigDownloadMirrorURLTemplate(mirror.DownloadMirrorURLTemplate()),
	)
	if err != nil {
		t.Fatalf("Failed to create tofudl instance (%v)", err)
	}
	return downloader
}

func TestVariables(t *testing.T) {
	tofuPath := t.TempDir() + "/tofu"
	if runtime.GOOS == "windows" {
		tofuPath += ".exe"
	}

	extractor, err := moduleschema.NewExternalTofuExtractor(
		moduleschema.ExternalTofuExtractorConfig{
			TofuPath: tofuPath,
		},
		logger.NewTestLogger(t),
		getTofuDL(t),
	)
	if err != nil {
		t.Fatalf("Failed to create tofu extractor (%v)", err)
	}

	metadata, err := extractor.Extract(context.Background(), "./testdata/variables")
	if err != nil {
		t.Fatalf("Failed to extract metadata (%v)", err)
	}

	for _, variable := range []string{
		"untyped", "str", "int", "def", "desc",
	} {
		if _, ok := metadata.RootModule.Variables[variable]; !ok {
			t.Fatalf("The '%s' variable was not found!", variable)
		}
	}
	if metadata.RootModule.Variables["desc"].Description != "This is a test" {
		t.Fatalf("The 'desc' variable has an incorrect description: %s", metadata.RootModule.Variables["desc"].Description)
	}
	if metadata.RootModule.Variables["def"].Default != 42.0 {
		t.Fatalf("The 'def' variable has an incorrect default value: %d", metadata.RootModule.Variables["desc"].Default)
	}
}

func TestModuleCall(t *testing.T) {
	tofuPath := t.TempDir() + "/tofu"
	if runtime.GOOS == "windows" {
		tofuPath += ".exe"
	}

	extractor, err := moduleschema.NewExternalTofuExtractor(
		moduleschema.ExternalTofuExtractorConfig{
			TofuPath: tofuPath,
		},
		logger.NewTestLogger(t),
		getTofuDL(t),
	)
	if err != nil {
		t.Fatalf("Failed to create tofu extractor (%v)", err)
	}

	metadata, err := extractor.Extract(context.Background(), "./testdata/modulecall")
	if err != nil {
		t.Fatalf("Failed to extract metadata (%v)", err)
	}

	_, ok := metadata.RootModule.ModuleCalls["foo"]
	if !ok {
		t.Fatalf("Module call not found.")
	}
	if metadata.RootModule.ModuleCalls["foo"].Source != "./module" {
		t.Fatalf("Incorrect module source: %s", metadata.RootModule.ModuleCalls["foo"].Source)
	}
}
