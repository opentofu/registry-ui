package providertypes

import (
	"fmt"
)

type DocItemKind string

const (
	DocItemKindRoot       DocItemKind = ""
	DocItemKindResource   DocItemKind = "resource"
	DocItemKindDataSource DocItemKind = "datasource"
	DocItemKindFunction   DocItemKind = "function"
	DocItemKindGuide      DocItemKind = "guide"
)

func (k DocItemKind) Validate() error {
	switch k {
	case DocItemKindRoot:
	case DocItemKindResource:
	case DocItemKindDataSource:
	case DocItemKindFunction:
	case DocItemKindGuide:
	default:
		return fmt.Errorf("invalid doc item kind: %s", k)
	}
	return nil
}
