package moduleindex

import (
	"context"
	"fmt"

	"github.com/opentofu/libregistry/types/module"

	"github.com/opentofu/registry-ui/internal/search"
	"github.com/opentofu/registry-ui/internal/search/searchtypes"
)

type moduleSearch struct {
	searchAPI search.API
}

func (m moduleSearch) indexModuleVersion(ctx context.Context, addr ModuleAddr, module Module, response ModuleVersion) error {
	if err := addr.Validate(); err != nil {
		return err
	}
	if err := response.ID.Validate(); err != nil {
		return err
	}

	versionItem := searchtypes.IndexItem{
		ID:          searchtypes.IndexID(indexPrefix + "/" + addr.String()),
		Type:        searchtypes.IndexTypeModule,
		Addr:        addr.String(),
		Version:     string(response.ID),
		Title:       addr.TargetSystem,
		Description: module.Description,
		LinkVariables: map[string]string{
			"namespace":     addr.Namespace,
			"name":          addr.Name,
			"target_system": addr.TargetSystem,
			"version":       string(response.ID),
		},
		ParentID:   "",
		Popularity: module.Popularity,
	}
	if err := m.searchAPI.AddItem(ctx, versionItem); err != nil {
		return err
	}

	for name, _ := range response.Submodules {
		submoduleItem := searchtypes.IndexItem{
			// We pick an ID without a version number so that the search index overwrites the submodule names.
			ID:          searchtypes.IndexID(indexPrefix + "/" + addr.String() + "/" + name),
			Type:        searchtypes.IndexTypeModuleSubmodule,
			Addr:        addr.String(),
			Version:     string(response.ID),
			Title:       name,
			Description: module.Description,
			LinkVariables: map[string]string{
				"namespace":     addr.Namespace,
				"name":          addr.Name,
				"target_system": addr.TargetSystem,
				"version":       string(response.ID),
				"submodule":     name,
			},
			ParentID:   versionItem.ID,
			Popularity: module.Popularity,
		}
		if err := m.searchAPI.AddItem(ctx, submoduleItem); err != nil {
			return fmt.Errorf("failed to add submodule %s item (%w)", name, err)
		}
	}
	return nil
}

func (m moduleSearch) removeModuleVersionFromSearchIndex(ctx context.Context, addr module.Addr, version module.VersionNumber) error {
	return m.searchAPI.RemoveVersionItems(ctx, searchtypes.IndexTypeModuleSubmodule, addr.String(), string(version))
}

func (m moduleSearch) removeModuleFromSearchIndex(ctx context.Context, addr ModuleAddr) error {
	return m.searchAPI.RemoveItem(ctx, searchtypes.IndexID(indexPrefix+"/"+addr.String()))
}
