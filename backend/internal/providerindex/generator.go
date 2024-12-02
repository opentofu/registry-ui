package providerindex

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/libregistry/metadata"
	"github.com/opentofu/libregistry/types/provider"
	"github.com/opentofu/libregistry/vcs"
	"github.com/opentofu/registry-ui/internal/blocklist"
	"golang.org/x/sync/errgroup"

	"github.com/opentofu/registry-ui/internal/providerindex/providerdocsource"
	"github.com/opentofu/registry-ui/internal/providerindex/providerindexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
	"github.com/opentofu/registry-ui/internal/search"

	"github.com/opentofu/registry-ui/internal/license"
)

// DocumentationGenerator is a tool to generate all index files for modules.
type DocumentationGenerator interface {
	// Generate generates all module index files incrementally and removes items no longer in the registry.
	Generate(ctx context.Context, opts ...Opts) error

	// GenerateNamespace generates provider index files incrementally for one namespace.
	GenerateNamespace(ctx context.Context, namespace string, opts ...Opts) error

	// GenerateNamespacePrefix generates provider index files incrementally for multiple namespaces matching the given
	// prefix.
	GenerateNamespacePrefix(ctx context.Context, namespacePrefix string, opts ...Opts) error

	// GenerateSingleProvider generates module index files for a single provider only.
	GenerateSingleProvider(ctx context.Context, addr provider.Addr, opts ...Opts) error
}

type GenerateConfig struct {
	Force               ForceRegenerate
	ForceRepoDataUpdate bool
}

type noForce struct {
}

