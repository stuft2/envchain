package internal

import (
	"context"
	"fmt"
	"os"
)

type Provider interface {
	Inject() error
}

type ContextProvider interface {
	InjectContext(context.Context) error
}

func SetEnvMap(vars map[string]string) error {
	for key, value := range vars {
		if _, ok := os.LookupEnv(key); ok {
			Debugf("environment variable %s already set", key)
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
		Debugf("environment variable %s set to %v", key, value)
	}
	return nil
}
