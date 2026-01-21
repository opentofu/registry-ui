package moduleindex

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/libregistry/metadata"
	"github.com/opentofu/libregistry/types/module"
	"github.com/opentofu/libregistry/vcs"
	"golang.org/x/sync/errgroup"

	"github.com/opentofu/registry-ui/internal/blocklist"

	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/license"
	"github.com/opentofu/registry-ui/internal/license/vcslinkfetcher"
	"github.com/opentofu/registry-ui/internal/moduleindex/moduleschema"
	"github.com/opentofu/registry-ui/internal/search"
)

//go:embed err_incompatible_license.md
var errorMessageIncompatibleLicense []byte

//go:embed err_no_readme.md
var errorNoReadme []byte

//go:embed err_blocked.md.tpl
var errorMessageBlocked []byte

var errorMessageBlockedTemplate = template.Must(template.New("").Parse(string(errorMessageBlocked)))

const indexPrefix = "modules"

// Generator is a tool to generate all index files for modules.
type Generator interface {
	// Generate generates all module index files incrementally and removes items no longer in the registry.
	Generate(ctx context.Context, opts ...Opts) error
	// GenerateNamespacePrefix generates module index files incrementally for all namespaces matching a given
	// prefix and removes modules from the index that no longer exist.
	GenerateNamespacePrefix(ctx context.Context, namespacePrefix string, opts ...Opts) error
	// GenerateNamespace generates module index files incrementally for one namespace and removes items no longer in the
	// registry.
	GenerateNamespace(ctx context.Context, namespace string, opts ...Opts) error
	// GenerateNamespaceAndName generates module index files incrementally for one namespace and removes items no longer in the
	// registry.
	GenerateNamespaceAndName(ctx context.Context, namespace string, name string, opts ...Opts) error
	// GenerateSingleModule generates module index files for a single module only.
	GenerateSingleModule(ctx context.Context, addr module.Addr, opts ...Opts) error
}

type GenerateConfig struct {
	Force               ForceRegenerate
	ForceRepoDataUpdate bool
}

type noForce struct{}

func (n noForce) MustRegenerateModule(_ context.Context, _ module.Addr) bool {
	return false
}

func (c *GenerateConfig) applyDefaults() error {
	if c.Force == nil {
		c.Force = &noForce{}
	}
	return nil
}

type Opts func(ctx context.Context, generateConfig *GenerateConfig) error

func WithForce(force ForceRegenerate) Opts {
	return func(_ context.Context, generateConfig *GenerateConfig) error {
		generateConfig.Force = force
		return nil
	}
}

func WithForceRepoDataUpdate(force bool) Opts {
	return func(_ context.Context, generateConfig *GenerateConfig) error {
		generateConfig.ForceRepoDataUpdate = force
		return nil
	}
}

func New(log logger.Logger, metadataAPI metadata.API, vcsClient vcs.Client, licenseDetector license.Detector, storage indexstorage.API, moduleSchemaExtractor moduleschema.Extractor, searchAPI search.API, blocklist blocklist.BlockList) Generator {
	return &generator{
		log:                   log.WithName("Module indexer"),
		metadataAPI:           metadataAPI,
		vcsClient:             vcsClient,
		licenseDetector:       licenseDetector,
		storage:               storage,
		moduleSchemaExtractor: moduleSchemaExtractor,
		search:                moduleSearch{searchAPI},
		blocklist:             blocklist,
	}
}

type generator struct {
	licenseDetector       license.Detector
	vcsClient             vcs.Client
	metadataAPI           metadata.API
	moduleSchemaExtractor moduleschema.Extractor
	storage               indexstorage.API
	log                   logger.Logger
	search                moduleSearch
	blocklist             blocklist.BlockList
}

func (g generator) GenerateSingleModule(ctx context.Context, addr module.Addr, opts ...Opts) error {
	addr = addr.Normalize()
	return g.generate(ctx, []module.Addr{addr}, func(moduleAddr ModuleAddr) bool {
		return !(addr.Equals(moduleAddr.Addr))
	}, opts)
}

func (g generator) GenerateNamespaceAndName(ctx context.Context, namespace string, name string, opts ...Opts) error {
	namespace = module.NormalizeNamespace(namespace)
	name = module.NormalizeName(name)
	moduleList, err := g.metadataAPI.ListModulesByNamespaceAndName(ctx, namespace, name)
	if err != nil {
		return err
	}
	return g.generate(ctx, moduleList, func(moduleAddr ModuleAddr) bool {
		return !(moduleAddr.Namespace == namespace && moduleAddr.Name == name)
	}, opts)
}

