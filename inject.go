package envault

import (
	"errors"

	"github.com/stuft2/envault/internal"
)

func Inject(providers ...internal.Provider) error {
	var errs []error
	for _, provider := range providers {
		internal.Debugf("injecting provider %T", provider)
		if err := provider.Inject(); err != nil {
			internal.Debugf("provider %T returned error: %v", provider, err)
			errs = append(errs, err)
		} else {
			internal.Debugf("provider %T finished", provider)
		}
	}
	// Returns nil if errs list is empty or all are entries nil
	return errors.Join(errs...)
}
