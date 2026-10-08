package license

import (
	"testing"

	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/opentofu/registry-ui/pkg/config"
)

func TestIsOSIApproved(t *testing.T) {
	tests := []struct {
		spdxID string
		want   bool
	}{
		{spdxID: "MIT", want: true},
		{spdxID: "GPL-3.0-only", want: true},
		{spdxID: "AGPL-3.0-only", want: true},
		{spdxID: "BSL-1.0", want: true},
		{spdxID: "BUSL-1.1", want: false},
		{spdxID: "MIT-feh", want: false},
		{spdxID: "GPL-2.0+", want: true},
		{spdxID: "unknown", want: false},
		{spdxID: "", want: false},
		{spdxID: "gPl-3.0-OnLy", want: true},
	}

	for _, test := range tests {
		t.Run(test.spdxID, func(t *testing.T) {
			if got := isOSIApproved(test.spdxID); got != test.want {
				t.Errorf("isOSIApproved(%q) = %t, want %t", test.spdxID, got, test.want)
			}
		})
	}
}

func TestBuildLicenseFileMapUsesOSIApproval(t *testing.T) {
	detector, err := New(config.LicenseConfig{}, nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	matches := []licensedb.Match{
		{License: "GPL-3.0-only", Confidence: 1, File: "LICENSE"},
		{License: "BUSL-1.1", Confidence: 1, File: "LICENSE"},
	}
	licenses := detector.buildLicenseFileMap(matches, "https://github.com/example/module")["LICENSE"]

	if len(licenses) != len(matches) {
		t.Fatalf("detected licenses = %d, want %d", len(licenses), len(matches))
	}
	if !licenses[0].IsCompatible {
		t.Error("GPL-3.0-only should be OSI-approved")
	}
	if licenses[1].IsCompatible {
		t.Error("BUSL-1.1 should not be OSI-approved")
	}
}