func (g generator) GenerateNamespacePrefix(ctx context.Context, namespacePrefix string, opts ...Opts) error {
	namespacePrefix = module.NormalizeNamespace(namespacePrefix)
	g.log.Info(ctx, "Listing all modules...")
	moduleListFull, err := g.metadataAPI.ListModules(ctx)
	if err != nil {
		return err
	}

	var moduleList []module.Addr
	for _, module := range moduleListFull {
		if strings.HasPrefix(module.Namespace, namespacePrefix) {
			moduleList = append(moduleList, module)
		}
	}

	return g.generate(ctx, moduleList, func(moduleAddr ModuleAddr) bool {
		return !strings.HasPrefix(moduleAddr.Namespace, namespacePrefix)
	}, opts)
}

func (g generator) GenerateNamespace(ctx context.Context, namespace string, opts ...Opts) error {
	namespace = module.NormalizeNamespace(namespace)
	g.log.Info(ctx, "Listing all modules...")
	moduleList, err := g.metadataAPI.ListModulesByNamespace(ctx, namespace)
	if err != nil {
		return err
	}
	return g.generate(ctx, moduleList, func(moduleAddr ModuleAddr) bool {
		return !(moduleAddr.Namespace == namespace)
	}, opts)
}

func (g generator) Generate(ctx context.Context, opts ...Opts) error {
	g.log.Info(ctx, "Listing all modules...")
	moduleList, err := g.metadataAPI.ListModules(ctx)
	if err != nil {
		return err
	}
	return g.generate(ctx, moduleList, func(moduleAddr ModuleAddr) bool {
		return false
	}, opts)
}

