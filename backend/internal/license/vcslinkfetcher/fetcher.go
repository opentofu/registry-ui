package vcslinkfetcher

import (
	"context"
	"errors"

	"github.com/opentofu/libregistry/vcs"
	"github.com/opentofu/registry-ui/internal/license"
)

// Fetcher creates a license fetcher from a VCS system.
func Fetcher(ctx context.Context, repository vcs.RepositoryAddr, version vcs.VersionNumber, vcsClient vcs.Client) license.LinkFetcher {
	return func(license *license.License) error {
		link, err := vcsClient.GetFileViewURL(ctx, repository, version, license.File)
		if err != nil {
			var noWebAccessError *vcs.NoWebAccessError
			if errors.As(err, &noWebAccessError) {
				return nil
			}
			return err
		}
		license.Link = link
		return nil
	}
}
