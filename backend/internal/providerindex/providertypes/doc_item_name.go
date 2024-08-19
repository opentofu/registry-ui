package providertypes

import (
	"fmt"
	"regexp"
	"strings"
)

type DocItemName string

const docItemNameMaxLength = 255

var docItemNameRe = regexp.MustCompile("^[a-zA-Z0-9 ._-]+$")

func (n DocItemName) Validate() error {
	if len(n) > docItemNameMaxLength || !docItemNameRe.MatchString(string(n)) {
		return fmt.Errorf("invalid doc item name: %s", n)
	}
	return nil
}

func (n DocItemName) Normalize() DocItemName {
	return DocItemName(strings.ToLower(string(n)))
}
