// Package server OpenTofu Registry Docs API
//
// @title OpenTofu Registry Docs API
// @description The API to fetch documentation index and documentation files from the OpenTofu registry.
// @version 1.0.0-beta
// @license.name MPL-2.0
// @servers.url https://api.opentofu.org
// @servers.description OpenTofu Registry API server
package server

import (
	"context"
	_ "embed"

	"github.com/opentofu/registry-ui/internal/indexstorage"
)

//go:embed openapi.yml
var openapiYaml []byte

//go:embed swagger.yml
var swaggerYaml []byte

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
	if err := w.storage.WriteFile(ctx, "swagger.yml", openapiYaml); err != nil {
		return err
	}
	if err := w.storage.WriteFile(ctx, "openapi.yml", openapiYaml); err != nil {
		return err
	}
	if err := w.storage.WriteFile(ctx, "index.html", indexHTML); err != nil {
		return err
	}
	return nil
}
