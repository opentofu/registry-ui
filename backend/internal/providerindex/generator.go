package providerindex

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/opentofu/libregistry/vcs"
	"github.com/opentofu/registry-ui/internal/registry/github"
	"github.com/opentofu/registry-ui/internal/registry/provider"

	"github.com/opentofu/registry-ui/internal/providerindex/providerdocsource"
	"github.com/opentofu/registry-ui/internal/providerindex/providerindexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"

	"github.com/opentofu/registry-ui/internal/license"
)

func GenerateDocumentation(log *slog.Logger, ctx context.Context, providers provider.List, licenseDetector license.Detector, destination providerindexstorage.API, db *sql.DB, workDir string) error {
	log = log.With(slog.String("name", "Provider indexer"))

	// TODO add filtering for removals to the function signature similar to modules.

	existingProviders, err := destination.GetProviderList(ctx)
	if err != nil {
		var notFound *providerindexstorage.ProviderListNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("failed to fetch provider version list (%w)", err)
		}
	}
	// TODO fill in from PG
	//var existingProviders providertypes.ProviderList

	lock := &sync.Mutex{}
	var providersToAdd []*providertypes.Provider
	var providersToRemove []providertypes.ProviderAddr

	processProviders := func(filterHC bool) provider.Action {
		return func(raw provider.Provider) error {
			addr := providertypes.Addr(raw)

			isHC := addr.Namespace == "hashicorp"
			if (filterHC && !isHC) || (!filterHC && isHC) {
				return nil
			}

			// We are fetching the provider entry from the megaindex and storing it
			// further down below as a separate index file so the frontend has an easier time
			// fetching it.
			providerEntry := existingProviders.GetProvider(addr)
			// originalProviderEntry serves the purpose of being an original copy to compare to
			// so we don't write the index if it hasn't actually been modified to save costs.
			var originalProviderEntry *providertypes.Provider
			needsAdd := false
			if providerEntry == nil {
				providerEntry = &providertypes.Provider{
					Addr:        addr,
					Link:        "",
					Description: "",
					Versions:    nil,
				}
				needsAdd = true
			} else {
				originalProviderEntry = providerEntry.DeepCopy()
			}

			// scrape the docs into their own directory
			if err := scrapeProvider(log, ctx, raw, providerEntry, workDir, licenseDetector, destination, db); err != nil {
				var notFound ProviderNotFoundError
				if errors.As(err, &notFound) {
					log.InfoContext(ctx, "Provider %s not found, removing from UI... (%v)", addr, err)
					lock.Lock()
					providersToRemove = append(providersToRemove, addr)
					lock.Unlock()
					return nil
				}
				log.ErrorContext(ctx, "Failed to scrape provider %s/%s: %v", addr.Namespace, addr.Name, err)
				return err
			}

			// Some providers may have versions detected by the registry, but somehow not in libregistry+scrape
			// Filter them out here for now
			if len(providerEntry.Versions) == 0 {
				log.InfoContext(ctx, "Provider %s does not have any versions, removing from UI...", addr)
				lock.Lock()
				providersToRemove = append(providersToRemove, addr)
				lock.Unlock()
				return nil
			}

			// Here we compare the provider entry to its original copy to make sure
			// we are only writing this index if needed. This is needed because writes
			// on R2 cost money, whereas reads don't and updating all the provider and
			// module indexes on every run costs ~300$ per month.
			if originalProviderEntry == nil || !originalProviderEntry.Equals(providerEntry) {
				if err := destination.StoreProvider(ctx, *providerEntry); err != nil {
					return fmt.Errorf("failed to store provider %s (%w)", addr, err)
				}
			}

			if needsAdd {
				lock.Lock()
				providersToAdd = append(providersToAdd, providerEntry)
				lock.Unlock()
			}

			return nil
		}
	}

	if err := providers.Parallel(25, processProviders(false)); err != nil {
		return err
	}
	if err := providers.Parallel(25, processProviders(true)); err != nil {
		return err
	}

	existingProviders.AddProviders(providersToAdd...)
	// TODO remove providers that are no longer needed.

	if err := destination.StoreProviderList(ctx, existingProviders); err != nil {
		return fmt.Errorf("failed to store provider list (%w)", err)
	}

	return nil
}

