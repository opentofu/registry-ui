package providertypes

import (
	"slices"

	"github.com/opentofu/libregistry/types/provider"
)

// Provider is a single provider with all its versions.
type Provider struct {
	// If you add something here, don't forget to update the Equals() and DeepCopy() functions below.

	// Addr holds the address of a provider. It can be split by / to obtain a namespace and name.

	Addr ProviderAddr `json:"addr"`
	// Warnings contains a list of warning strings issued to the OpenTofu client when fetching the provider info. This
	// typically indicates a deprecation or move of the provider to another location.
	Warnings []string `json:"warnings,omitempty"`
	// Link contains the link to the repository this provider was built from. Note that this may not match the
	// Addr field since the repository may be different. Note that this field may not be available for all
	// providers.
	Link string `json:"link"`
	// CanonicalAddr stores the canonical address of the provider. If this is set, it signals that there
	// is an alias in place. The canonical address describes the repository to ultimately fetch the data from.
	CanonicalAddr *ProviderAddr `json:"canonical_addr"`
	// ReverseAliases contains a list of providers that are aliases of the current one. This field is the inverse of
	// CanonicalAddr.
	ReverseAliases []ProviderAddr `json:"reverse_aliases"`
	// Description is the extracted description for the provider. This may be empty.
	Description string `json:"description"`
	// Popularity indicates how popular the underlying repository is in the VCS system.
	Popularity int `json:"popularity"`
	// ForkCount indicates how many forks this provider has.
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
	Versions []ProviderVersionDescriptor `json:"versions"`
	IsBlocked bool `json:"is_blocked"`
	BlockedReason string `json:"blocked_reason,omitempty"`

	// If you add something here, don't forget to update the Equals() and DeepCopy() functions below.
}

// Equals returns true if and only of all parameters of the two providers are equal (with a deep comparison).
func (p *Provider) Equals(other *Provider) bool {
	if p == other {
		return true
	}
	if p == nil || other == nil {
		return false
	}
	if len(p.ReverseAliases) != len(other.ReverseAliases) {
		return false
	} else {
		for i := range len(p.ReverseAliases) {
			if !p.ReverseAliases[i].Equals(other.ReverseAliases[i].Addr) {
				return false
			}
		}
	}
	return p.Addr.Equals(other.Addr.Addr) && slices.Equal(p.Warnings, other.Warnings) && p.Link == other.Link &&
		(p.CanonicalAddr == other.CanonicalAddr || p.CanonicalAddr.Equals(other.CanonicalAddr.Addr)) &&
		p.Description == other.Description && p.Popularity == other.Popularity && p.ForkCount == other.ForkCount &&
		p.ForkOfLink == other.ForkOfLink && p.ForkOf == other.ForkOf &&
		p.UpstreamPopularity == other.UpstreamPopularity && p.UpstreamForkCount == other.UpstreamForkCount &&
		slices.Equal(p.Versions, other.Versions) && p.BlockedReason == other.BlockedReason
}

// DeepCopy creates a deep copy of the Provider.
func (p *Provider) DeepCopy() *Provider {
	warnings := make([]string, len(p.Warnings))
	copy(warnings, p.Warnings)

	var canonicalAddr *ProviderAddr
	if p.CanonicalAddr != nil {
		canonicalAddr = &ProviderAddr{
			Addr:      p.CanonicalAddr.Addr,
			Display:   p.CanonicalAddr.Display,
			Namespace: p.CanonicalAddr.Namespace,
			Name:      p.CanonicalAddr.Name,
		}
	}

	reverseAliases := make([]ProviderAddr, len(p.ReverseAliases))
	copy(reverseAliases, p.ReverseAliases)

	versions := make([]ProviderVersionDescriptor, len(p.Versions))
	copy(versions, p.Versions)

	return &Provider{
		Addr:               p.Addr,
		Warnings:           warnings,
		Link:               p.Link,
		CanonicalAddr:      canonicalAddr,
		ReverseAliases:     reverseAliases,
		Description:        p.Description,
		Popularity:         p.Popularity,
		ForkCount:          p.ForkCount,
		ForkOfLink:         p.ForkOfLink,
		ForkOf:             p.ForkOf,
		UpstreamPopularity: p.UpstreamPopularity,
		UpstreamForkCount:  p.UpstreamForkCount,
		Versions:           versions,
		IsBlocked:          p.IsBlocked,
		BlockedReason:      p.BlockedReason,
	}
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