func (g generator) generate(ctx context.Context, moduleList []module.Addr, blockRemoval func(moduleAddr ModuleAddr) bool, opts []Opts) error {
	cfg := GenerateConfig{}
	for _, opt := range opts {
		if err := opt(ctx, &cfg); err != nil {
			return err
		}
	}
	if err := cfg.applyDefaults(); err != nil {
		return err
	}

	indexPath := "index.json"
	modules := ModuleList{
		// Leave this a slice so the JSON marshalling doesn't include a null.
		Modules: []*Module{},
	}

	g.log.Info(ctx, "Reading module index file...")
	modulesData, err := g.storage.ReadFile(ctx, indexstorage.Path(indexPath))
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to open %s (%w)", indexPath, err)
		}
	} else {
		if err := json.Unmarshal(modulesData, &modules); err != nil {
			return fmt.Errorf("corrupt %s (%w)", indexPath, err)
		}
	}

	var modulesToAdd []*Module
	var modulesToRemove []ModuleAddr
	var modulesToRemoveForce []ModuleAddr
	var eg errgroup.Group
	eg.SetLimit(25)
	lock := &sync.Mutex{}
	for _, moduleAddr := range moduleList {
		moduleAddr := Addr(moduleAddr)
		if err := moduleAddr.Validate(); err != nil {
			g.log.Info(ctx, "Module %s has an invalid address, skipping...", moduleAddr.String())
			continue
		}
		eg.Go(func() error {
			forceModule := cfg.Force.MustRegenerateModule(ctx, moduleAddr.Addr)

			blocked, blockedReason := g.blocklist.IsModuleBlocked(moduleAddr.Addr)

			moduleIndexPath := path.Join(moduleAddr.Namespace, moduleAddr.Name, moduleAddr.TargetSystem, "index.json")
			// We are fetching the module entry from the megaindex and storing it
			// further down below as a separate index file so the frontend has an easier time
			// fetching it.
			entry := modules.GetModule(moduleAddr.Addr)
			// originalEntry serves the purpose of being an original copy to compare to
			// so we don't write the index if it hasn't actually been modified to save costs.
			var originalEntry *Module
			needsAdd := false
			if entry == nil {
				entry = &Module{
					Addr:          moduleAddr,
					Description:   "",
					Versions:      nil,
					IsBlocked:     blocked,
					BlockedReason: blockedReason,
				}
				needsAdd = true
			} else {
				originalEntry = entry.DeepCopy()
			}

			if entry.IsBlocked != blocked {
				// If the blocked status has changed, we need to reindex the entire module to make sure the content
				// is gone or re-added, depending on the flag.
				forceModule = true
				g.log.Info(ctx, "Module %s changed blocked status, reindexing all versions...", moduleAddr)
			}
			entry.IsBlocked = blocked
			entry.BlockedReason = blockedReason

			g.log.Info(ctx, "Getting module metadata for %s...", moduleAddr)
			moduleMetadata, err := g.metadataAPI.GetModule(ctx, moduleAddr.Addr)
			if err != nil {
				return fmt.Errorf("failed to fetch metadata for module %s (%w)", moduleAddr, err)
			}

			var versionsToAdd []ModuleVersionDescriptor
			var versionsToUpdate []ModuleVersionDescriptor
			var versionsToRemove module.VersionList
			metadataVersions := moduleMetadata.Versions
			metadataVersions.Sort()
			// Make sure we index in the reverse order so the search index gets updated correctly.
			slices.Reverse(metadataVersions)
			// TODO when an older version is re-indexed, the search entry overwrites the newer version. However,
			//      currently this is an acceptable tradeoff as it should not normally happen.

			repoInfoFetched := false

			if cfg.ForceRepoDataUpdate {
				g.fetchRepoInfo(ctx, entry)
				repoInfoFetched = true
			}

			for _, ver := range metadataVersions {
				if err := ver.Validate(); err != nil {
					g.log.Warn(ctx, "Module %s version %s has an invalid version number, skipping...", moduleAddr.String(), ver.Version)
					versionsToRemove = append(versionsToRemove, ver)
					continue
				}
				vcsVer := ver.Version.ToVCSVersion()
				ver = ver.Normalize()
				hasVersion := entry.HasVersion(ver.Version)
				if hasVersion && !forceModule {
					g.log.Info(ctx, "The index already has version %s for module %s, skipping...", ver.Version, moduleAddr.String())
					continue
				}

				if !repoInfoFetched {
					// Make sure to fetch the description for the search index:
					g.fetchRepoInfo(ctx, entry)
					repoInfoFetched = true
				}

				publicationTime := time.Time{}
				vcsVersion, err := g.vcsClient.GetTagVersion(ctx, moduleAddr.ToRepositoryAddr(), vcsVer)
				if err != nil {
					var versionNotFound *vcs.VersionNotFoundError
					if errors.As(err, &versionNotFound) {
						g.log.Warn(ctx, "Module %s version %s not found in VCS system, skipping... (%v)", moduleAddr.String(), ver.Version, err)
						continue
					}
					g.log.Warn(ctx, "Cannot determine publication time for module %s version %s (%v)", moduleAddr.String(), ver.Version, err)
				} else {
					publicationTime = vcsVersion.Created
				}
				modVersion := ModuleVersionDescriptor{
					ID:        ver.Version,
					Published: publicationTime,
				}
				if err := g.generateModuleVersion(ctx, moduleAddr, *entry, modVersion, vcsVer); err != nil {
					var repoNotFound *vcs.RepositoryNotFoundError
					if errors.As(err, &repoNotFound) {
						g.log.Info(ctx, "The repository for the module %s has been removed from the VCS system, queueing removal from index.", moduleAddr.String())
						lock.Lock()
						modulesToRemove = append(modulesToRemove, moduleAddr)
						lock.Unlock()
						return nil
					}
					var versionNotFound *vcs.VersionNotFoundError
					if !errors.As(err, &versionNotFound) {
						g.log.Error(ctx, "Module indexing for %s version %s failed (%v)", moduleAddr.String(), ver.Version, err)
						return fmt.Errorf("failed to generate module %s version %s (%w)", moduleAddr.String(), ver.Version, err)
					}
					g.log.Info(ctx, "The version %s for the module %s has been removed from the VCS system, queueing removal from index.", ver.Version, moduleAddr.String())
					versionsToRemove = append(versionsToRemove, ver)
					if err := g.search.removeModuleVersionFromSearchIndex(ctx, moduleAddr.Addr, ver.Version); err != nil {
						return fmt.Errorf("failed to remove module %s version %s from search index (%w)", moduleAddr, ver.Version, err)
					}
				} else {
					if hasVersion {
						versionsToUpdate = append(versionsToUpdate, modVersion)
						// TODO we currently don't remove submodules and examples that no longer exist. This should
						//      be addressed by diffing the existing and the new version.
					} else {
						versionsToAdd = append(versionsToAdd, modVersion)
					}
				}
			}
			entry.AddVersions(versionsToAdd...)
			entry.UpdateVersions(versionsToUpdate...)
			removedVersions := entry.RemoveVersions(versionsToRemove, moduleMetadata.Versions)
			for _, version := range removedVersions {
				if err := g.removeModuleVersion(ctx, moduleAddr, version); err != nil {
					return fmt.Errorf("cannot remove module data for %s version %s (%w)", moduleAddr, version, err)
				}
			}

			if len(entry.Versions) == 0 {
				g.log.Info(ctx, "Module %s has no versions, queueing for removal from index.", entry.Addr.String())
				lock.Lock()
				modulesToRemoveForce = append(modulesToRemoveForce, moduleAddr)
				lock.Unlock()
				if err := g.search.removeModuleFromSearchIndex(ctx, entry.Addr); err != nil {
					return fmt.Errorf("failed to remove module from search index (%w)", err)
				}
			} else {
				if needsAdd {
					lock.Lock()
					modulesToAdd = append(modulesToAdd, entry)
					lock.Unlock()
				}
				// Here we compare the module entry to its original copy to make sure
				// we are only writing this index if needed. This is needed because writes
				// on R2 cost money, whereas reads don't and updating all the provider and
				// module indexes on every run costs ~300$ per month.
				if originalEntry == nil || !originalEntry.Equals(entry) {
					versionListing, err := json.Marshal(entry)
					if err != nil {
						return fmt.Errorf("failed to marshal module index for %s (%w)", entry.Addr, err)
					}
					if err := g.storage.WriteFile(ctx, indexstorage.Path(moduleIndexPath), versionListing); err != nil {
						return fmt.Errorf("failed to write the module index for %s (%w)", entry.Addr, err)
					}
				}
			}

			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	modules.addModules(modulesToAdd...)
	removedModules := modules.removeModules(modulesToRemove, moduleList, blockRemoval, modulesToRemoveForce)
	for _, modAddr := range removedModules {
		if err := g.removeModule(ctx, modAddr); err != nil {
			return fmt.Errorf("cannot remove module %s (%w)", modAddr, err)
		}
	}
	for _, m := range modules.Modules {
		if err := m.Validate(); err != nil {
			// We are ignoring invalid version numbers here because the dataset contains them, but when a module is
			// refreshed, it should be excluded above.
			var invalidVersionNumber *module.InvalidVersionNumber
			if !errors.As(err, &invalidVersionNumber) {
				return fmt.Errorf("invalid module (%w)", err)
			}
		}
	}
	marshalled, err := json.Marshal(modules)
	if err != nil {
		return fmt.Errorf("failed to marshal %s (%w)", indexPath, err)
	}
	if err := g.storage.WriteFile(ctx, indexstorage.Path(indexPath), marshalled); err != nil {
		return fmt.Errorf("failed to write %s (%w)", indexPath, err)
	}
	return nil
}

func (g generator) removeModule(ctx context.Context, moduleAddr ModuleAddr) error {
	modPath := path.Join(moduleAddr.Namespace, moduleAddr.Name, moduleAddr.TargetSystem)
	return g.storage.RemoveAll(ctx, indexstorage.Path(modPath))
}

func (g generator) removeModuleVersion(ctx context.Context, moduleAddr ModuleAddr, version module.VersionNumber) error {
	indexPath := path.Join(moduleAddr.Namespace, moduleAddr.Name, moduleAddr.TargetSystem, string(version))
	if err := g.storage.RemoveAll(ctx, indexstorage.Path(indexPath)); err != nil {
		return err
	}
	return nil
}

func (g generator) generateModuleVersion(ctx context.Context, moduleAddr ModuleAddr, entry Module, ver ModuleVersionDescriptor, vcsVersion vcs.VersionNumber) error {
	g.log.Info(ctx, "Generating index artifacts for module %s version %s...", moduleAddr.String(), ver.ID)
	indexPath := path.Join(moduleAddr.Namespace, moduleAddr.Name, moduleAddr.TargetSystem, string(ver.ID), "index.json")
	result := ModuleVersion{
		ModuleVersionDescriptor: ver,
		Details: Details{
			BaseDetails: BaseDetails{
				Readme:      false,
				Variables:   map[string]Variable{},
				Outputs:     map[string]Output{},
				SchemaError: "",
			},
			Providers:    []ProviderDependency{},
			Dependencies: []ModuleDependency{},
			Resources:    []Resource{},
		},
		VCSRepository:       "",
		Licenses:            nil,
		IncompatibleLicense: false,
		Link:                "",
		Examples:            map[string]Example{},
		Submodules:          map[string]Submodule{},
	}
	g.log.Info(ctx, "Reading module index for %s version %s...", moduleAddr, ver.ID)
	contents, err := g.storage.ReadFile(ctx, indexstorage.Path(indexPath))
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read %s (%w)", indexPath, err)
		}
	} else {
		if err := json.Unmarshal(contents, &result); err != nil {
			return fmt.Errorf("module descriptor %s corrupt (%w)", indexPath, err)
		}
	}

	g.log.Info(ctx, "Checking out module %s version %s...", moduleAddr, ver.ID)
	workingCopy, err := g.vcsClient.Checkout(ctx, moduleAddr.ToRepositoryAddr(), vcsVersion)
	if err != nil {
		return fmt.Errorf("failed to check out %s version %s (%w)", moduleAddr.String(), ver.ID, err)
	}
	defer func() {
		if err := workingCopy.Close(); err != nil {
			g.log.Error(ctx, "Failed to close working copy for %s (%v)", moduleAddr, err)
		}
	}()

	g.log.Info(ctx, "Updating license for module %s version %s...", moduleAddr, ver.ID)
	if err := g.refreshLicense(ctx, moduleAddr, ver, &result, workingCopy); err != nil {
		return fmt.Errorf("failed to fetch licenses for %s version %s (%w)", moduleAddr, ver.ID, err)
	}
	licenseOK := result.Licenses.IsRedistributable()
	result.IncompatibleLicense = !licenseOK

	result.Link, err = workingCopy.Client().GetVersionBrowseURL(ctx, workingCopy.Repository(), workingCopy.Version())
	if err != nil {
		g.log.Warn(ctx, "Cannot determine browse URL for module repository %s version %s (%v)", workingCopy.Repository(), workingCopy.Version(), err)
	}

	g.log.Info(ctx, "Updating module details for %s version %s...", moduleAddr, ver.ID)
	if err := g.refreshModuleDetails(ctx, moduleAddr, ver, &result.Details, workingCopy, licenseOK, entry.IsBlocked, entry.BlockedReason, "", ""); err != nil {
		return fmt.Errorf("failed to extract module defaults for %s version %s (%w)", moduleAddr, ver.ID, err)
	}

	if err := g.extractSubmodules(ctx, moduleAddr, ver, &result, workingCopy, licenseOK, entry.IsBlocked, entry.BlockedReason); err != nil {
		return err
	}

	if err := g.extractExamples(ctx, moduleAddr, ver, &result, workingCopy, licenseOK, entry.IsBlocked, entry.BlockedReason); err != nil {
		return err
	}

	contents, err = json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal %s (%w)", indexPath, err)
	}
	if err := g.storage.WriteFile(ctx, indexstorage.Path(indexPath), contents); err != nil {
		return fmt.Errorf("failed to write %s (%w)", indexPath, err)
	}

	if err := g.search.indexModuleVersion(ctx, moduleAddr, entry, result); err != nil {
		return fmt.Errorf("failed to index module %s version %s for search (%w)", moduleAddr, ver.ID, err)
	}

	return nil
}

