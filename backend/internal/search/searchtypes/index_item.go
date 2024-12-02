package searchtypes

import (
	"fmt"
	"time"
)

type IndexItem struct {
	// The ID is used to ensure that we are only storing one item with a specific ID across all versions.
	// This should not be consumed by the search index.
	ID            IndexID           `json:"id"`
	Type          IndexType         `json:"type"`
	Addr          string            `json:"addr"`
	Version       string            `json:"version"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	LinkVariables map[string]string `json:"link"`
	ParentID      IndexID           `json:"parent_id"`
	LastUpdated   time.Time         `json:"last_updated"`
	Popularity    int               `json:"popularity"`
	Warnings      int               `json:"warnings"`
}

func (i IndexItem) Equals(other IndexItem) bool {
	if i.ID != other.ID || i.Type != other.Type || i.Addr != other.Addr || i.Version != other.Version ||
		i.Title != other.Title || i.Description != other.Description || i.ParentID != other.ParentID || i.Popularity != other.Popularity {
		return false
	}
	if len(i.LinkVariables) != len(other.LinkVariables) {
		return false
	}
	for k, v := range i.LinkVariables {
		if other.LinkVariables[k] != v {
			return false
		}
	}
	// We ignore LastUpdated since we want to compare the contents.
	return true
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
	if i.Popularity < 0 {
		return fmt.Errorf("the popularity cannot be negative")
	}
	return nil
}
