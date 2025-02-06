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
	"bytes"
	"context"
	_ "embed"
	"io"
	"net/http"

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

	// Pull Redoc script from Redocly CDN
	redocFileName := "redoc.standalone.js"
	redocCDNURL := "https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"

	resp, err := http.Get(redocCDNURL)
	defer resp.Body.Close()

	buf := bytes.NewBuffer(nil)

	if _, err = io.Copy(buf, resp.Body); err != nil {
		return err
	}
	if err = w.storage.WriteFile(ctx, indexstorage.Path(redocFileName), buf.Bytes()); err != nil {
		return err
	}

	return nil
}
