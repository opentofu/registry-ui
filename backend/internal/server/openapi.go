// Package server OpenTofu Registry Docs API
//
// The API to fetch documentation index and documentation files from the OpenTofu registry.
//
//	Version: 1.0.0-beta
//	License: MPL-2.0
//	Host: api.opentofu.org
//	Schemes: https
//
// swagger:meta
package server

import (
	"context"
	_ "embed"

	"github.com/opentofu/registry-ui/internal/indexstorage"
)

//go:embed openapi.yml
var openapiYaml []byte

//go:embed index.html
var indexHTML []byte

type OpenAPIWriter interface {
	Write(ctx context.Context) error
}

func NewWriter(storage indexstorage.API) (OpenAPIWriter, error) {
	return &writer{
		storage: storage,
	}, nil
}

type writer struct {
	storage indexstorage.API
}

func (w writer) Write(ctx context.Context) error {
	if err := w.storage.WriteFile(ctx, "openapi.yml", openapiYaml); err != nil {
		return err
	}
	if err := w.storage.WriteFile(ctx, "index.html", indexHTML); err != nil {
		return err
	}
	return nil
}
