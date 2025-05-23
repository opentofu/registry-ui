package moduleindex

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/opentofu/libregistry/types/module"
)

type Module struct {
	// If you add a field here, update Equals() and DeepCopy() below.

	Addr          ModuleAddr                `json:"addr"`
	Description   string                    `json:"description"`
	Versions      []ModuleVersionDescriptor `json:"versions"`
	IsBlocked     bool                      `json:"is_blocked"`
	BlockedReason string                    `json:"blocked_reason,omitempty"`

	// Popularity indicates how popular the underlying repository is in the VCS system.
	Popularity int `json:"popularity"`
	// ForkCount indicates how many forks this provider has.
	ForkCount int `json:"fork_count"`
	// ForkOfLink may contain a link to a repository this provider is forked from.
	ForkOfLink string `json:"fork_of_link,omitempty"`
	// ForkOf indicates which module this repository is forked from. This field may be empty even if
	// the ForkOfLink field is filled.
	ForkOf ModuleAddr `json:"fork_of,omitempty"`
	// UpstreamPopularity contains the popularity of the original repository this repository is forked of.
	UpstreamPopularity int `json:"upstream_popularity"`
	// UpstreamForkCount contains the number of forks of the upstream repository.
	UpstreamForkCount int `json:"upstream_fork_count"`

	// If you add a field here, update Equals() and DeepCopy() below.
}

// Equals compares every parameter of the two modules and returns true if both are equal on a deep comparison.
func (m *Module) Equals(other *Module) bool {
	if m == other {
		return true
	}
	if m == nil || other == nil {
		return false
	}
	return m.Addr.Equals(other.Addr.Addr) && m.Description == other.Description &&
		slices.Equal(m.Versions, other.Versions) && m.IsBlocked == other.IsBlocked &&
		m.BlockedReason == other.BlockedReason && m.Popularity == other.Popularity && m.ForkCount == other.ForkCount &&
		m.ForkOfLink == other.ForkOfLink && m.ForkOf.Equals(other.ForkOf.Addr) &&
		m.UpstreamPopularity == other.UpstreamPopularity && m.UpstreamForkCount == other.UpstreamForkCount
}

// DeepCopy creates a deep copy of the module, ensuring that all new data structures are independent.
func (m *Module) DeepCopy() *Module {
	versions := make([]ModuleVersionDescriptor, len(m.Versions))
	copy(versions, m.Versions)

	return &Module{
		Addr:               m.Addr,
		Description:        m.Description,
		Versions:           versions,
		IsBlocked:          m.IsBlocked,
		BlockedReason:      m.BlockedReason,
		Popularity:         m.Popularity,
		ForkCount:          m.ForkCount,
		ForkOfLink:         m.ForkOfLink,
		ForkOf:             m.ForkOf,
		UpstreamPopularity: m.UpstreamPopularity,
		UpstreamForkCount:  m.UpstreamForkCount,
	}
}

func (m *Module) Validate() error {
	if err := m.Addr.Validate(); err != nil {
		return err
	}
	if len(m.Versions) == 0 {
		return fmt.Errorf("module with no versions (%s)", m.Addr)
	}
	for _, ver := range m.Versions {
		if err := ver.Validate(); err != nil {
			return fmt.Errorf("invalid version in module %s", m.Addr)
		}
	}
	return nil
}

// ModuleAddr describes a module address enriched with data for the API. Use the Addr() function
// to generate this from a module.Addr.
type ModuleAddr struct {
	module.Addr

	// Contains the display version of the addr presentable to the end user. This may be
	// capitalized.
	Display string `json:"display"`
	// Contains the namespace of the addr.
	Namespace string `json:"namespace"`
	// Contains the name of the addr.
	Name string `json:"name"`
	// Contains the target system of the addr.
	Target string `json:"target"`
}

func Addr(addr module.Addr) ModuleAddr {
	a := ModuleAddr{
		Addr: addr,
	}
	a.fill()
	return a
}

type marshalledAddr struct {
	Display   string `json:"display"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Target    string `json:"target"`
}

func (m *ModuleAddr) UnmarshalJSON(data []byte) error {
	marshalled := marshalledAddr{}
	if err := json.Unmarshal(data, &marshalled); err != nil {
		return err
	}
	m.Addr = module.Addr{Namespace: marshalled.Namespace, Name: marshalled.Name, TargetSystem: marshalled.Target}
	m.fill()
	return nil
}

func (m ModuleAddr) MarshalJSON() ([]byte, error) {
	result := marshalledAddr{
		Display:   m.Display,
		Namespace: m.Namespace,
		Name:      m.Name,
		Target:    m.Target,
	}

	result.Display = m.Display
	result.Namespace = m.Addr.Namespace
	result.Name = m.Addr.Name
	result.Target = m.Addr.TargetSystem

	return json.Marshal(result)
}

func (m *ModuleAddr) fill() {
	m.Display = m.Addr.String()
	m.Namespace = m.Addr.Namespace
	m.Name = m.Addr.Name
	m.Target = m.Addr.TargetSystem
}

func (m *Module) Compare(other Module) int {
	return m.Addr.Compare(other.Addr.Addr)
}

func (m *Module) HasVersion(version module.VersionNumber) bool {
	version = version.Normalize()
	for _, ver := range m.Versions {
		if ver.ID.Normalize() == version {
			return true
		}
	}
	return false
}

func (m *Module) AddVersions(versions ...ModuleVersionDescriptor) {
	if len(versions) == 0 {
		return
	}
	m.Versions = append(m.Versions, versions...)

	slices.SortStableFunc(m.Versions, func(a, b ModuleVersionDescriptor) int {
		return -a.ID.Compare(b.ID)
	})
}

func (m *Module) RemoveVersions(in module.VersionList, notIn module.VersionList) []module.VersionNumber {
	inVersionNumberMap := map[module.VersionNumber]struct{}{}
	notInVersionNumberMap := map[module.VersionNumber]struct{}{}
	for _, version := range in {
		inVersionNumberMap[version.Version.Normalize()] = struct{}{}
	}
	for _, version := range notIn {
		notInVersionNumberMap[version.Version.Normalize()] = struct{}{}
	}
	var removedVersions []module.VersionNumber

	var newVersions []ModuleVersionDescriptor
	for _, ver := range m.Versions {
		id := ver.ID.Normalize()
		_, notInOK := notInVersionNumberMap[id]
		_, inOK := inVersionNumberMap[id]
		if notInOK && !inOK {
			newVersions = append(newVersions, ver)
		} else {
			removedVersions = append(removedVersions, id)
		}
	}
	m.Versions = newVersions
	return removedVersions
}

func (m *Module) UpdateVersions(updatedVersions ...ModuleVersionDescriptor) {
	for _, updatedVersion := range updatedVersions {
		for i, existingVersion := range m.Versions {
			if existingVersion.ID.Compare(updatedVersion.ID) == 0 {
				m.Versions[i].Published = updatedVersion.Published
				m.Versions[i].ID = updatedVersion.ID
			}
		}
	}

}