func scrapeProvider(log *slog.Logger, ctx context.Context, raw provider.Provider, providerData *providertypes.Provider, workDir string, licenseDetector license.Detector, destination providerindexstorage.API, db *sql.DB) error {
	addr := providertypes.Addr(raw)

	log.DebugContext(ctx, "Generating index for provider %s/%s", addr.Namespace, addr.Name)

	providerData.Link = raw.RepositoryURL()

	meta, err := raw.ReadMetadata()
	if err != nil {
		return err
	}

	// Reverse the version order to ensure that the search index is updated with newer versions overriding older
	// versions.
	slices.Reverse(meta.Versions)
	var versionsToAdd []providertypes.ProviderVersionDescriptor
	var versionsToUpdate []providertypes.ProviderVersionDescriptor

	providerData.Warnings = meta.Warnings

	//providerVersions(db, providerData)

	forceRepoInfo := true // CAM72CAM This should be done once a day/week/something?
	if forceRepoInfo {    //|| len(versionsToAdd) != 0 || len(versionsToUpdate) != 0 {
		extractRepoInfo(ctx, log, raw, providerData)
	}

	for _, version := range meta.Versions {
		hasVersion := providerData.HasVersion("v" + version.Version)
		if hasVersion && false {
			log.DebugContext(ctx, "The provider index already has %s version %s, skipping...", addr, version.Version)
			continue
		}

		// FUTURE: consider use git --work-tree=$TAG_DIR checkout $TAG  -- . for per-version checkouts

		providerVersion, err := scrapeVersion(ctx, log, raw, providerData, version, workDir, licenseDetector, destination)
		if err != nil {
			var repoNotFound *RepositoryNotFoundError
			if errors.As(err, &repoNotFound) {
				log.WarnContext(ctx, "Repository not found for provider %s, skipping... (%v)", addr, err)
				/* TODO CAM72CAM return &metadata.ProviderNotFoundError{
					ProviderAddr: addr,
					Cause:        err,
				}*/
			}
			var versionNotFound *VersionNotFoundError
			if errors.As(err, &versionNotFound) {
				log.WarnContext(ctx, "Version %s not found for provider %s, skipping (%v)", version.Version, addr, err)
				// We don't care, don't add it to the version list.
				continue
			}
			return err
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if err := insertItems(tx, providerData, providerVersion); err != nil {
			err = errors.Join(err, tx.Rollback())
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		// TEMPORARY

		if hasVersion {
			// TODO: currently we don't remove documents that may not be there anymore. This should be addressed by
			//       diffing the old and new descriptor.
			versionsToUpdate = append(versionsToUpdate, providerVersion.ProviderVersionDescriptor)
		} else {
			versionsToAdd = append(versionsToAdd, providerVersion.ProviderVersionDescriptor)
		}
	}

	providerData.AddVersions(versionsToAdd...)
	providerData.UpdateVersions(versionsToUpdate...)

	// TODO remove versions that no longer exist.

	/* TODO cam72cam
	if !canonicalAddr.Equals(addr.Addr) {
		canonicalAddrStruct := providertypes.Addr(canonicalAddr)
		providerData.CanonicalAddr = &canonicalAddrStruct
	} else {
		providerData.CanonicalAddr = nil
	}

	reverseAliases, err := metadataAPI.GetProviderReverseAliases(ctx, addr.Addr)
	if err != nil {
		return err
	}
	providerData.ReverseAliases = make([]providertypes.ProviderAddr, len(reverseAliases))
	for i, reverseAlias := range reverseAliases {
		providerData.ReverseAliases[i] = providertypes.Addr(reverseAlias)
	}
	*/

	return nil
}

func extractRepoInfo(ctx context.Context, log *slog.Logger, raw provider.Provider, providerData *providertypes.Provider) {
	addr := providertypes.Addr(raw)

	// Make sure to fetch the description for the search index:
	repoInfo, err := getRepositoryInfo(ctx, raw)
	if err != nil {
		var repoNotFound *vcs.RepositoryNotFoundError
		if errors.As(err, &repoNotFound) {
			log.WarnContext(ctx, "Repository not found for provider %s, skipping... (%v)", addr.String(), err)
			return
		}
		// We handle description errors as soft errors because they are purely presentational.
		log.WarnContext(ctx, "Cannot update repository description for provider %s (%v)", addr.String(), err)
		return
	}
	providerData.Description = repoInfo.Description
	providerData.Popularity = repoInfo.StargazersCount
	providerData.ForkCount = repoInfo.ForkCount

	forkRepo := repoInfo.Parent
	if forkRepo == nil {
		return
	}
	providerData.ForkOfLink = forkRepo.HTMLUrl

	/*CAM72CAM This is Useless!
	forkedAddr, err := lrprovider.AddrFromRepository(*forkRepo)
	if err != nil {
		log.WarnContext(ctx, "Cannot convert repository name %s to a provider addr (%v)", forkRepo.String(), err)
		return
	}
	_, err = metadataAPI.GetProvider(ctx, forkedAddr, false)
	if err != nil {
		return
	}
	providerData.ForkOf = providertypes.Addr(forkedAddr)

	upstreamRepoInfo, err := vcsClient.GetRepositoryInfo(ctx, *forkRepo)
	if err != nil {
		log.WarnContext(ctx, "Cannot fetch upstream repository info for %s (%v)", forkRepo.String(), err)
		return
	}
	providerData.UpstreamPopularity = upstreamRepoInfo.Popularity
	providerData.UpstreamForkCount = upstreamRepoInfo.ForkCount*/
}

func scrapeVersion(ctx context.Context, log *slog.Logger, raw provider.Provider, providerDetails *providertypes.Provider, version provider.Version, workDir string, licenseDetector license.Detector, destination providerindexstorage.API) (providertypes.ProviderVersion, error) {
	addr := providertypes.Addr(raw)

	// We get the VCS version before normalizing as the tag name may be different.
	log.InfoContext(ctx, "Scraping documentation for %s version %s...", addr, version.Version)

	repoPath := path.Join(workDir, addr.Namespace, addr.Name)
	if _, err := os.Stat(repoPath); errors.Is(err, fs.ErrNotExist) {
		err := os.MkdirAll(repoPath, 0700)
		if err != nil {
			return providertypes.ProviderVersion{}, err
		}

		gitClone := exec.Command("git", "clone", raw.RepositoryURL(), repoPath)
		gitCloneOut, err := gitClone.Output()
		log.Info(string(gitCloneOut)) // CAM72CAM TODO BETTER OUTPUT
		if err != nil {
			// TODO CAM72CAM ExitError.Stderr
			return providertypes.ProviderVersion{}, RepositoryNotFoundError{
				RepositoryAddr: raw.RepositoryURL(),
				Cause:          err,
			}
		}
	} else if err != nil {
		return providertypes.ProviderVersion{}, err
	}

	// TODO CAM72CAM v prefix handling
	gitCheckout := exec.Command("git", "reset", "--hard", "v"+version.Version)
	gitCheckout.Dir = repoPath
	gitCheckoutOut, err := gitCheckout.Output()
	log.Info(string(gitCheckoutOut)) // CAM72CAM TODO BETTER OUTPUT
	if err != nil {
		// TODO CAM72CAM check tags to see if it's been deleted
		// TODO CAM72CAM ExitError.Stderr
		return providertypes.ProviderVersion{}, VersionNotFoundError{
			RepositoryAddr: raw.RepositoryURL(),
			Version:        "v" + version.Version,
			Cause:          err,
		}
	}

	gitCreated := exec.Command("git", "show", "--no-patch", "--no-notes", "--pretty=%cI", "HEAD")
	gitCreated.Dir = repoPath
	gitCreatedOut, err := gitCreated.Output()
	if err != nil {
		// TODO CAM72CAM ExitError.Stderr
		return providertypes.ProviderVersion{}, err
	}
	created, err := time.Parse(time.RFC3339, strings.TrimSpace(string(gitCreatedOut)))
	if err != nil {
		return providertypes.ProviderVersion{}, err
	}

	versionDescriptor := providertypes.ProviderVersionDescriptor{
		ID:        "v" + version.Version,
		Published: created.UTC(),
	}

	versionData, err := providerdocsource.Process(ctx, raw, repoPath, versionDescriptor, licenseDetector, log, destination)
	if err != nil {
		return providertypes.ProviderVersion{}, err
	}

	/* TODO cam72cam
	if err := search.indexProviderVersion(ctx, addr.Addr, providerDetails, versionData); err != nil {
		return versionData, err
	}*/

	return versionData, nil
}

// Lifted from libregistry
type ProviderNotFoundError struct {
	ProviderAddr providertypes.ProviderAddr
	Cause        error
}

func (m ProviderNotFoundError) Error() string {
	if m.Cause != nil {
		return "Provider not found: " + m.ProviderAddr.String() + " (" + m.Cause.Error() + ")"
	}
	return "Provider not found: " + m.ProviderAddr.String()
}

func (m ProviderNotFoundError) Unwrap() error {
	return m.Cause
}

// From libregistry
type repoInfoResponse struct {
	Description     string `json:"description"`
	StargazersCount int    `json:"stargazers_count"`
	ForkCount       int    `json:"forks_count"`
	Parent          *struct {
		HTMLUrl string `json:"html_url"`
		Name    string `json:"name"`
		Owner   struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"parent"`
}

func getRepositoryInfo(ctx context.Context, raw provider.Provider) (repoInfoResponse, error) {
	var response repoInfoResponse

	// TODO cam72cam retry logic
	// TODO  cam72cam errors

	token, err := github.EnvAuthToken()
	if err != nil {
		return response, err
	}

	client := http.DefaultClient

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/"+raw.EffectiveNamespace()+"/"+raw.RepositoryName(), nil)
	if err != nil {
		return response, err
	}
	req.Header.Set("User-Agent", github.UserAgent)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("Expected 200, got %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)

	return response, err
}

type RepositoryNotFoundError struct {
	RepositoryAddr string
	Cause          error
}

func (r RepositoryNotFoundError) Error() string {
	if r.Cause != nil {
		return "Repository not found: " + r.RepositoryAddr + " (" + r.Cause.Error() + ")"
	}
	return "Repository not found: " + r.RepositoryAddr
}

func (r RepositoryNotFoundError) Unwrap() error {
	return r.Cause
}

type VersionNotFoundError struct {
	RepositoryAddr string
	Version        string
	Cause          error
}

func (v VersionNotFoundError) Error() string {
	if v.Cause != nil {
		return "Version " + v.Version + " not found in repository" + v.RepositoryAddr + " (" + v.Cause.Error() + ")"
	}
	return "Version " + v.Version + " not found in repository" + v.RepositoryAddr
}

func (v VersionNotFoundError) Unwrap() error {
	return v.Cause
}
