package provider

import "time"

// IndexResponse represents the response from provider indexing operations
type IndexResponse struct {
	Namespace      string    `json:"namespace"`
	Name           string    `json:"name"`
	Version        string    `json:"version"`
	ProcessedAt    time.Time `json:"processed_at"`
	DocumentsCount int       `json:"documents_count"`
	LicensesCount  int       `json:"licenses_count"`
	Success        bool      `json:"success"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

// ProviderVersion represents a specific version of a provider with metadata
type ProviderVersion struct {
	Namespace   string            `json:"namespace"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Source      string            `json:"source"`
	Tag         string            `json:"tag"`
	PublishedAt *time.Time        `json:"published_at"`
	Downloads   int64             `json:"downloads"`
	Docs        map[string]string `json:"docs"`
	Licenses    []string          `json:"licenses"`
}

