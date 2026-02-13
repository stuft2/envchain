package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/stuft2/envault"
	"github.com/stuft2/envault/internal"
	"github.com/stuft2/envault/providers/dotenv"
	"github.com/stuft2/envault/providers/vault"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("envault", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	dotenvPath := fs.String("dotenv", ".env", "path to a dotenv file (empty to skip)")
	vaultPath := fs.String("vault-path", "", "Vault KV v2 secret path to read (empty to skip)")
	verbose := fs.Bool("verbose", false, "enable verbose logging")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: envault [flags] -- command [args...]\n\n")
		fmt.Fprintf(fs.Output(), "Flags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	rest := fs.Args()
	if len(rest) == 0 {
		fs.Usage()
		return 2
	}

	if *verbose {
		logger := log.New(os.Stderr, "envault: ", log.LstdFlags)
		internal.SetLogger(logger)
		internal.Debugf("verbose logging enabled")
	}

	providers := gatherProviders(*dotenvPath, *vaultPath)
	if err := envault.Inject(providers...); err != nil {
		fmt.Fprintf(os.Stderr, "envault: %v\n", err)
		return 1
	}

	cmd := exec.Command(rest[0], rest[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "envault: failed to execute %q: %v\n", rest[0], err)
		return 1
	}

	return 0
}

func gatherProviders(dotenvPath, vaultPath string) []internal.Provider {
	var providers []internal.Provider
	if dotenvPath != "" {
		providers = append(providers, dotenv.NewProvider(dotenvPath))
	}
	if vaultPath != "" {
		providers = append(providers, vault.NewProvider(vaultPath))
	}
	return providers
}
