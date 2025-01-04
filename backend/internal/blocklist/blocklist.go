package blocklist

import (
	"encoding/json"
	"os"

	"github.com/opentofu/libregistry/types/module"
	"github.com/opentofu/libregistry/types/provider"
)

func New() BlockList {
	return &blockList{
		Providers: map[provider.Addr]string{},
		Modules:   map[module.Addr]string{},
	}
}

type BlockList interface {
	// LoadFile loads a blocklist from a file.
	LoadFile(file string) error
	// IsModuleBlocked returns true if the module is blocked in addition with a reason why it is blocked.
	IsModuleBlocked(addr module.Addr) (bool, string)
	// IsProviderBlocked returns true if the provider is blocked in addition with a reason why it is blocked.
	IsProviderBlocked(addr provider.Addr) (bool, string)
}

// BlockList describes a dataset constraining provider or module documentation from being shown. It has the same
// effect as if there was an incompatible license.
type blockList struct {
	// Providers holds a map of provider addresses and the explanation why it was blocked. The addr may contain only
	// a namespace, in which case the entire namespace is blocked.
	// Do not interrogate this field directly, use IsProviderBlocked() instead.
	Providers blockListType[provider.Addr] `json:"providers"`
	// Modules holds a map of module addresses and the explanation why it was blocked. The addr may contain only
	//	// a namespace, in which case the entire namespace is blocked.
	// Do not interrogate this field directly, use IsModuleBlocked() instead.
	Modules blockListType[module.Addr] `json:"modules"`
}

type blockListTypeType interface {
	provider.Addr | module.Addr
}

type blockListType[T blockListTypeType] map[T]string

func (b *blockListType[T]) UnmarshalJSON(data []byte) error {
	tmp := map[string]string{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	for k, v := range tmp {
		marshalledK, err := json.Marshal(k)
		if err != nil {
			return err
		}
		var typedK T
		if err := json.Unmarshal(marshalledK, &typedK); err != nil {
			return err
		}
		(*b)[typedK] = v
	}
	return nil
}

func (b *blockList) LoadFile(file string) error {
	blockListContents, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(blockListContents, &b); err != nil {
		return err
	}
	return nil
}

// IsModuleBlocked returns true if the module is blocked in addition with a reason why it is blocked.
// Use this function instead of querying the struct directly for future compatibility.
func (b *blockList) IsModuleBlocked(addr module.Addr) (bool, string) {
	reason, ok := b.Modules[addr]
	if ok {
		return ok, reason
	}
	reason, ok = b.Modules[module.Addr{Namespace: addr.Namespace}]
	return ok, reason
}

// IsProviderBlocked returns true if the provider is blocked in addition with a reason why it is blocked.
// Use this function instead of querying the struct directly for future compatibility.
func (b *blockList) IsProviderBlocked(addr provider.Addr) (bool, string) {
	reason, ok := b.Providers[addr]
	if ok {
		return ok, reason
	}
	reason, ok = b.Providers[provider.Addr{Namespace: addr.Namespace, Name: addr.Name}]
	if ok {
		return ok, reason
	}
	reason, ok = b.Providers[provider.Addr{Namespace: addr.Namespace}]
	return ok, reason
}
