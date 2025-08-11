package license

import (
	"reflect"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/opentofu/registry-ui/pkg/config"
)

func TestShouldIgnoreLicenseFile(t *testing.T) {
	tests := []struct {
		path           string
		shouldIgnore   bool
		expectedReason string
	}{
		{"LICENSE", false, ""},
		{"docs/LICENSE", false, ""},
		{"THIRD_PARTY_LICENSES.txt", true, "ignored filename"},
		{"vendor/pkg/LICENSE", true, "dependency directory"},
		{"node_modules/lib/LICENSE", true, "dependency directory"},
		{"examples/LICENSE", true, "examples/test directory"},
		{"test/fixtures/LICENSE", true, "examples/test directory"},
		{"PATENTS", true, "ignored filename"},
		{"NOTICE", true, "ignored filename"},
		{"src/LICENSE", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			shouldIgnore, reason := shouldIgnoreLicenseFile(tt.path)
			if shouldIgnore != tt.shouldIgnore {
				t.Errorf("shouldIgnoreLicenseFile(%q) = %v, want %v", tt.path, shouldIgnore, tt.shouldIgnore)
			}
			if reason != tt.expectedReason {
				t.Errorf("shouldIgnoreLicenseFile(%q) reason = %q, want %q", tt.path, reason, tt.expectedReason)
			}
		})
	}
}

func TestIsDocumentationDirectory(t *testing.T) {
	tests := []struct {
		path  string
		isDoc bool
	}{
		{"LICENSE", false},
		{"docs/LICENSE", true},
		{"doc/LICENSE", true},
		{"website/docs/LICENSE", true},
		{"documentation/LICENSE", true},
		{"src/docs/LICENSE", false}, // nested, shouldn't count
		{"mydocs/LICENSE", false},
		{"docs-old/LICENSE", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isDocumentationDirectory(tt.path)
			if result != tt.isDoc {
				t.Errorf("isDocumentationDirectory(%q) = %v, want %v", tt.path, result, tt.isDoc)
			}
		})
	}
}

func TestPathDepth(t *testing.T) {
	tests := []struct {
		path  string
		depth int
	}{
		{"LICENSE", 0},
		{"docs/LICENSE", 1},
		{"website/docs/LICENSE", 2},
		{"very/deep/nested/path/LICENSE", 4},
		{"", 0},
		{".", 0},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := pathDepth(tt.path)
			if result != tt.depth {
				t.Errorf("pathDepth(%q) = %v, want %v", tt.path, result, tt.depth)
			}
		})
	}
}

func TestLicenseFileSorting(t *testing.T) {
	files := []string{
		"src/LICENSE",          // depth 1, not docs
		"LICENSE",              // depth 0, not docs
		"docs/LICENSE",         // depth 1, docs
		"website/docs/LICENSE", // depth 2, docs
		"lib/LICENSE",          // depth 1, not docs
	}

	// Expected order: docs first, then by depth, then alphabetical
	expected := []string{
		"docs/LICENSE",         // docs, depth 1
		"website/docs/LICENSE", // docs, depth 2
		"LICENSE",              // not docs, depth 0
		"lib/LICENSE",          // not docs, depth 1 (alphabetically before src)
		"src/LICENSE",          // not docs, depth 1
	}

	slices.SortFunc(files, func(a, b string) int {
		aIsDoc := isDocumentationDirectory(a)
		bIsDoc := isDocumentationDirectory(b)

		if aIsDoc != bIsDoc {
			if aIsDoc {
				return -1
			}
			return 1
		}

		aDepth := pathDepth(a)
		bDepth := pathDepth(b)
		if aDepth != bDepth {
			return aDepth - bDepth
		}

		return slices.Compare([]rune(a), []rune(b))
	})

	if !reflect.DeepEqual(files, expected) {
		t.Errorf("Sorting failed:\nGot:      %v\nExpected: %v", files, expected)
	}
}

func TestDetectIntegration(t *testing.T) {
	testFS := fstest.MapFS{
		"LICENSE": &fstest.MapFile{
			Data: []byte("Some license text"),
		},
		"vendor/LICENSE": &fstest.MapFile{
			Data: []byte("Should be ignored"),
		},
	}

	detector, err := New(
		config.LicenseConfig{
			CompatibleLicenses: []string{"Apache-2.0", "MIT"},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	_, err = detector.Detect(t.Context(), testFS, "https://github.com/blah/terraform-provider-blah")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
}

func TestDetectEmptyRepository(t *testing.T) {
	testFS := fstest.MapFS{}

	detector, err := New(
		config.LicenseConfig{
			CompatibleLicenses: []string{"Apache-2.0", "MIT"},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	licenses, err := detector.Detect(t.Context(), testFS, "https://github.com/blah/terraform-provider-blah")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(licenses) != 0 {
		t.Errorf("Expected no licenses for empty repository, got %d", len(licenses))
	}
}
