package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stuft2/envault"
	"github.com/stuft2/envault/internal"
	"github.com/stuft2/envault/providers/dotenv"
	"github.com/stuft2/envault/providers/vault"
)

func TestRunHelpReturnsZeroAndUsage(t *testing.T) {
	t.Cleanup(func() { internal.SetLogger(nil) })

	var stderr bytes.Buffer
	code := runWithDeps(
		[]string{"-h"},
		strings.NewReader(""),
		&bytes.Buffer{},
		&stderr,
		func(...internal.Provider) error { return nil },
		func(_, _ string) []internal.Provider { return nil },
		func(string, []string, io.Reader, io.Writer, io.Writer) (int, error) { return 0, nil },
	)

	if code != 0 {
		t.Fatalf("runWithDeps code=%d want 0", code)
	}
	if got := stderr.String(); !strings.Contains(got, "Usage: envault [flags] -- command [args...]") {
		t.Fatalf("usage output missing; got %q", got)
	}
}

func TestRunMissingCommandReturnsTwo(t *testing.T) {
	t.Cleanup(func() { internal.SetLogger(nil) })

	var stderr bytes.Buffer
	code := runWithDeps(
		nil,
		strings.NewReader(""),
		&bytes.Buffer{},
		&stderr,
		func(...internal.Provider) error { return nil },
		func(_, _ string) []internal.Provider { return nil },
		func(string, []string, io.Reader, io.Writer, io.Writer) (int, error) { return 0, nil },
	)
	if code != 2 {
		t.Fatalf("runWithDeps code=%d want 2", code)
	}
}

func TestRunWrappedCommandExitCodePassthrough(t *testing.T) {
	t.Cleanup(func() { internal.SetLogger(nil) })

	var stderr bytes.Buffer
	code := runWithDeps(
		[]string{"--", "anything"},
		strings.NewReader(""),
		&bytes.Buffer{},
		&stderr,
		func(...internal.Provider) error { return nil },
		func(_, _ string) []internal.Provider { return nil },
		func(string, []string, io.Reader, io.Writer, io.Writer) (int, error) { return 7, nil },
	)
	if code != 7 {
		t.Fatalf("runWithDeps code=%d want 7", code)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr=%q want empty", got)
	}
}

func TestRunNonExistentCommandReturnsExecutionError(t *testing.T) {
	t.Cleanup(func() { internal.SetLogger(nil) })

	var stderr bytes.Buffer
	code := runWithDeps(
		[]string{"--", "definitely-not-a-real-command-envault-test"},
		strings.NewReader(""),
		&bytes.Buffer{},
		&stderr,
		func(...internal.Provider) error { return nil },
		func(_, _ string) []internal.Provider { return nil },
		defaultCommandExecutor,
	)
	if code != 1 {
		t.Fatalf("runWithDeps code=%d want 1", code)
	}
	msg := stderr.String()
	if !strings.Contains(msg, `failed to execute "definitely-not-a-real-command-envault-test"`) {
		t.Fatalf("stderr missing execution error message; got %q", msg)
	}
}

func TestRunDotenvEmptyDisablesDotenvProvider(t *testing.T) {
	t.Cleanup(func() { internal.SetLogger(nil) })

	var got []internal.Provider
	inject := func(providers ...internal.Provider) error {
		got = providers
		return nil
	}

	code := runWithDeps(
		[]string{"-dotenv", "", "--", "echo"},
		strings.NewReader(""),
		&bytes.Buffer{},
		&bytes.Buffer{},
		inject,
		gatherProviders,
		func(string, []string, io.Reader, io.Writer, io.Writer) (int, error) { return 0, nil },
	)
	if code != 0 {
		t.Fatalf("runWithDeps code=%d want 0", code)
	}
	if len(got) != 0 {
		t.Fatalf("providers len=%d want 0", len(got))
	}
}

func TestGatherProvidersVaultPathAddsProvider(t *testing.T) {
	providers := gatherProviders("", "kvv2/testing-center/dev/env-vars")
	if len(providers) != 1 {
		t.Fatalf("providers len=%d want 1", len(providers))
	}
	if _, ok := providers[0].(vault.Provider); !ok {
		t.Fatalf("provider type=%T want vault.Provider", providers[0])
	}
}

