package searchtypes

import (
	"fmt"
)

type IndexID string

func (i IndexID) Validate() error {
	if i == "" {
		return fmt.Errorf("the ID must not be empty")
	}
	return nil
}
