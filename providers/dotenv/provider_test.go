package dotenv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	fp := filepath.Join(dir, name)
	if err := os.WriteFile(fp, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", fp, err)
	}
	return fp
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
}

func TestProviderInjectLoadsEnvFile(t *testing.T) {
	const key = "ENV_TEST_DOTENV_LOAD"
	unsetEnv(t, key)

	tmp := t.TempDir()
	path := writeFile(t, tmp, ".env", fmt.Sprintf("%s=value\n", key))

	p := NewProvider(path)
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if got := os.Getenv(key); got != "value" {
		t.Fatalf("%s=%q want %q", key, got, "value")
	}
}

func TestProviderInjectUsesDefaultPath(t *testing.T) {
	const key = "ENV_TEST_DOTENV_DEFAULT_PATH"
	unsetEnv(t, key)

	tmp := t.TempDir()
	writeFile(t, tmp, ".env", fmt.Sprintf("%s=ok\n", key))

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	p := NewProvider("")
	if want := "./.env"; p.Path != want {
		t.Fatalf("Path=%q want %q", p.Path, want)
	}
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if got := os.Getenv(key); got != "ok" {
		t.Fatalf("%s=%q want %q", key, got, "ok")
	}
}

func TestProviderInjectMissingFileIsNoop(t *testing.T) {
	const key = "ENV_TEST_DOTENV_MISSING"
	unsetEnv(t, key)

	p := NewProvider("/no/such/file.env")
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if _, ok := os.LookupEnv(key); ok {
		t.Fatalf("unexpected env var %s present", key)
	}
}

func TestProviderInjectReadError(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "as-dir.env")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}

	p := NewProvider(dir)
	err := p.Inject()
	if err == nil || !strings.Contains(err.Error(), "read") {
		t.Fatalf("want read error, got %v", err)
	}
}

func TestProviderInjectDoesNotOverwriteExisting(t *testing.T) {
	const keep = "ENV_TEST_DOTENV_KEEP"
	const add = "ENV_TEST_DOTENV_ADD"

	t.Setenv(keep, "original")
	unsetEnv(t, add)

	tmp := t.TempDir()
	content := fmt.Sprintf("%s=fromdotenv\n%s=new\n", keep, add)
	path := writeFile(t, tmp, ".env", content)

	p := NewProvider(path)
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if got := os.Getenv(keep); got != "original" {
		t.Fatalf("%s=%q want %q (no overwrite)", keep, got, "original")
	}
	if got := os.Getenv(add); got != "new" {
		t.Fatalf("%s=%q want %q", add, got, "new")
	}
}