func (g generator) refreshLicense(ctx context.Context, moduleAddr ModuleAddr, moduleVersion ModuleVersionDescriptor, result *ModuleVersion, workingCopy vcs.WorkingCopy) error {
	var err error
	result.Licenses, err = g.licenseDetector.Detect(ctx, workingCopy, license.WithLinkFetcher(vcslinkfetcher.Fetcher(
		ctx,
		workingCopy.Repository(),
		workingCopy.Version(),
		g.vcsClient,
	)))
	return err
}

func (g generator) refreshModuleDetails(ctx context.Context, moduleAddr ModuleAddr, ver ModuleVersionDescriptor, d *Details, workingCopy vcs.WorkingCopy, licenseOK bool, blocked bool, blockedReason string, sourcePrefix string, dstPrefix string) error {
	var err error
	if d.Readme, d.EditLink, err = g.extractReadme(ctx, moduleAddr, ver, workingCopy, licenseOK, blocked, blockedReason, sourcePrefix, dstPrefix); err != nil {
		return err
	}

	rawDirectory, err := workingCopy.RawDirectory()
	if err != nil {
		// No raw directory support, skip the rest. (This is not a problem with GitHub, only with fakevcs.)
		// TODO fakevcs should support RawDirectory in libregistry.
		return nil
	}

	dir := path.Join(rawDirectory, sourcePrefix)
	if err := g.extractModuleSchema(ctx, dir, d, licenseOK, blocked); err != nil {
		return err
	}

	return nil
}

