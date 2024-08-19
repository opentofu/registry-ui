package moduleindex

import (
	"time"

	"github.com/opentofu/libregistry/types/module"
)

// ModuleVersionDescriptor describes a single version.
//
// swagger:model
type ModuleVersionDescriptor struct {
	// required: true
	ID module.VersionNumber `json:"id"`
	// required: true
	Published time.Time `json:"published"`
}

func (d ModuleVersionDescriptor) Validate() error {
	if err := d.ID.Validate(); err != nil {
		return err
	}
	return nil
}
