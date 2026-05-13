package license

import (
	"testing"

	"github.com/opentofu/registry-ui/pkg/config"
)

func TestListSelected(t *testing.T) {
	cfg := config.LicenseConfig{
		ConfidenceThreshold:         0.85,
		ConfidenceOverrideThreshold: 0.98,
	}

	t.Run("empty list returns empty", func(t *testing.T) {
		var l List
		result := l.Selected(cfg)
		if len(result) != 0 {
			t.Errorf("expected empty list, got %d items", len(result))
		}
	})

	t.Run("no license above override threshold returns only those above base threshold", func(t *testing.T) {
		l := List{
			{SPDX: "MIT", Confidence: 0.90},
			{SPDX: "Apache-2.0", Confidence: 0.87},
			{SPDX: "GPL-2.0", Confidence: 0.60}, // below base, not selected
		}
		result := l.Selected(cfg)
		if len(result) != 2 {
			t.Errorf("expected 2 items, got %d", len(result))
		}
	})

	t.Run("license at exactly override threshold returns only that one", func(t *testing.T) {
		l := List{
			{SPDX: "MIT", Confidence: 0.98},
			{SPDX: "Apache-2.0", Confidence: 0.87},
			{SPDX: "GPL-2.0", Confidence: 0.60},
		}
		result := l.Selected(cfg)
		if len(result) != 1 {
			t.Errorf("expected 1 item, got %d", len(result))
		}
		if result[0].SPDX != "MIT" {
			t.Errorf("expected MIT, got %s", result[0].SPDX)
		}
	})

	t.Run("license above override threshold returns only that one", func(t *testing.T) {
		l := List{
			{SPDX: "MIT", Confidence: 0.99},
			{SPDX: "Apache-2.0", Confidence: 0.92},
			{SPDX: "GPL-2.0", Confidence: 0.50},
		}
		result := l.Selected(cfg)
		if len(result) != 1 {
			t.Errorf("expected 1 item, got %d", len(result))
		}
		if result[0].SPDX != "MIT" {
			t.Errorf("expected MIT, got %s", result[0].SPDX)
		}
	})

	t.Run("returns first match when multiple exceed override threshold", func(t *testing.T) {
		l := List{
			{SPDX: "MIT", Confidence: 0.99},
			{SPDX: "Apache-2.0", Confidence: 0.99},
		}
		result := l.Selected(cfg)
		if len(result) != 1 {
			t.Errorf("expected 1 item, got %d", len(result))
		}
		if result[0].SPDX != "MIT" {
			t.Errorf("expected first item MIT, got %s", result[0].SPDX)
		}
	})

	t.Run("nothing above base threshold returns empty", func(t *testing.T) {
		l := List{
			{SPDX: "GPL-2.0", Confidence: 0.60},
			{SPDX: "MIT", Confidence: 0.70},
		}
		result := l.Selected(cfg)
		if len(result) != 0 {
			t.Errorf("expected empty, got %d items", len(result))
		}
	})
}