func (g generator) refreshExampleDetails(ctx context.Context, moduleAddr ModuleAddr, ver ModuleVersionDescriptor, e *Example, workingCopy vcs.WorkingCopy, licenseOK bool, blocked bool, blockedReason string, sourcePrefix string, dstPrefix string) error {
	var err error
	if e.Readme, e.EditLink, err = g.extractReadme(ctx, moduleAddr, ver, workingCopy, licenseOK, blocked, blockedReason, sourcePrefix, dstPrefix); err != nil {
		return err
	}

	rawDirectory, err := workingCopy.RawDirectory()
	if err != nil {
		// No raw directory support, skip the rest. (This is not a problem with GitHub, only with fakevcs.)
		// TODO fakevcs should support RawDirectory in libregistry.
		return nil
	}

	dir := path.Join(rawDirectory, sourcePrefix)
	if err := g.extractExampleSchema(ctx, dir, e, licenseOK, blocked); err != nil {
		return err
	}

	return nil
}

func (g generator) extractReadme(ctx context.Context, moduleAddr ModuleAddr, ver ModuleVersionDescriptor, workingCopy vcs.WorkingCopy, licenseOK bool, blocked bool, blockedReason string, sourcePrefix string, dstPrefix string) (bool, string, error) {
	hasReadme := false
	var readme []byte
	sourcePath := path.Join(sourcePrefix, "README.md")
	fh, err := workingCopy.Open(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			readme = errorNoReadme
		} else {
			return false, "", fmt.Errorf("failed to open README.md")
		}
	} else {
		switch {
		case blocked:
			wr := &bytes.Buffer{}
			if err := errorMessageBlockedTemplate.Execute(wr, blockedReason); err != nil {
				return false, "", fmt.Errorf("failed to create template (%w)", err)
			}
			readme = wr.Bytes()
		case !licenseOK:
			readme = errorMessageIncompatibleLicense
		default:
			readme, err = io.ReadAll(fh)
			if err != nil {
				return false, "", fmt.Errorf("failed to read README.md (%w)", err)
			}
		}
		hasReadme = true
		_ = fh.Close()
	}
	readmePath := path.Join(moduleAddr.Namespace, moduleAddr.Name, moduleAddr.TargetSystem, string(ver.ID), "README.md")
	if dstPrefix != "" {
		readmePath = path.Join(moduleAddr.Namespace, moduleAddr.Name, moduleAddr.TargetSystem, string(ver.ID), dstPrefix, "README.md")
	}
	if err := g.storage.WriteFile(ctx, indexstorage.Path(readmePath), readme); err != nil {
		return hasReadme, "", fmt.Errorf("failed to write README.md at %s (%w)", readmePath, err)
	}
	readmeViewURL := ""
	if hasReadme {
		readmeViewURL, err = workingCopy.Client().GetFileViewURL(ctx, workingCopy.Repository(), workingCopy.Version(), sourcePath)
		if err != nil {
			g.log.Warn(ctx, "Cannot determine edit link for %s (%v)", readmePath, err)
		}
	}
	return hasReadme, readmeViewURL, nil
}

