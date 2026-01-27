package moduleschema_test

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/tofudl"
	"github.com/opentofu/tofudl/mockmirror"
	"github.com/zclconf/go-cty/cty"

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

	tests := []struct {
		varName  string
		expected cty.Type
	}{
		{
			varName:  "simple_string",
			expected: cty.String,
		},
		{
			varName:  "list_of_strings",
			expected: cty.List(cty.String),
		},
		{
			varName: "map_of_objects",
			expected: cty.Map(cty.ObjectWithOptionalAttrs(map[string]cty.Type{
				"filesystem": cty.String,
				"size":       cty.Number,
			}, []string{"filesystem"})),
		},
		{
			varName: "list_of_objects",
			expected: cty.List(cty.Object(map[string]cty.Type{
				"protocol":  cty.String,
				"range":     cty.String,
				"rule_type": cty.String,
			})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.varName, func(t *testing.T) {
			variable, ok := metadata.RootModule.Variables[tt.varName]
			if !ok {
				t.Fatalf("Variable %s not found", tt.varName)
			}

			if !variable.Type.Equals(tt.expected) {
				t.Errorf("Type mismatch for %s: got %v, want %v", tt.varName, variable.Type, tt.expected)
			}
		})
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

// TestInferredTypeVariable tests that variables without explicit type declarations (type inferred from default)
// are properly handled. OpenTofu's metadata dump omits the "type" field for such variables, resulting in cty.NilType.
// This was causing panics in production (coalfire-cf/organization/aws v1.1.9).
// https://github.com/opentofu/registry-ui/actions/runs/21405798331/job/61629184889
func TestInferredTypeVariable(t *testing.T) {
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

	// Verify that the variable with inferred type has cty.NilType in the raw metadata
	variable, ok := metadata.RootModule.Variables["inferred_type_from_default"]
	if !ok {
		t.Fatalf("Variable 'inferred_type_from_default' not found in metadata")
	}

	if variable.Type != cty.NilType {
		t.Errorf("Expected cty.NilType for inferred type variable in raw metadata, got %v", variable.Type)
	}

	// Verify the default value is present (used for type inference in generator)
	if variable.Default == nil {
		t.Error("Expected default value to be present for type inference")
	} else if str, ok := variable.Default.(string); !ok || str != "ALL" {
		t.Errorf("Expected default value 'ALL', got %v", variable.Default)
	}

	t.Logf("Successfully extracted metadata with cty.NilType variable (will be inferred to cty.String by generator)")
}
