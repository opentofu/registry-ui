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

	_, ok = metadata.ProviderConfig["opentofu"]
	if !ok {
		t.Fatalf("Provider call not found.")
	}
	if metadata.ProviderConfig["opentofu"].VersionConstraint != "1.6.0" {
		t.Fatalf("Incorrect provider name: %s != %s", metadata.ProviderConfig["opentofu"].VersionConstraint, "1.6.0")
	}
}

// TestComplexTypes tests parsing modules with complex type expressions (issue #3442)
// These types include map(object({...})), list(object({...})), etc.
func TestComplexTypes(t *testing.T) {
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

	metadata, err := extractor.Extract(t.Context(), "./testdata/complex_types")
	if err != nil {
		t.Fatalf("Failed to extract metadata (%v)", err)
	}

	// Verify simple types are still strings
	if _, ok := metadata.RootModule.Variables["simple_string"]; !ok {
		t.Fatalf("simple_string variable not found")
	}
	if metadata.RootModule.Variables["simple_string"].Type != "string" {
		t.Errorf("simple_string type = %v, want string", metadata.RootModule.Variables["simple_string"].Type)
	}

	if _, ok := metadata.RootModule.Variables["list_of_strings"]; !ok {
		t.Fatalf("list_of_strings variable not found")
	}
	listType := metadata.RootModule.Variables["list_of_strings"].Type
	if _, isSlice := listType.([]any); !isSlice {
		// If it's still a string, that's also acceptable (depends on tofu version)
		if _, isString := listType.(string); !isString {
			t.Errorf("list_of_strings type = %T, want []any or string", listType)
		}
	}

	if _, ok := metadata.RootModule.Variables["map_of_objects"]; !ok {
		t.Fatalf("map_of_objects variable not found")
	}
	mapType := metadata.RootModule.Variables["map_of_objects"].Type
	if mapType == nil {
		t.Errorf("map_of_objects type should not be nil")
	}

	if _, ok := metadata.RootModule.Variables["list_of_objects"]; !ok {
		t.Fatalf("list_of_objects variable not found")
	}

	t.Logf("list_of_objects type: %T", metadata.RootModule.Variables["list_of_objects"].Type)
	t.Logf("map_of_objects type: %T", metadata.RootModule.Variables["map_of_objects"].Type)
}