func (g generator) extractSubmodules(ctx context.Context, addr ModuleAddr, ver ModuleVersionDescriptor, m *ModuleVersion, workingCopy vcs.WorkingCopy, licenseOK bool, blocked bool, blockedReason string) error {
	g.log.Info(ctx, "Extracting submodules for module %s version %s...", addr, ver.ID)
	// Note: we extract the fact that a submodule exists even if the license is not OK because we just index the fact
	// that the submodule exists. However, we do not index the contents of the submodule.

	const directoryPrefix = "modules"
	// We only want to handle situations where the file is a directory, so we check if the directory exists.
	// and if it's a file, we skip it.
	if fileInfo, err := workingCopy.Open(directoryPrefix); err == nil {
		defer fileInfo.Close()
		if stat, err := fileInfo.Stat(); err == nil && !stat.IsDir() {
			// it's not a directory, so we skip it.
			return nil
		}
	}
	entries, err := workingCopy.ReadDir(directoryPrefix)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "." || name == ".." {
			continue
		}
		submodule := Submodule{
			Details: Details{
				BaseDetails: BaseDetails{
					Readme:      false,
					Variables:   map[string]Variable{},
					Outputs:     map[string]Output{},
					SchemaError: "",
				},
				Providers:    []ProviderDependency{},
				Dependencies: []ModuleDependency{},
				Resources:    []Resource{},
			},
		}
		sourcePrefix := path.Join(directoryPrefix, name)
		dstPrefix, err := g.sanitizePath(directoryPrefix, name)
		if err != nil {
			// Invalid path, ignore.
			g.log.Warn(ctx, "Failed to index %s version %s submodule %s (%v)", addr, ver.ID, dstPrefix, err)
			continue
		}
		if err := g.refreshModuleDetails(ctx, addr, ver, &submodule.Details, workingCopy, licenseOK, blocked, blockedReason, sourcePrefix, dstPrefix); err != nil {
			return fmt.Errorf("failed to refresh details for submodule %s (%w)", sourcePrefix, err)
		}

		m.Submodules[name] = submodule
	}

	return nil
}

