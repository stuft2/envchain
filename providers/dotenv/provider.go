package dotenv

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/joho/godotenv"
	"github.com/stuft2/envault/internal"
)

type Provider struct {
	Path string
}

func NewProvider(dotenvPath string) Provider {
	if dotenvPath == "" {
		dotenvPath = "./.env"
	}
	return Provider{
		Path: dotenvPath,
	}
}

func (p Provider) Inject() error {
	internal.Debugf("dotenv: reading %s", p.Path)
	b, err := os.ReadFile(p.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			internal.Debugf("dotenv: %s not found", p.Path)
			return nil
		}
		return fmt.Errorf("read %q: %w", p.Path, err)
	}
	m, err := godotenv.Unmarshal(string(b))
	if err != nil {
		return fmt.Errorf("parse %q as dotenv: %w", p.Path, err)
	}
	internal.Debugf("dotenv: loaded %d variables from %s", len(m), p.Path)
	err = internal.SetEnvMap(m)
	if err != nil {
		return fmt.Errorf("cannot set env vars provided by .env (%q): %w", p.Path, err)
	}
	internal.Debugf("dotenv: finished applying variables from %s", p.Path)
	return nil
}

var _ internal.Provider = (*Provider)(nil)