func TestRunVerboseEnablesDebugLoggingWithoutSecretLeakage(t *testing.T) {
	t.Cleanup(func() { internal.SetLogger(nil) })

	tmp := t.TempDir()
	dotenvPath := fmt.Sprintf("%s/.env", tmp)
	if err := os.WriteFile(dotenvPath, []byte("TEST_VERBOSE_SECRET=supersecret\n"), 0o600); err != nil {
		t.Fatalf("write dotenv file: %v", err)
	}

	var stderr bytes.Buffer
	code := runWithDeps(
		[]string{"-verbose", "-dotenv", dotenvPath, "--", "echo"},
		strings.NewReader(""),
		&bytes.Buffer{},
		&stderr,
		envault.Inject,
		gatherProviders,
		func(string, []string, io.Reader, io.Writer, io.Writer) (int, error) { return 0, nil },
	)
	if code != 0 {
		t.Fatalf("runWithDeps code=%d want 0", code)
	}

	logs := stderr.String()
	if !strings.Contains(logs, "verbose logging enabled") {
		t.Fatalf("expected verbose debug log; got %q", logs)
	}
	if strings.Contains(logs, "supersecret") {
		t.Fatalf("secret value leaked in logs: %q", logs)
	}
}

func TestRunProviderErrorReturnsOne(t *testing.T) {
	t.Cleanup(func() { internal.SetLogger(nil) })

	var stderr bytes.Buffer
	code := runWithDeps(
		[]string{"--", "echo"},
		strings.NewReader(""),
		&bytes.Buffer{},
		&stderr,
		func(...internal.Provider) error { return errors.New("inject failed") },
		func(_, _ string) []internal.Provider { return []internal.Provider{dotenv.NewProvider(".env")} },
		func(string, []string, io.Reader, io.Writer, io.Writer) (int, error) { return 0, nil },
	)
	if code != 1 {
		t.Fatalf("runWithDeps code=%d want 1", code)
	}
	if got := stderr.String(); !strings.Contains(got, "envault: inject failed") {
		t.Fatalf("stderr missing provider failure; got %q", got)
	}
}

func TestDefaultCommandExecutorReturnsExitCode(t *testing.T) {
	code, err := defaultCommandExecutor("sh", []string{"-c", "exit 23"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("defaultCommandExecutor err=%v", err)
	}
	if code != 23 {
		t.Fatalf("defaultCommandExecutor code=%d want 23", code)
	}
}

func TestRunPassesDotenvAndVaultFlagsToProviderGathering(t *testing.T) {
	t.Cleanup(func() { internal.SetLogger(nil) })

	var gotDotenvPath, gotVaultPath string
	gather := func(dotenvPath, vaultPath string) []internal.Provider {
		gotDotenvPath = dotenvPath
		gotVaultPath = vaultPath
		return nil
	}

	code := runWithDeps(
		[]string{"-dotenv", ".env.local", "-vault-path", "kvv2/app/dev/env", "--", "echo"},
		strings.NewReader(""),
		&bytes.Buffer{},
		&bytes.Buffer{},
		func(...internal.Provider) error { return nil },
		gather,
		func(string, []string, io.Reader, io.Writer, io.Writer) (int, error) { return 0, nil },
	)
	if code != 0 {
		t.Fatalf("runWithDeps code=%d want 0", code)
	}
	if gotDotenvPath != ".env.local" || gotVaultPath != "kvv2/app/dev/env" {
		t.Fatalf("gather args=(%q,%q) want (%q,%q)", gotDotenvPath, gotVaultPath, ".env.local", "kvv2/app/dev/env")
	}
}

func TestGatherProvidersIncludesDotenvWhenEnabled(t *testing.T) {
	providers := gatherProviders(".env", "")
	if len(providers) != 1 {
		t.Fatalf("providers len=%d want 1", len(providers))
	}
	if _, ok := providers[0].(dotenv.Provider); !ok {
		t.Fatalf("provider type=%T want dotenv.Provider", providers[0])
	}
}

func TestRunInvalidFlagReturnsTwo(t *testing.T) {
	t.Cleanup(func() { internal.SetLogger(nil) })

	var stderr bytes.Buffer
	code := runWithDeps(
		[]string{"-nope"},
		strings.NewReader(""),
		&bytes.Buffer{},
		&stderr,
		func(...internal.Provider) error { return nil },
		func(_, _ string) []internal.Provider { return nil },
		func(string, []string, io.Reader, io.Writer, io.Writer) (int, error) { return 0, nil },
	)
	if code != 2 {
		t.Fatalf("runWithDeps code=%d want 2", code)
	}
	if got := stderr.String(); !strings.Contains(got, "flag provided but not defined") {
		t.Fatalf("stderr missing flag parse error; got %q", got)
	}
}

func TestGatherProvidersOrderDotenvThenVault(t *testing.T) {
	providers := gatherProviders(".env", "kvv2/service/dev/env-vars")
	if len(providers) != 2 {
		t.Fatalf("providers len=%d want 2", len(providers))
	}

	gotTypes := []string{reflect.TypeOf(providers[0]).String(), reflect.TypeOf(providers[1]).String()}
	wantTypes := []string{"dotenv.Provider", "vault.Provider"}
	if !reflect.DeepEqual(gotTypes, wantTypes) {
		t.Fatalf("provider order=%v want %v", gotTypes, wantTypes)
	}
}
