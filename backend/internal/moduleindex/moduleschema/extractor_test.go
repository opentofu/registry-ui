package moduleschema_test

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/tofudl"
	"github.com/opentofu/tofudl/mockmirror"

	"github.com/opentofu/registry-ui/internal/moduleindex/moduleschema"
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

	// Verify simple_string is a string type
	if _, ok := metadata.RootModule.Variables["simple_string"]; !ok {
		t.Fatalf("simple_string variable not found")
	}
	simpleStringType := metadata.RootModule.Variables["simple_string"].Type
	if simpleStringType.FriendlyName() != "string" {
		t.Errorf("simple_string type = %v, want string", simpleStringType.FriendlyName())
	}

	if _, ok := metadata.RootModule.Variables["list_of_strings"]; !ok {
		t.Fatalf("list_of_strings variable not found")
	}
	listType := metadata.RootModule.Variables["list_of_strings"].Type
	if !listType.IsListType() {
		t.Errorf("list_of_strings type should be a list type, got: %v", listType.FriendlyName())
	}

	if _, ok := metadata.RootModule.Variables["map_of_objects"]; !ok {
		t.Fatalf("map_of_objects variable not found")
	}
	mapType := metadata.RootModule.Variables["map_of_objects"].Type
	if !mapType.IsMapType() {
		t.Errorf("map_of_objects should be a map type, got: %v", mapType.FriendlyName())
	}

	if _, ok := metadata.RootModule.Variables["list_of_objects"]; !ok {
		t.Fatalf("list_of_objects variable not found")
	}
}

// TestComplexTypesJSONMarshaling verifies that complex types can be marshaled back to JSON correctly.
// This means that the UI in search.opentofu.org will receive the types as their JSON representation.
func TestComplexTypesJSONMarshaling(t *testing.T) {
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

	metadata, err := extractor.Extract(context.Background(), "./testdata/complex_types")
	if err != nil {
		t.Fatalf("Failed to extract metadata (%v)", err)
	}

	tests := []struct {
		varName      string
		expectedJSON string
	}{
		{
			varName:      "simple_string",
			expectedJSON: `"string"`,
		},
		{
			varName:      "list_of_strings",
			expectedJSON: `["list","string"]`,
		},
		{
			varName:      "map_of_objects",
			expectedJSON: `["map",["object",{"filesystem":"string","size":"number"},["filesystem"]]]`,
		},
		{
			varName:      "list_of_objects",
			expectedJSON: `["list",["object",{"protocol":"string","range":"string","rule_type":"string"}]]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.varName, func(t *testing.T) {
			variable, ok := metadata.RootModule.Variables[tt.varName]
			if !ok {
				t.Fatalf("Variable %s not found", tt.varName)
			}

			// Marshal the type back to JSON (same as generator.go does)
			jsonBytes, err := variable.Type.MarshalJSON()
			if err != nil {
				t.Fatalf("Failed to marshal type for %s: %v", tt.varName, err)
			}

			jsonStr := string(jsonBytes)
			t.Logf("%s type JSON: %s", tt.varName, jsonStr)

			// Verify exact match
			if jsonStr != tt.expectedJSON {
				t.Errorf("JSON mismatch for %s:\nGot:      %s\nExpected: %s",
					tt.varName, jsonStr, tt.expectedJSON)
			}
		})
	}
}
