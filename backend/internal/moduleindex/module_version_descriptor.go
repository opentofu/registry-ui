package moduleindex

import (
	"time"

	"github.com/opentofu/libregistry/types/module"
)

// ModuleVersionDescriptor describes a single version.
type ModuleVersionDescriptor struct {
	ID module.VersionNumber `json:"id"`
	Published time.Time `json:"published"`
}

func (d ModuleVersionDescriptor) Validate() error {
	if err := d.ID.Validate(); err != nil {
		return err
	}
	return nil
}
