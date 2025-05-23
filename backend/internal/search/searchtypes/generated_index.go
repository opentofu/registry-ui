package searchtypes

import (
	"time"
)

type GeneratedIndexItemType string

const (
	GeneratedIndexItemHeader GeneratedIndexItemType = "header"
	GeneratedIndexItemAdd    GeneratedIndexItemType = "add"
	GeneratedIndexItemDelete GeneratedIndexItemType = "delete"
)

type GeneratedIndexHeader struct {
	LastUpdated time.Time `json:"last_updated"`
}

type ItemDeletion struct {
	ID        IndexID   `json:"id"`
	DeletedAt time.Time `json:"deleted_at"`
}

type GeneratedIndexItem struct {
	Type     GeneratedIndexItemType `json:"type"`
	Header   *GeneratedIndexHeader  `json:"header,omitempty"`
	Addition *IndexItem             `json:"addition,omitempty"`
	Deletion *ItemDeletion          `json:"deletion,omitempty"`
}
