package searchtypes

import (
	"fmt"
)

type IndexType string

const (
	IndexTypeProvider           IndexType = "provider"
	IndexTypeProviderResource   IndexType = "provider/resource"
	IndexTypeProviderDatasource IndexType = "provider/datasource"
	IndexTypeProviderFunction   IndexType = "provider/function"
	IndexTypeModule             IndexType = "module"
	IndexTypeModuleSubmodule    IndexType = "module/submodule"
)

func (i IndexType) Validate() error {
	switch i {
	case IndexTypeProvider:
	case IndexTypeProviderResource:
	case IndexTypeProviderDatasource:
	case IndexTypeProviderFunction:
	case IndexTypeModule:
	case IndexTypeModuleSubmodule:
	default:
		return fmt.Errorf("invalid index type: %s", i)
	}
	return nil
}
