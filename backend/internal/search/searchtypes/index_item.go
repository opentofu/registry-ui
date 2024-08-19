package searchtypes

import (
	"fmt"
)

type IndexItem struct {
	ID            IndexID           `json:"id"`
	Type          IndexType         `json:"type"`
	Addr          string            `json:"addr"`
	Version       string            `json:"version"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	LinkVariables map[string]string `json:"link"`
	ParentID      IndexID           `json:"parent_id"`
}

func (i IndexItem) Validate() error {
	if err := i.ID.Validate(); err != nil {
		return fmt.Errorf("invalid index item ID: %s (%w)", i.ID, err)
	}
	if err := i.Type.Validate(); err != nil {
		return fmt.Errorf("invalid index item type: %s (%w)", i.Type, err)
	}
	if i.Title == "" {
		return fmt.Errorf("empty index item title")
	}
	if len(i.LinkVariables) == 0 {
		return fmt.Errorf("no link variables")
	}
	if i.ParentID != "" {
		if err := i.ParentID.Validate(); err != nil {
			return fmt.Errorf("invalid parent ID: %s (%w)", i.ParentID, err)
		}
	}
	if i.Addr == "" {
		return fmt.Errorf("the addr field cannot be empty")
	}
	if i.Version == "" {
		return fmt.Errorf("the version field cannot be empty")
	}
	return nil
}
