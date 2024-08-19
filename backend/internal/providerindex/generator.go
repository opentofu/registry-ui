package providerindex

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/libregistry/metadata"
	"github.com/opentofu/libregistry/types/provider"
	"github.com/opentofu/libregistry/vcs"
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
	Generate(ctx context.Context) error

	// GenerateNamespace generates module index files incrementally for one namespace and removes items no longer in the
	// registry.
	GenerateNamespace(ctx context.Context, namespace string) error
}

func NewDocumentationGenerator(
	log logger.Logger,
	metadataAPI metadata.API,
	vcsClient vcs.Client,
	licenseDetector license.Detector,
	source providerdocsource.API,
	destination providerindexstorage.API,
	searchAPI search.API,
) DocumentationGenerator {
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
}

func (d *documentationGenerator) Generate(ctx context.Context) error {
	d.log.Info(ctx, "Listing all providers...")
	providerList, err := d.metadataAPI.ListProviders(ctx, true)
	if err != nil {
		return err
	}
	d.log.Info(ctx, "Loaded %d providers", len(providerList))

	err = d.scrape(ctx, providerList)
	if err != nil {
		return err
	}

	return nil
}

func (d *documentationGenerator) GenerateNamespace(ctx context.Context, namespace string) error {
	d.log.Info(ctx, "Listing all providers in namespace %s...", namespace)
	providerList, err := d.metadataAPI.ListProvidersByNamespace(ctx, namespace, true)
	if err != nil {
		return err
	}
	d.log.Info(ctx, "Loaded %d providers", len(providerList))

	err = d.scrape(ctx, providerList)
	if err != nil {
		return err
	}

	return nil
}

func (d *documentationGenerator) scrape(ctx context.Context, providers []provider.Addr) error {
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
			providerEntry := existingProviders.GetProvider(addr)
			needsAdd := false
			if providerEntry == nil {
				providerEntry = &providertypes.Provider{
					Addr:        providertypes.Addr(addr),
					Description: "",
					Versions:    nil,
				}
				needsAdd = true
			}

			// scrape the docs into their own directory
			if err := d.scrapeProvider(ctx, providertypes.Addr(addr), providerEntry); err != nil {
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

func (d *documentationGenerator) scrapeProvider(ctx context.Context, addr providertypes.ProviderAddr, providerData *providertypes.Provider) error {
	d.log.Trace(ctx, "Generating index for provider %s/%s", addr.Namespace, addr.Name)

	canonicalAddr, err := d.metadataAPI.GetProviderCanonicalAddr(ctx, addr.Addr)
	if err != nil {
		return err
	}

	meta, err := d.metadataAPI.GetProvider(ctx, canonicalAddr, false)
	if err != nil {
		return err
	}

	// Reverse the version order to ensure that the search index is updated with newer versions overriding older
	// versions.
	slices.Reverse(meta.Versions)
	repoInfoFetched := false
	var versionsToAdd []providertypes.ProviderVersionDescriptor
	for _, version := range meta.Versions {
		if err := version.Version.Validate(); err != nil {
			d.log.Warn(ctx, "Invalid version number for provider %s: %s, skipping... (%v)", addr, version.Version, err)
			continue
		}
		if providerData.HasVersion(version.Version) {
			d.log.Debug(ctx, "The provider index already has %s version %s, skipping...", addr, version.Version)
			continue
		}
		if !repoInfoFetched {
			// Make sure to fetch the description for the search index:
			repoInfoFetched = true
			repoInfo, err := d.vcsClient.GetRepositoryInfo(ctx, addr.ToRepositoryAddr())
			if err != nil {
				var repoNotFound *vcs.RepositoryNotFoundError
				if errors.As(err, &repoNotFound) {
					d.log.Warn(ctx, "Repository not found for provider %s, skipping... (%v)", addr.String(), err)
					break
				}
				// We handle description errors as soft errors because they are purely presentational.
				d.log.Warn(ctx, "Cannot update repository description for provider %s (%v)", addr.String(), err)
			} else {
				providerData.Description = repoInfo.Description
			}
		}

		providerVersion, err := d.scrapeVersion(ctx, addr, canonicalAddr, version)
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
				d.log.Warn(ctx, "Version %s not found for provider %s, skipping (%v)", addr.Addr, err)
				// We don't care, don't add it to the version list.
				continue
			}
			return err
		}
		versionsToAdd = append(versionsToAdd, providerVersion.ProviderVersionDescriptor)
	}

	providerData.AddVersions(versionsToAdd...)

	return nil
}

func (d *documentationGenerator) scrapeVersion(ctx context.Context, addr providertypes.ProviderAddr, canonicalAddr provider.Addr, version provider.Version) (providertypes.ProviderVersion, error) {
	version.Version = version.Version.Normalize()
	d.log.Info(ctx, "Scraping documentation for %s version %s...", addr, version.Version)

	// TODO get the release date instead of the tag date
	tag, err := d.vcsClient.GetTagVersion(ctx, canonicalAddr.ToRepositoryAddr(), version.Version.ToVCSVersion())
	if err != nil {
		return providertypes.ProviderVersion{}, err
	}

	workingCopy, err := d.vcsClient.Checkout(ctx, canonicalAddr.ToRepositoryAddr(), version.Version.ToVCSVersion())
	if err != nil {
		return providertypes.ProviderVersion{}, err
	}
	defer func() {
		if err := workingCopy.Close(); err != nil {
			d.log.Error(ctx, "Failed to close working copy for %s (%v)", addr, err)
		}
	}()

	providerData, err := d.source.Describe(ctx, workingCopy)
	if err != nil {
		return providertypes.ProviderVersion{}, err
	}

	versionDescriptor := providertypes.ProviderVersionDescriptor{
		ID:        version.Version.Normalize(),
		Published: tag.Created,
	}

	versionData, err := providerData.Store(ctx, addr.Addr, versionDescriptor, d.destination)
	if err != nil {
		return versionData, fmt.Errorf("failed to store documentation for %s version %s (%w)", addr, version.Version, err)
	}

	if err := d.search.indexProviderVersion(ctx, addr.Addr, versionData); err != nil {
		return versionData, err
	}

	return versionData, nil
}
