package license

import "testing"

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
