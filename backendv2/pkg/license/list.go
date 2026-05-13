package license

import (
	"strings"

	"github.com/opentofu/registry-ui/pkg/config"
)

// List is a list of licenses found in a repository.
type List []License

func (l List) HasIncompatible() bool {
	for _, license := range l {
		if !license.IsCompatible {
			return true
		}
	}
	return false
}

func (l List) IsRedistributable() bool {
	// We check for incompatible licenses to avoid mistaking a license in a subdirectory for the main license
	// of the project.
	return len(l) > 0 && !l.HasIncompatible()
}

// Selected returns the authoritative subset of licenses:
// if any license has confidence >= overrideThreshold, return only that license;
// otherwise return all licenses with confidence >= baseThreshold.
// The list is assumed to be sorted by the detection order priority.
func (l List) Selected(cfg config.LicenseConfig) List {
	for _, lic := range l {
		if lic.Confidence >= cfg.ConfidenceOverrideThreshold {
			return List{lic}
		}
	}
	var result List
	for _, lic := range l {
		if lic.Confidence >= cfg.ConfidenceThreshold {
			result = append(result, lic)
		}
	}
	return result
}

func (l List) String() string {
	str := make([]string, len(l))
	for i, license := range l {
		str[i] = license.SPDX
	}
	return strings.Join(str, ", ")
}
