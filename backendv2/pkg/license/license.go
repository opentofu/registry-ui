package license

// License describes a license found in a repository. Note: the license detection is best effort. When displaying the
// license to the user, always show a link to the actual license and warn users that they have to inspect the license
// themselves.
type License struct {
	// SPDX is the SPDX identifier for the license.
	SPDX string `json:"spdx"`
	// Confidence indicates how accurate the license detection is.
	Confidence float32 `json:"confidence"`
	// IsCompatible signals whether the SPDX identifier is OSI-approved under the backend policy.
	IsCompatible bool `json:"is_compatible"`
	// File holds the file in the repository where the license was detected.
	File string `json:"file"`
	// Link may contain a link to the license file for humans to view. This may be empty.
	Link string `json:"link,omitempty"`
}
