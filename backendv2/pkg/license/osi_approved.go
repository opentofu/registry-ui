package license

import "buf.build/go/spdx"

func isOSIApproved(spdxID string) bool {
	license, ok := spdx.LicenseForID(spdxID)
	return ok && license.OSIApproved
}
