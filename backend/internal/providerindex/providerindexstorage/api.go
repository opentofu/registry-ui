// Package providerindexstorage holds the API to access the low-level storage API.
package providerindexstorage

import (
	"context"

	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

type API interface {
	// GetProviderList returns a provider list from storage. If there is no provider list, it will return a ready-to-use
	// empty provider list alongside a *ProviderListNotFoundError.
	//
	// swagger:operation GET /registry/docs/providers/index.json Providers GetProviderList
	// ---
	// produces:
	// - application/json
	// responses:
	//   '200':
	//     description: A list of all providers.
	//     schema:
	//       '$ref': '#/definitions/ProviderList'
	GetProviderList(ctx context.Context) (providertypes.ProviderList, error)
	StoreProviderList(ctx context.Context, providerList providertypes.ProviderList) error

	// GetProvider returns a list of all versions of a provider.
	//
	// swagger:operation GET /registry/docs/providers/{namespace}/{name}/index.json Providers GetProvider
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: namespace
	//   in: path
	//   required: true
	//   description: Namespace of the provider, all lower case.
	//   type: string
	// - name: name
	//   in: path
	//   required: true
	//   description: Name of the provider, all lower case.
	//   type: string
	// responses:
	//   '200':
	//     description: A list of all versions of a provider with metadata.
	//     schema:
	//       '$ref': '#/definitions/Provider'
	GetProvider(ctx context.Context, providerAddr providertypes.ProviderAddr) (providertypes.Provider, error)
	StoreProvider(ctx context.Context, provider providertypes.Provider) error
	DeleteProvider(ctx context.Context, providerAddr providertypes.ProviderAddr) error

	// GetProviderVersion returns the details of one specific provider version.
	//
	// swagger:operation GET /registry/docs/providers/{namespace}/{name}/{version}/index.json Providers GetProviderVersion
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: namespace
	//   in: path
	//   required: true
	//   description: Namespace of the provider, all lower case.
	//   type: string
	// - name: name
	//   in: path
	//   required: true
	//   description: Name of the provider, all lower case.
	//   type: string
	// - name: version
	//   in: path
	//   required: true
	//   description: Version number of the provider with the "v" prefix.
	//   type: string
	// responses:
	//   '200':
	//     description: The details of a specific provider version.
	//     schema:
	//       '$ref': '#/definitions/ProviderVersion'
	GetProviderVersion(ctx context.Context, providerAddr providertypes.ProviderAddr, version string) (providertypes.ProviderVersion, error)
	StoreProviderVersion(ctx context.Context, providerAddr providertypes.ProviderAddr, providerVersion providertypes.ProviderVersion) error
	DeleteProviderVersion(ctx context.Context, providerAddr providertypes.ProviderAddr, version string) error

	// GetProviderDoc returns a root provider document if it exists.
	//
	// swagger:operation GET /registry/docs/providers/{namespace}/{name}/{version}/index.md Providers GetProviderDoc
	// ---
	// parameters:
	// - name: namespace
	//   in: path
	//   required: true
	//   description: Namespace of the provider, all lower case.
	//   type: string
	// - name: name
	//   in: path
	//   required: true
	//   description: Name of the provider, all lower case.
	//   type: string
	// - name: version
	//   in: path
	//   required: true
	//   description: Version number of the provider with the "v" prefix.
	//   type: string
	// produces:
	// - text/markdown
	// responses:
	//   '200':
	//     description: The contents of the document.
	//     schema:
	//       type: file
	GetProviderDoc(ctx context.Context, providerAddr providertypes.ProviderAddr, version string) ([]byte, error)
	StoreProviderDoc(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, data []byte) error

	// GetProviderDocItem returns a provider document.
	//
	// swagger:operation GET /registry/docs/providers/{namespace}/{name}/{version}/{kind}s/{document}.md Providers GetProviderDocItem
	// ---
	// parameters:
	// - name: namespace
	//   in: path
	//   required: true
	//   description: Namespace of the provider, all lower case.
	//   type: string
	// - name: name
	//   in: path
	//   required: true
	//   description: Name of the provider, all lower case.
	//   type: string
	// - name: version
	//   in: path
	//   required: true
	//   description: Version number of the provider with the "v" prefix.
	//   type: string
	// - name: kind
	//   in: path
	//   required: true
	//   description: The kind of document to fetch.
	//   type: string
	//   enum: ["resource","datasource","function","guide"]
	// - name: document
	//   in: path
	//   required: true
	//   description: The name of the document without the .md suffix.
	//   type: string
	// produces:
	// - text/markdown
	// responses:
	//   '200':
	//     description: The contents of the document.
	//     schema:
	//       type: file
	GetProviderDocItem(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, kind providertypes.DocItemKind, name providertypes.DocItemName) ([]byte, error)
	StoreProviderDocItem(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, kind providertypes.DocItemKind, name providertypes.DocItemName, data []byte) error

	// GetProviderCDKTFDoc returns a CDKTF root document.
	//
	// swagger:operation GET /registry/docs/providers/{namespace}/{name}/{version}/cdktf/{language}/index.md Providers GetProviderCDKTFDoc
	// ---
	// parameters:
	// - name: namespace
	//   in: path
	//   required: true
	//   description: Namespace of the provider, all lower case.
	//   type: string
	// - name: name
	//   in: path
	//   required: true
	//   description: Name of the provider, all lower case.
	//   type: string
	// - name: version
	//   in: path
	//   required: true
	//   description: Version number of the provider with the "v" prefix.
	//   type: string
	// - name: language
	//   in: path
	//   required: true
	//   description: The CDKTF language to fetch.
	//   type: string
	//   enum: ["typescript","python","go","csharp","java"]
	// produces:
	// - text/markdown
	// responses:
	//   '200':
	//     description: The contents of the document.
	//     schema:
	//       type: file
	GetProviderCDKTFDoc(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, language providertypes.CDKTFLanguage) ([]byte, error)
	StoreProviderCDKTFDoc(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, language providertypes.CDKTFLanguage, data []byte) error

	// GetProviderCDKTFDocItem returns a CDKTF document item.
	//
	// swagger:operation GET /registry/docs/providers/{namespace}/{name}/{version}/cdktf/{language}/{kind}s/{document}.md Providers GetProviderCDKTFDocItem
	// ---
	// parameters:
	// - name: namespace
	//   in: path
	//   required: true
	//   description: Namespace of the provider, all lower case.
	//   type: string
	// - name: name
	//   in: path
	//   required: true
	//   description: Name of the provider, all lower case.
	//   type: string
	// - name: version
	//   in: path
	//   required: true
	//   description: Version number of the provider with the "v" prefix.
	//   type: string
	// - name: language
	//   in: path
	//   required: true
	//   description: The CDKTF language to fetch.
	//   type: string
	//   enum: ["typescript","python","go","csharp","java"]
	// - name: kind
	//   in: path
	//   required: true
	//   description: The kind of document to fetch.
	//   type: string
	//   enum: ["resource","datasource","function","guide"]
	// - name: document
	//   in: path
	//   required: true
	//   description: The name of the document without the .md suffix.
	//   type: string
	// produces:
	// - text/markdown
	// responses:
	//   '200':
	//     description: The contents of the document.
	//     schema:
	//       type: file
	GetProviderCDKTFDocItem(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, language providertypes.CDKTFLanguage, kind providertypes.DocItemKind, name providertypes.DocItemName) ([]byte, error)
	StoreProviderCDKTFDocItem(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, language providertypes.CDKTFLanguage, kind providertypes.DocItemKind, name providertypes.DocItemName, data []byte) error
}

func New(indexStorageAPI indexstorage.API) (API, error) {
	return &storage{indexStorageAPI: indexStorageAPI}, nil
}

const cdktfDirName = "cdktf"

type storage struct {
	indexStorageAPI indexstorage.API
}
