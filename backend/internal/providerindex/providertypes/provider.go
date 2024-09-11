package providertypes

import (
	"slices"

	"github.com/opentofu/libregistry/types/provider"
)

// Provider is a single provider with all its versions.
//
// swagger:model Provider
type Provider struct {
	// Addr holds the address of a provider. It can be split by / to obtain a namespace and name.
	//
	// required: true
	Addr ProviderAddr `json:"addr"`
	// CanonicalAddr stores the canonical address of the provider. If this is set, it signals that there
	// is an alias in place. The canonical address describes the repository to ultimately fetch the data from.
	//
	// required: false
	CanonicalAddr *ProviderAddr `json:"canonical_addr"`
	// ReverseAliases contains a list of providers that are aliases of the current one. This field is the inverse of
	// CanonicalAddr.
	// required: false
	ReverseAliases []ProviderAddr `json:"reverse_aliases"`
	// Description is the extracted description for the provider. This may be empty.
	//
	// required: true
	Description string `json:"description"`
	// Popularity indicates how popular the underlying repository is in the VCS system.
	// required: true
	Popularity int `json:"popularity"`
	// ForkCount indicates how many forks this provider has.
	// required: true
	ForkCount int `json:"fork_count"`
	// ForkOfLink may contain a link to a repository this provider is forked from.
	ForkOfLink string `json:"fork_of_link,omitempty"`
	// ForkOf indicates which provider this repository is forked from. This field may be empty even if
	// the ForkOfLink field is filled.
	ForkOf ProviderAddr `json:"fork_of,omitempty"`
	// UpstreamPopularity contains the popularity of the original repository this repository is forked of.
	UpstreamPopularity int `json:"upstream_popularity"`
	// UpstreamForkCount contains the number of forks of the upstream repository.
	UpstreamForkCount int `json:"upstream_fork_count"`
	// Versions holds the list of versions this provider supports.
	//
	// required: true
	Versions []ProviderVersionDescriptor `json:"versions"`
	// required: true
	IsBlocked bool `json:"is_blocked"`
	// required: false
	BlockedReason string `json:"blocked_reason,omitempty"`
}

func (p *Provider) Compare(other Provider) int {
	return p.Addr.Compare(other.Addr.Addr)
}

func (p *Provider) HasVersion(version provider.VersionNumber) bool {
	version = version.Normalize()
	for _, ver := range p.Versions {
		if ver.ID.Normalize() == version {
			return true
		}
	}
	return false
}

func (p *Provider) AddVersions(versions ...ProviderVersionDescriptor) {
	if len(versions) == 0 {
		return
	}
	p.Versions = append(p.Versions, versions...)

	slices.SortStableFunc(p.Versions, func(a, b ProviderVersionDescriptor) int {
		return -a.ID.Compare(b.ID)
	})
}

func (p *Provider) RemoveVersions(in provider.VersionList, notIn provider.VersionList) []provider.VersionNumber {
	inVersionNumberMap := map[provider.VersionNumber]struct{}{}
	notInVersionNumberMap := map[provider.VersionNumber]struct{}{}
	for _, version := range in {
		inVersionNumberMap[version.Version.Normalize()] = struct{}{}
	}
	for _, version := range notIn {
		notInVersionNumberMap[version.Version.Normalize()] = struct{}{}
	}
	var removedVersions []provider.VersionNumber

	var newVersions []ProviderVersionDescriptor
	for _, ver := range p.Versions {
		id := ver.ID.Normalize()
		_, notInOK := notInVersionNumberMap[id]
		_, inOK := inVersionNumberMap[id]
		if notInOK && !inOK {
			newVersions = append(newVersions, ver)
		} else {
			removedVersions = append(removedVersions, id)
		}
	}
	p.Versions = newVersions
	return removedVersions
}

func (p *Provider) UpdateVersions(updatedVersions ...ProviderVersionDescriptor) {
	for _, updatedVersion := range updatedVersions {
		for i, existingVersion := range p.Versions {
			if existingVersion.ID.Compare(updatedVersion.ID) == 0 {
				p.Versions[i].ID = updatedVersion.ID
				p.Versions[i].Published = updatedVersion.Published
			}
		}
	}
}
