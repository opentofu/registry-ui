package moduleindex

import (
	"slices"

	"github.com/opentofu/libregistry/types/module"
)

type ModuleList struct {
	Modules []*Module `json:"modules"`
}

func (m *ModuleList) addModules(modules ...*Module) {
	m.Modules = append(m.Modules, modules...)
	slices.SortStableFunc(m.Modules, func(a, b *Module) int {
		return a.Compare(*b)
	})
}

func (m *ModuleList) removeModules(in []ModuleAddr, notIn []module.Addr, blockRemoval func(moduleAddr ModuleAddr) bool, force []ModuleAddr) []ModuleAddr {
	removeMap := map[ModuleAddr]struct{}{}
	notRemoveMap := map[ModuleAddr]struct{}{}
	forceMap := map[ModuleAddr]struct{}{}
	for _, mod := range in {
		removeMap[mod] = struct{}{}
	}
	for _, mod := range notIn {
		notRemoveMap[Addr(mod)] = struct{}{}
	}
	for _, f := range force {
		forceMap[f] = struct{}{}
	}

	//goland:noinspection GoPreferNilSlice
	newModules := []*Module{}
	var removedModules []ModuleAddr
	for _, mod := range m.Modules {
		_, shouldNotRemove := notRemoveMap[mod.Addr]
		_, shouldRemove := removeMap[mod.Addr]
		_, shouldForceRemove := forceMap[mod.Addr]
		if !blockRemoval(mod.Addr) && ((shouldRemove && !shouldNotRemove) || shouldForceRemove) {
			removedModules = append(removedModules, mod.Addr)
		} else {
			newModules = append(newModules, mod)
		}
	}
	m.Modules = newModules
	return removedModules
}

func (m *ModuleList) HasModule(moduleAddr module.Addr) bool {
	for _, mod := range m.Modules {
		if mod.Addr.Equals(moduleAddr) {
			return true
		}
	}
	return false
}

func (m *ModuleList) GetModule(moduleAddr module.Addr) *Module {
	for i, mod := range m.Modules {
		mod := mod
		if mod.Addr.Equals(moduleAddr) {
			return m.Modules[i]
		}
	}
	return nil
}
