package main

import (
	"database/sql"
	"time"
)

type SearchIndexItem struct {
	Type     string `json:"type"`
	Addition struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		Addr        string `json:"addr"`
		Version     string `json:"version"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Link        struct {
			// Provider data
			ID        string `json:"id"`
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Version   string `json:"version"`

			// Module data
			Submodule    string `json:"submodule"`
			TargetSystem string `json:"target_system"`
		} `json:"link"`
		ParentID    string    `json:"parent_id"`
		LastUpdated time.Time `json:"last_updated"`
		Popularity  int       `json:"popularity"`
		Warnings    int       `json:"warnings"`
	} `json:"addition"`

	Deletion struct {
		ID        string    `json:"id"`
		DeletedAt time.Time `json:"deleted_at"`
	} `json:"deletion"`
}

type SearchHeader struct {
	ItemType string `json:"type"`
	Header   struct {
		LastUpdated time.Time `json:"last_updated"`
	}
}

type ImportJob struct {
	ID          int
	CreatedAt   sql.NullTime
	CompletedAt sql.NullTime
	Successful  bool
}
