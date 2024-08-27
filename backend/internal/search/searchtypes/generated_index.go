package searchtypes

import (
	"time"
)

// enum: ["header","add","delete"]
type GeneratedIndexItemType string

const (
	GeneratedIndexItemHeader GeneratedIndexItemType = "header"
	GeneratedIndexItemAdd    GeneratedIndexItemType = "add"
	GeneratedIndexItemDelete GeneratedIndexItemType = "delete"
)

// swagger:model GeneratedIndexHeader
type GeneratedIndexHeader struct {
	LastUpdated time.Time `json:"last_updated"`
}

// swagger:model
type ItemDeletion struct {
	// required: true
	ID IndexID `json:"id"`
	// required: true
	DeletedAt time.Time `json:"deleted_at"`
}

// swagger:model
type GeneratedIndexItem struct {
	// required: true
	Type GeneratedIndexItemType `json:"type"`
	// required: false
	Header *GeneratedIndexHeader `json:"header,omitempty"`
	// required: false
	Addition *IndexItem `json:"addition,omitempty"`
	// required: false
	Deletion *ItemDeletion `json:"deletion,omitempty"`
}
