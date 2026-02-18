package envchain

import (
	"context"
	"errors"

	"github.com/stuft2/envchain/internal"
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

func InjectWithContext(ctx context.Context, providers ...internal.Provider) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var errs []error
	for _, provider := range providers {
		internal.Debugf("injecting provider %T with context", provider)

		var err error
		if contextProvider, ok := provider.(internal.ContextProvider); ok {
			err = contextProvider.InjectContext(ctx)
		} else {
			err = provider.Inject()
		}

		if err != nil {
			internal.Debugf("provider %T returned error: %v", provider, err)
			errs = append(errs, err)
		} else {
			internal.Debugf("provider %T finished", provider)
		}
	}

	// Returns nil if errs list is empty or all are entries nil
	return errors.Join(errs...)
}
