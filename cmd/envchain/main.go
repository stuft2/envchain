package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/stuft2/envchain"
	"github.com/stuft2/envchain/internal"
	"github.com/stuft2/envchain/providers/dotenv"
	"github.com/stuft2/envchain/providers/vault"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	return runWithDeps(args, os.Stdin, os.Stdout, os.Stderr, envchain.Inject, gatherProviders, defaultCommandExecutor)
}

type injectFunc func(...internal.Provider) error
type gatherProvidersFunc func(dotenvPath, vaultPath string) []internal.Provider
type commandExecutorFunc func(name string, args []string, stdin io.Reader, stdout, stderr io.Writer) (int, error)

func runWithDeps(
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	inject injectFunc,
	gather gatherProvidersFunc,
	executeCommand commandExecutorFunc,
) int {
	fs := flag.NewFlagSet("envchain", flag.ContinueOnError)
	fs.SetOutput(stderr)

	dotenvPath := fs.String("dotenv", ".env", "path to a dotenv file (empty to skip)")
	vaultPath := fs.String("vault-path", "", "Vault KV v2 secret path to read (empty to skip)")
	verbose := fs.Bool("verbose", false, "enable verbose logging")

	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: envchain [flags] -- command [args...]")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "Flags:")
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
		logger := log.New(stderr, "envchain: ", log.LstdFlags)
		internal.SetLogger(logger)
		internal.Debugf("verbose logging enabled")
	}

	providers := gather(*dotenvPath, *vaultPath)
	if err := inject(providers...); err != nil {
		fmt.Fprintf(stderr, "envchain: %v\n", err)
		return 1
	}

	exitCode, err := executeCommand(rest[0], rest[1:], stdin, stdout, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "envchain: failed to execute %q: %v\n", rest[0], err)
		return 1
	}
	return exitCode
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

func defaultCommandExecutor(name string, args []string, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return 0, err
	}
	return 0, nil
}