func (g generator) extractExamples(ctx context.Context, moduleAddr ModuleAddr, ver ModuleVersionDescriptor, m *ModuleVersion, workingCopy vcs.WorkingCopy, licenseOK bool, blocked bool, blockedReason string) error {
	g.log.Info(ctx, "Extracting examples for module %s version %s...", moduleAddr, ver.ID)
	// Note: we extract the fact that an example exists even if the license is not OK because we just index the fact
	// that the submodule exists. However, we do not index the contents of the submodule.
	const directoryPrefix = "examples"
	// We only want to handle situations where the file is a directory, so we check if the directory exists.
	// and if it's a file, we skip it.
	if fileInfo, err := workingCopy.Open(directoryPrefix); err == nil {
		defer fileInfo.Close()
		if stat, err := fileInfo.Stat(); err == nil && !stat.IsDir() {
			// it's not a directory, so we skip it.
			return nil
		}
	}
	entries, err := workingCopy.ReadDir(directoryPrefix)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			g.log.Info(ctx, "Skipping non-directory entry %s in examples for module %s version %s", entry.Name(), moduleAddr, ver.ID)
			continue
		}
		name := entry.Name()
		if name == "." || name == ".." {
			continue
		}
		example := Example{
			BaseDetails: BaseDetails{
				Readme:    false,
				Variables: map[string]Variable{},
				Outputs:   map[string]Output{},
			},
		}
		srcPrefix := path.Join(directoryPrefix, name)
		dstPrefix, err := g.sanitizePath(directoryPrefix, name)
		if err != nil {
			// Invalid path, don't include in the index
			g.log.Warn(ctx, "Failed to index %s version %s example %s (%v)", moduleAddr, ver.ID, dstPrefix, err)
			continue
		}
		if err := g.refreshExampleDetails(
			ctx,
			moduleAddr,
			ver,
			&example,
			workingCopy,
			licenseOK,
			blocked,
			blockedReason,
			srcPrefix,
			dstPrefix,
		); err != nil {
			return fmt.Errorf("failed to refresh details for example %s (%w)", srcPrefix, err)
		}

		m.Examples[name] = example
	}

	return nil
}

func (g generator) extractModuleSchema(ctx context.Context, directory string, d *Details, licenseOK bool, blocked bool) error {
	if !licenseOK || blocked {
		return nil
	}
	moduleSchema, err := g.moduleSchemaExtractor.Extract(ctx, directory)
	if err != nil {
		var extractionFailed *moduleschema.SchemaExtractionFailedError
		if errors.As(err, &extractionFailed) {
			// TODO add better errors when the tofu binary is not available vs. when the module has a problem.
			d.SchemaError = extractionFailed.OutputString()
			return nil
		}
		return err
	}

	rootModuleSchema := moduleSchema.RootModule

	g.extractModuleVariables(ctx, rootModuleSchema, &d.BaseDetails)
	g.extractModuleOutputs(rootModuleSchema, &d.BaseDetails)
	g.extractModuleDependencies(rootModuleSchema, d)
	g.extractProviderDependencies(moduleSchema, d)
	g.extractModuleResources(rootModuleSchema, d)
	return nil
}

func (g generator) extractExampleSchema(ctx context.Context, directory string, e *Example, licenseOK bool, blocked bool) error {
	if !licenseOK || blocked {
		return nil
	}
	moduleSchema, err := g.moduleSchemaExtractor.Extract(ctx, directory)
	if err != nil {
		var extractionFailed *moduleschema.SchemaExtractionFailedError
		if errors.As(err, &extractionFailed) {
			// TODO add better errors when the tofu binary is not available vs. when the module has a problem.
			e.SchemaError = extractionFailed.OutputString()
			return nil
		}
		return err
	}

	rootModuleSchema := moduleSchema.RootModule

	g.extractModuleVariables(ctx, rootModuleSchema, &e.BaseDetails)
	g.extractModuleOutputs(rootModuleSchema, &e.BaseDetails)

	return nil
}

