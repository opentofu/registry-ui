package providertypes

import (
	"slices"

	"github.com/opentofu/libregistry/types/provider"
)

// TODO: move the request/response handling into a dedicated package alongside a proper web API.

// ProviderList is a list of providers.
type ProviderList struct {
	// Providers holds the list of providers.
	Providers []*Provider `json:"providers"`
}

func (m *ProviderList) AddProviders(providers ...*Provider) {
	m.Providers = append(m.Providers, providers...)
	slices.SortStableFunc(m.Providers, func(a, b *Provider) int {
		return a.Compare(*b)
	})
}

func (m *ProviderList) RemoveProviders(in []ProviderAddr, notIn []ProviderAddr, blockRemoval func(providerAddr ProviderAddr) bool, force []ProviderAddr) []ProviderAddr {
	removeMap := map[ProviderAddr]struct{}{}
	notRemoveMap := map[ProviderAddr]struct{}{}
	forceMap := map[ProviderAddr]struct{}{}
	for _, mod := range in {
		removeMap[mod] = struct{}{}
	}
	for _, mod := range notIn {
		notRemoveMap[mod] = struct{}{}
	}
	for _, f := range force {
		forceMap[f] = struct{}{}
	}

	//goland:noinspection GoPreferNilSlice
	newProviders := []*Provider{}
	var removedProviders []ProviderAddr
	for _, mod := range m.Providers {
		_, shouldNotRemove := notRemoveMap[mod.Addr]
		_, shouldRemove := removeMap[mod.Addr]
		_, shouldForceRemove := forceMap[mod.Addr]
		if !blockRemoval(mod.Addr) && ((shouldRemove && !shouldNotRemove) || shouldForceRemove) {
			removedProviders = append(removedProviders, mod.Addr)
		} else {
			newProviders = append(newProviders, mod)
		}
	}
	m.Providers = newProviders
	return removedProviders
}

func (m *ProviderList) HasProvider(providerAddr provider.Addr) bool {
	for _, mod := range m.Providers {
		if mod.Addr.Equals(providerAddr) {
			return true
		}
	}
	return false
}

func (m *ProviderList) GetProvider(providerAddr provider.Addr) *Provider {
	for i, mod := range m.Providers {
		mod := mod
		if mod.Addr.Equals(providerAddr) {
			return m.Providers[i]
		}
	}
	return nil
}
