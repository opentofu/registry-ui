package providertypes

import (
	"fmt"
)

type CDKTFLanguage string

const (
	CDKTFLanguagePython     CDKTFLanguage = "python"
	CDKTFLanguageTypescript CDKTFLanguage = "typescript"
	CDKTFLanguageCSharp     CDKTFLanguage = "csharp"
	CDKTFLanguageJava       CDKTFLanguage = "java"
	CDKTFLanguageGo         CDKTFLanguage = "go"
)

func (l CDKTFLanguage) Validate() error {
	switch l {
	case CDKTFLanguagePython:
	case CDKTFLanguageTypescript:
	case CDKTFLanguageCSharp:
	case CDKTFLanguageJava:
	case CDKTFLanguageGo:
	default:
		return fmt.Errorf("invalid CDKTF language: %s", l)
	}
	return nil
}
