package license

import "strings"

// List is a list of licenses found in a repository.
type List []License

func (l List) HasCompatible() bool {
	for _, license := range l {
		if license.IsCompatible {
			return true
		}
	}
	return false
}

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

func (l List) Explain() string {
	if l.IsRedistributable() {
		licenses := map[string]struct{}{}
		for _, license := range l {
			licenses[license.SPDX] = struct{}{}
		}
		var licenseList []string
		for license := range licenses {
			licenseList = append(licenseList, license)
		}
		return "This project is redistributable because it contains the following licenses: " + strings.Join(licenseList, ", ")
	}
	if len(l) == 0 {
		return "This project is not redistributable because it contains no licenses."
	}
	incompatibleLicenses := map[string]struct{}{}
	compatibleLicenses := map[string]struct{}{}
	for _, license := range l {
		if !license.IsCompatible {
			incompatibleLicenses[license.SPDX] = struct{}{}
		} else {
			compatibleLicenses[license.SPDX] = struct{}{}
		}
	}
	var incompatibleLicenseList []string
	for license := range incompatibleLicenses {
		incompatibleLicenseList = append(incompatibleLicenseList, license)
	}
	var compatibleLicenseList []string
	for license := range compatibleLicenses {
		compatibleLicenseList = append(compatibleLicenseList, license)
	}
	if len(compatibleLicenses) > 0 {
		return "This project is not redistributable because it contains the following incompatible licenses: " + strings.Join(incompatibleLicenseList, ",") + ". It also contains the following compatible licenses: " + strings.Join(compatibleLicenseList, ", ")
	}
	return "This project is not redistributable because it contains the following incompatible licenses: " + strings.Join(incompatibleLicenseList, ",")
}

func (l List) String() string {
	str := make([]string, len(l))
	for i, license := range l {
		str[i] = license.SPDX
	}
	return strings.Join(str, ", ")
}
