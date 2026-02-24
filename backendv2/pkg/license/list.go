package license

import "strings"

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

func (l List) String() string {
	str := make([]string, len(l))
	for i, license := range l {
		str[i] = license.SPDX
	}
	return strings.Join(str, ", ")
}