func (n noForce) MustRegenerateProvider(_ context.Context, _ provider.Addr) bool {
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

func NewDocumentationGenerator(log logger.Logger, metadataAPI metadata.API, vcsClient vcs.Client, licenseDetector license.Detector, source providerdocsource.API, destination providerindexstorage.API, searchAPI search.API, blocklist blocklist.BlockList) DocumentationGenerator {
	return &documentationGenerator{
		log:             log.WithName("Provider indexer"),
		metadataAPI:     metadataAPI,
		vcsClient:       vcsClient,
		licenseDetector: licenseDetector,
		source:          source,
		destination:     destination,
		search: providerSearch{
			searchAPI: searchAPI,
		},
		blocklist: blocklist,
	}
}

type documentationGenerator struct {
	log             logger.Logger
	metadataAPI     metadata.API
	vcsClient       vcs.Client
	licenseDetector license.Detector
	search          providerSearch
	source          providerdocsource.API
	destination     providerindexstorage.API
	blocklist       blocklist.BlockList
}

func (d *documentationGenerator) GenerateSingleProvider(ctx context.Context, addr provider.Addr, opts ...Opts) error {
	addr = addr.Normalize()
	if err := d.scrape(ctx, []provider.Addr{addr}, opts); err != nil {
		return err
	}

	return nil
}

func (d *documentationGenerator) Generate(ctx context.Context, opts ...Opts) error {
	d.log.Info(ctx, "Listing all providers...")
	providerList, err := d.metadataAPI.ListProviders(ctx, true)
	if err != nil {
		return err
	}
	d.log.Info(ctx, "Loaded %d providers", len(providerList))

	err = d.scrape(ctx, providerList, opts)
	if err != nil {
		return err
	}

	return nil
}

func (d *documentationGenerator) GenerateNamespace(ctx context.Context, namespace string, opts ...Opts) error {
	d.log.Info(ctx, "Listing all providers in namespace %s...", namespace)
	providerList, err := d.metadataAPI.ListProvidersByNamespace(ctx, namespace, true)

	d.log.Info(ctx, "Loaded %d providers", len(providerList))

	err = d.scrape(ctx, providerList, opts)
	if err != nil {
		return err
	}

	return nil
}

func (d *documentationGenerator) GenerateNamespacePrefix(ctx context.Context, namespacePrefix string, opts ...Opts) error {
	d.log.Info(ctx, "Listing all providers with the namespace prefix %s...", namespacePrefix)
	providerListFull, err := d.metadataAPI.ListProviders(ctx, true)
	if err != nil {
		return err
	}
	var providerList []provider.Addr
	for _, providerAddr := range providerListFull {
		if strings.HasPrefix(providerAddr.Namespace, namespacePrefix) {
			providerList = append(providerList, providerAddr)
		}
	}
	d.log.Info(ctx, "Loaded %d providers", len(providerList))

	err = d.scrape(ctx, providerList, opts)
	if err != nil {
		return err
	}

	return nil
}

func (d *documentationGenerator) scrape(ctx context.Context, providers []provider.Addr, opts []Opts) error {
	// TODO add filtering for removals to the function signature similar to modules.

	cfg := GenerateConfig{}
	for _, opt := range opts {
		if err := opt(ctx, &cfg); err != nil {
			return err
		}
	}
	if err := cfg.applyDefaults(); err != nil {
		return err
	}

	existingProviders, err := d.destination.GetProviderList(ctx)
	if err != nil {
		var notFound *providerindexstorage.ProviderListNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("failed to fetch provider version list (%w)", err)
		}
	}

	var eg errgroup.Group
	eg.SetLimit(25)

	lock := &sync.Mutex{}
	var providersToAdd []*providertypes.Provider
	var providersToRemove []provider.Addr
	for _, addr := range providers {
		eg.Go(func() error {
			blocked, blockedReason := d.blocklist.IsProviderBlocked(addr)

			providerEntry := existingProviders.GetProvider(addr)
			needsAdd := false
			if providerEntry == nil {
				providerEntry = &providertypes.Provider{
					Addr:          providertypes.Addr(addr),
					Link:          "",
					Description:   "",
					Versions:      nil,
					IsBlocked:     blocked,
					BlockedReason: blockedReason,
				}
				needsAdd = true
			}

			// scrape the docs into their own directory
			if err := d.scrapeProvider(ctx, providertypes.Addr(addr), providerEntry, cfg, blocked, blockedReason); err != nil {
				var notFound *metadata.ProviderNotFoundError
				if errors.As(err, &notFound) {
					d.log.Info(ctx, "Provider %s not found, removing from UI... (%v)", addr, err)
					lock.Lock()
					providersToRemove = append(providersToRemove, addr)
					lock.Unlock()
					return nil
				}
				d.log.Error(ctx, "Failed to scrape provider %s/%s: %v", addr.Namespace, addr.Name, err)
				return err
			}

			if err := d.destination.StoreProvider(ctx, *providerEntry); err != nil {
				return fmt.Errorf("failed to store provider %s (%w)", addr, err)
			}

			if needsAdd {
				lock.Lock()
				providersToAdd = append(providersToAdd, providerEntry)
				lock.Unlock()
			}

			return nil

		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}
	existingProviders.AddProviders(providersToAdd...)
	// TODO remove providers that are no longer needed.

	if err := d.destination.StoreProviderList(ctx, existingProviders); err != nil {
		return fmt.Errorf("failed to store provider list (%w)", err)
	}

	return nil
}

func (d *documentationGenerator) scrapeProvider(ctx context.Context, addr providertypes.ProviderAddr, providerData *providertypes.Provider, cfg GenerateConfig, blocked bool, blockedReason string) error {
	d.log.Trace(ctx, "Generating index for provider %s/%s", addr.Namespace, addr.Name)

	canonicalAddr, err := d.metadataAPI.GetProviderCanonicalAddr(ctx, addr.Addr)
	if err != nil {
		return err
	}

	meta, err := d.metadataAPI.GetProvider(ctx, canonicalAddr, false)
	if err != nil {
		return err
	}

	if meta.CustomRepository != "" {
		providerData.Link = meta.CustomRepository
	} else {
		link, err := d.vcsClient.GetRepositoryBrowseURL(ctx, canonicalAddr.ToRepositoryAddr())
		if err != nil {
			providerData.Link = link
		}
	}

	// Reverse the version order to ensure that the search index is updated with newer versions overriding older
	// versions.
	slices.Reverse(meta.Versions)
	repoInfoFetched := false
	var versionsToAdd []providertypes.ProviderVersionDescriptor
	var versionsToUpdate []providertypes.ProviderVersionDescriptor

	forceProvider := cfg.Force.MustRegenerateProvider(ctx, addr.Addr)
	if blocked != providerData.IsBlocked {
		// If the blocked status has changed, force re-generating everything to make sure all previous content is gone.
		forceProvider = true
		d.log.Info(ctx, "Provider %s changed blocked status, reindexing all versions...", addr)
	}

	providerData.Warnings = meta.Warnings

	if cfg.ForceRepoDataUpdate {
		d.extractRepoInfo(ctx, addr, providerData)
		repoInfoFetched = true
	}

	for _, version := range meta.Versions {
		if err := version.Version.Validate(); err != nil {
			d.log.Warn(ctx, "Invalid version number for provider %s: %s, skipping... (%v)", addr, version.Version, err)
			continue
		}
		hasVersion := providerData.HasVersion(version.Version)
		if hasVersion && !forceProvider {
			d.log.Debug(ctx, "The provider index already has %s version %s, skipping...", addr, version.Version)
			continue
		}
		if !repoInfoFetched {
			d.extractRepoInfo(ctx, addr, providerData)
			repoInfoFetched = true
		}

		providerVersion, err := d.scrapeVersion(ctx, addr, canonicalAddr, providerData, version, blocked, blockedReason)
		if err != nil {
			var repoNotFound *vcs.RepositoryNotFoundError
			if errors.As(err, &repoNotFound) {
				d.log.Warn(ctx, "Repository not found for provider %s, skipping... (%v)", addr.Addr, err)
				return &metadata.ProviderNotFoundError{
					ProviderAddr: addr.Addr,
					Cause:        err,
				}
			}
			var versionNotFound *vcs.VersionNotFoundError
			if errors.As(err, &versionNotFound) {
				d.log.Warn(ctx, "Version %s not found for provider %s, skipping (%v)", version.Version, addr.Addr, err)
				// We don't care, don't add it to the version list.
				continue
			}
			return err
		}
		if hasVersion {
			versionsToUpdate = append(versionsToUpdate, providerVersion.ProviderVersionDescriptor)
			// TODO: currently we don't remove documents that may not be there anymore. This should be addressed by
			//       diffing the old and new descriptor.
		} else {
			versionsToAdd = append(versionsToAdd, providerVersion.ProviderVersionDescriptor)
		}
	}

	providerData.AddVersions(versionsToAdd...)
	providerData.UpdateVersions(versionsToUpdate...)

	// TODO remove versions that no longer exist.

	if !canonicalAddr.Equals(addr.Addr) {
		canonicalAddrStruct := providertypes.Addr(canonicalAddr)
		providerData.CanonicalAddr = &canonicalAddrStruct
	} else {
		providerData.CanonicalAddr = nil
	}

	reverseAliases, err := d.metadataAPI.GetProviderReverseAliases(ctx, addr.Addr)
	if err != nil {
		return err
	}
	providerData.ReverseAliases = make([]providertypes.ProviderAddr, len(reverseAliases))
	for i, reverseAlias := range reverseAliases {
		providerData.ReverseAliases[i] = providertypes.Addr(reverseAlias)
	}

	return nil
}

func (d *documentationGenerator) extractRepoInfo(ctx context.Context, addr providertypes.ProviderAddr, providerData *providertypes.Provider) {
	// Make sure to fetch the description for the search index:
	repoInfo, err := d.vcsClient.GetRepositoryInfo(ctx, addr.ToRepositoryAddr())
	if err != nil {
		var repoNotFound *vcs.RepositoryNotFoundError
		if errors.As(err, &repoNotFound) {
			d.log.Warn(ctx, "Repository not found for provider %s, skipping... (%v)", addr.String(), err)
			return
		}
		// We handle description errors as soft errors because they are purely presentational.
		d.log.Warn(ctx, "Cannot update repository description for provider %s (%v)", addr.String(), err)
		return
	}
	providerData.Description = repoInfo.Description
	providerData.Popularity = repoInfo.Popularity
	providerData.ForkCount = repoInfo.ForkCount

	forkRepo := repoInfo.ForkOf
	if forkRepo == nil {
		return
	}
	link, err := d.vcsClient.GetRepositoryBrowseURL(ctx, *forkRepo)
	if err != nil {
		d.log.Warn(ctx, "Cannot determine repository browse URL for %s (%v)", forkRepo.String(), err)
		return
	}
	providerData.ForkOfLink = link

	forkedAddr, err := provider.AddrFromRepository(*forkRepo)
	if err != nil {
		d.log.Warn(ctx, "Cannot convert repository name %s to a provider addr (%v)", forkRepo.String(), err)
		return
	}
	_, err = d.metadataAPI.GetProvider(ctx, forkedAddr, false)
	if err != nil {
		return
	}
	providerData.ForkOf = providertypes.Addr(forkedAddr)

	upstreamRepoInfo, err := d.vcsClient.GetRepositoryInfo(ctx, *forkRepo)
	if err != nil {
		d.log.Warn(ctx, "Cannot fetch upstream repository info for %s (%v)", forkRepo.String(), err)
		return
	}
	providerData.UpstreamPopularity = upstreamRepoInfo.Popularity
	providerData.UpstreamForkCount = upstreamRepoInfo.ForkCount
}

func (d *documentationGenerator) scrapeVersion(ctx context.Context, addr providertypes.ProviderAddr, canonicalAddr provider.Addr, providerDetails *providertypes.Provider, version provider.Version, blocked bool, blockedReason string) (providertypes.ProviderVersion, error) {
	// We get the VCS version before normalizing as the tag name may be different.
	vcsVersion := version.Version.ToVCSVersion()
	version.Version = version.Version.Normalize()
	d.log.Info(ctx, "Scraping documentation for %s version %s...", addr, version.Version)

	// TODO get the release date instead of the tag date
	tag, err := d.vcsClient.GetTagVersion(ctx, canonicalAddr.ToRepositoryAddr(), vcsVersion)
	if err != nil {
		var verNotFoundError *vcs.VersionNotFoundError
		if !errors.As(err, &verNotFoundError) {
			return providertypes.ProviderVersion{}, err
		}

		// Fallback for missing/added v prefix.
		if strings.HasPrefix(string(vcsVersion), "v") {
			vcsVersion = vcs.VersionNumber(strings.TrimPrefix(string(vcsVersion), "v"))
		} else {
			vcsVersion = "v" + vcsVersion
		}
		tag, err = d.vcsClient.GetTagVersion(ctx, canonicalAddr.ToRepositoryAddr(), vcsVersion)
		if err != nil {
			return providertypes.ProviderVersion{}, err
		}
	}

	workingCopy, err := d.vcsClient.Checkout(ctx, canonicalAddr.ToRepositoryAddr(), vcsVersion)
	if err != nil {
		return providertypes.ProviderVersion{}, err
	}
	defer func() {
		if err := workingCopy.Close(); err != nil {
			d.log.Error(ctx, "Failed to close working copy for %s (%v)", addr, err)
		}
	}()

	providerData, err := d.source.Describe(ctx, workingCopy, blocked, blockedReason)
	if err != nil {
		return providertypes.ProviderVersion{}, err
	}

	versionDescriptor := providertypes.ProviderVersionDescriptor{
		ID:        version.Version.Normalize(),
		Published: tag.Created,
	}

	// Make sure we delete all old data in case a re-indexing needs to institute a block and remove everything.
	if err := d.destination.DeleteProviderVersion(ctx, addr.Addr, versionDescriptor.ID); err != nil {
		return providertypes.ProviderVersion{}, err
	}

	versionData, err := providerData.Store(ctx, addr.Addr, versionDescriptor, d.destination)
	if err != nil {
		return versionData, fmt.Errorf("failed to store documentation for %s version %s (%w)", addr, version.Version, err)
	}

	if err := d.search.indexProviderVersion(ctx, addr.Addr, providerDetails, versionData); err != nil {
		return versionData, err
	}

	return versionData, nil
}
