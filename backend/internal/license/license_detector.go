package license

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
)

// List is a list of licenses found in a repository.
//
// swagger:model LicenseList
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

// License describes a license found in a repository. Note: the license detection is best effort. When displaying the
// license to the user, always show a link to the actual license and warn users that they have to inspect the license
// themselves.
//
// swagger:model License
type License struct {
	// SPDX is the SPDX identifier for the license.
	// required: true
	SPDX string `json:"spdx"`
	// Confidence indicates how accurate the license detection is.
	// required: true
	Confidence float32 `json:"confidence"`
	// IsCompatible signals if the license is compatible with the OpenTofu project.
	// required: true
	IsCompatible bool `json:"is_compatible"`
	// File holds the file in the repository where the license was detected.
	// required: true
	File string `json:"file"`
	// Link may contain a link to the license file for humans to view. This may be empty.
	// required: false
	Link string `json:"link,omitempty"`
}

type Detector func(dir string, url string) (List, error)

func NewDetector(licenses []string) Detector {
	confidenceThreshold := float32(0.85)
	confidenceOverrideThreshold := float32(0.98)

	licenseMap := map[string]struct{}{}
	for _, license := range licenses {
		licenseMap[strings.ToLower(license)] = struct{}{}
	}

	return func(dir, url string) (List, error) {
		lf, err := filer.FromDirectory(dir)
		if err != nil {
			return nil, err
		}

		licenses, err := licensedb.Detect(lf)
		if err != nil {
			if errors.Is(err, licensedb.ErrNoLicenseFound) {
				return nil, nil
			}
			return nil, fmt.Errorf("error detecting licenses: %w", err)
		}
		var result []License
		for license, match := range licenses {
			if strings.HasPrefix(license, "deprecated_") {
				continue
			}
			if match.Confidence >= confidenceThreshold {
				_, isCompatible := licenseMap[strings.ToLower(license)]
				l := License{
					SPDX:         license,
					Confidence:   match.Confidence,
					IsCompatible: isCompatible,
					File:         match.File,
					Link:         url + "/" + match.File,
				}
				if match.Confidence >= confidenceOverrideThreshold {
					return []License{
						l,
					}, nil
				}
				result = append(result, l)
			}
		}
		return result, nil
	}
}
