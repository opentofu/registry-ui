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
	// Description is the extracted description for the provider. This may be empty.
	//
	// required: true
	Description string `json:"description"`
	// Versions holds the list of versions this provider supports.
	//
	// required: true
	Versions []ProviderVersionDescriptor `json:"versions"`
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
