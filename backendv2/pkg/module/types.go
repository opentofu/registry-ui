package module

import "time"

// IndexResponse represents the response from module indexing operations
type IndexResponse struct {
	Namespace      string    `json:"namespace"`
	Name           string    `json:"name"`
	Target         string    `json:"target"`
	Version        string    `json:"version"`
	ProcessedAt    time.Time `json:"processed_at"`
	DocumentsCount int       `json:"documents_count"`
	LicensesCount  int       `json:"licenses_count"`
	Success        bool      `json:"success"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

// ModuleVersion represents a specific version of a module with metadata
type ModuleVersion struct {
	Namespace   string            `json:"namespace"`
	Name        string            `json:"name"`
	Target      string            `json:"target"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Source      string            `json:"source"`
	Tag         string            `json:"tag"`
	PublishedAt *time.Time        `json:"published_at"`
	Downloads   int64             `json:"downloads"`
	Variables   map[string]string `json:"variables"`
	Outputs     map[string]string `json:"outputs"`
	Licenses    []string          `json:"licenses"`
}