func (g generator) extractModuleOutputs(moduleSchema moduleschema.ModuleSchema, d *BaseDetails) {
	for outputName, output := range moduleSchema.Outputs {
		if _, ok := d.Outputs[outputName]; ok {
			continue
		}
		d.Outputs[outputName] = Output{
			Description: output.Description,
			Sensitive:   output.Sensitive,
		}
	}
	for outputName := range d.Outputs {
		if _, ok := moduleSchema.Variables[outputName]; !ok {
			delete(d.Variables, outputName)
		}
	}
}

func (g generator) extractModuleVariables(ctx context.Context, moduleSchema moduleschema.ModuleSchema, d *BaseDetails) {
	for variableName, variable := range moduleSchema.Variables {
		if _, ok := d.Variables[variableName]; ok {
			continue
		}
		d.Variables[variableName] = Variable{
			Type:        variable.Type,
			Default:     variable.Default,
			Description: variable.Description,
			Sensitive:   variable.Sensitive,
			Required:    variable.Required,
		}
	}
	for variableName := range d.Variables {
		if _, ok := moduleSchema.Variables[variableName]; !ok {
			delete(d.Variables, variableName)
		}
	}
}

func (g generator) extractModuleDependencies(moduleSchema moduleschema.ModuleSchema, d *Details) {
	result := make([]ModuleDependency, len(moduleSchema.ModuleCalls))
	i := 0
	for moduleCallName, moduleCall := range moduleSchema.ModuleCalls {
		result[i] = ModuleDependency{
			Name:              moduleCallName,
			VersionConstraint: moduleCall.VersionConstraint,
			Source:            moduleCall.Source,
		}
		i++
	}
	d.Dependencies = result
}

func (g generator) extractProviderDependencies(schema moduleschema.Schema, d *Details) {
	result := make([]ProviderDependency, len(schema.ProviderConfig))
	i := 0
	for providerCallName, providerCall := range schema.ProviderConfig {
		result[i] = ProviderDependency{
			Name:              providerCallName,
			FullName:          providerCall.FullName,
			VersionConstraint: providerCall.VersionConstraint,
			Alias:             "",
		}
		i++
	}

	d.Providers = result
}

func (g generator) extractModuleResources(moduleSchema moduleschema.ModuleSchema, d *Details) {
	result := make([]Resource, len(moduleSchema.Resources))
	for i, resource := range moduleSchema.Resources {
		result[i] = Resource{
			Address: resource.Address,
			Type:    resource.Type,
			Name:    resource.Name,
		}
	}
	d.Resources = result
}

func (g generator) fetchRepoInfo(ctx context.Context, entry *Module) {
	repoInfo, err := g.vcsClient.GetRepositoryInfo(ctx, entry.Addr.ToRepositoryAddr())
	if err != nil {
		var repoNotFound *vcs.RepositoryNotFoundError
		if errors.As(err, &repoNotFound) {
			g.log.Warn(ctx, "Repository not found for module %s, skipping... (%v)", entry.Addr.String(), err)
			return
		}
		// We handle description errors as soft errors because they are purely presentational.
		g.log.Warn(ctx, "Cannot update repository description for module %s (%v)", entry.Addr.String(), err)
	}
	entry.Description = repoInfo.Description
	entry.Popularity = repoInfo.Popularity
	entry.ForkCount = repoInfo.ForkCount

	forkRepo := repoInfo.ForkOf
	if forkRepo == nil {
		return
	}
	link, err := g.vcsClient.GetRepositoryBrowseURL(ctx, *forkRepo)
	if err != nil {
		g.log.Warn(ctx, "Cannot determine repository browse URL for %s (%v)", forkRepo.String(), err)
		return
	}
	entry.ForkOfLink = link

	forkedAddr, err := module.AddrFromRepository(*forkRepo)
	if err != nil {
		g.log.Warn(ctx, "Cannot convert repository name %s to a module addr (%v)", forkRepo.String(), err)
		return
	}
	_, err = g.metadataAPI.GetModule(ctx, forkedAddr)
	if err != nil {
		return
	}
	entry.ForkOf = Addr(forkedAddr)

	upstreamRepoInfo, err := g.vcsClient.GetRepositoryInfo(ctx, *forkRepo)
	if err != nil {
		g.log.Warn(ctx, "Cannot fetch upstream repository info for %s (%v)", forkRepo.String(), err)
		return
	}
	entry.UpstreamPopularity = upstreamRepoInfo.Popularity
	entry.UpstreamForkCount = upstreamRepoInfo.ForkCount
}

func (g generator) sanitizePath(prefix string, name string) (string, error) {
	name = strings.ReplaceAll(name, " ", "")
	return path.Join(prefix, name), indexstorage.Path(name).Validate()
}
