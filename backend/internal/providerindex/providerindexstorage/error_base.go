package providerindexstorage

import (
	"fmt"
)

type BaseError struct {
	Cause error
}

func (b BaseError) Unwrap() error {
	return b.Cause
}

func (b BaseError) Error() string {
	if b.Cause != nil {
		return b.Cause.Error()
	}
	return fmt.Sprintf("Unspecified %T", b)
}
