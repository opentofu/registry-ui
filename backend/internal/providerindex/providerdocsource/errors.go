package providerdocsource

import (
	_ "embed"
	"text/template"
)

//go:embed err_blocked.md.tpl
var errorMessageBlocked []byte

var errorMessageBlockedTemplate = template.Must(template.New("").Parse(string(errorMessageBlocked)))

//go:embed err_incompatible_license.md
var errorIncompatibleLicense []byte

//go:embed err_file_too_large.md.tpl
var errorTooLarge []byte
