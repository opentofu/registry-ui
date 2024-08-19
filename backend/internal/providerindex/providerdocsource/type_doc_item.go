package providerdocsource

import (
	"context"
	"fmt"

	"github.com/opentofu/libregistry/types/provider"
	"github.com/opentofu/registry-ui/internal/providerindex/providerindexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

type docItem struct {
	Name string `json:"name"` // The name of the resource, data source, or function

	Title       string `yaml:"page_title" json:"title"`        // The page title taken from the frontmatter
	Subcategory string `yaml:"subcategory" json:"subcategory"` // The subcategory of the resource, data source, or function, taken from the frontmatter
	Description string `yaml:"description" json:"description"` // The description of the resource.

	EditLink string `json:"edit_link"`

	contents []byte
}

func (d docItem) Store(
	ctx context.Context,
	addr provider.Addr,
	version providertypes.ProviderVersionDescriptor,
	storage providerindexstorage.API,
	language providertypes.CDKTFLanguage,
	itemKind providertypes.DocItemKind,
	itemName providertypes.DocItemName,
) error {
	// TODO validate addr
	if err := version.ID.Validate(); err != nil {
		return err
	}
	if err := itemKind.Validate(); err != nil {
		return err
	}
	if err := itemName.Validate(); err != nil {
		return err
	}

	if language == "" {
		if err := storage.StoreProviderDocItem(
			ctx,
			addr,
			version.ID,
			itemKind,
			itemName,
			d.contents,
		); err != nil {
			return fmt.Errorf("failed to store documentation item %s (%w)", itemName, err)
		}
		return nil
	}
	if err := language.Validate(); err != nil {
		return err
	}
	if err := storage.StoreProviderCDKTFDocItem(
		ctx,
		addr,
		version.ID,
		language,
		itemKind,
		itemName,
		d.contents,
	); err != nil {
		return fmt.Errorf("failed to store CDKTF documentation item %s for language %s (%w)", itemName, language, err)
	}
	return nil
}

func (d docItem) ToProviderTypes(_ context.Context) providertypes.ProviderDocItem {
	title := d.Title
	if title == "" {
		title = d.Name
	}
	return providertypes.ProviderDocItem{
		Name:        providertypes.DocItemName(d.Name),
		Title:       title,
		Subcategory: d.Subcategory,
		Description: d.Description,
		EditLink:    d.EditLink,
	}
}

func (d docItem) GetName(_ context.Context) (string, error) {
	return d.Name, nil
}

func (d docItem) GetTitle(_ context.Context) (string, error) {
	if d.Title != "" {
		return d.Title, nil
	}
	return d.Name, nil
}

func (d docItem) GetSubcategory(_ context.Context) (string, error) {
	return d.Subcategory, nil
}

func (d docItem) GetDescription(_ context.Context) (string, error) {
	return d.Description, nil
}

func (d docItem) GetContents(_ context.Context) ([]byte, error) {
	return d.contents, nil
}
