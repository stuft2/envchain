package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type vaultMock struct {
	srv           *httptest.Server
	LastToken     string
	LastNamespace string
	LastPath      string
}

func (v *vaultMock) URL() string { return v.srv.URL }
func (v *vaultMock) Close()      { v.srv.Close() }

func newVaultSuccess(t *testing.T, data map[string]any) *vaultMock {
	t.Helper()
	vm := &vaultMock{}
	vm.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vm.LastToken = r.Header.Get("X-Vault-Token")
		vm.LastNamespace = r.Header.Get("X-Vault-Namespace")
		vm.LastPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"data": data,
			},
		})
	}))
	return vm
}

func newVaultError(t *testing.T, status int, body string) *vaultMock {
	t.Helper()
	vm := &vaultMock{}
	vm.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vm.LastToken = r.Header.Get("X-Vault-Token")
		vm.LastNamespace = r.Header.Get("X-Vault-Namespace")
		vm.LastPath = r.URL.Path
		http.Error(w, body, status)
	}))
	return vm
}

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

func TestProviderInjectMissingToken(t *testing.T) {
	p := Provider{
		Context: context.Background(),
		Address: "https://vault.example",
		Path:    "v1/kv/data/app",
	}
	err := p.Inject()
	if err == nil || !strings.Contains(err.Error(), "no token") {
		t.Fatalf("want missing token error, got %v", err)
	}
}

func TestProviderInjectMissingPath(t *testing.T) {
	p := Provider{
		Context: context.Background(),
		Address: "https://vault.example",
		Token:   "t",
	}
	err := p.Inject()
	if err == nil || !strings.Contains(err.Error(), "no secret path") {
		t.Fatalf("want missing path error, got %v", err)
	}
}

func TestProviderInjectInvalidAddr(t *testing.T) {
	p := Provider{
		Context: context.Background(),
		Address: ":// bad url",
		Token:   "t",
		Path:    "v1/kv/data/app",
	}
	err := p.Inject()
	if err == nil || !strings.Contains(err.Error(), "invalid VAULT_ADDR") {
		t.Fatalf("want invalid addr error, got %v", err)
	}
}

func TestNewProviderReadsTokenFromHomeFile(t *testing.T) {
	const key = "ENV_TEST_VAULT_HOME"
	unsetEnv(t, key)

	vm := newVaultSuccess(t, map[string]any{key: "value"})
	defer vm.Close()

	t.Setenv("VAULT_ADDR", vm.URL())
	t.Setenv("VAULT_TOKEN", "")
	home := t.TempDir()
	writeFile(t, home, ".vault-token", "fromfile")
	t.Setenv("HOME", home)

	p := NewProvider("v1/kv/data/app")
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if vm.LastToken != "fromfile" {
		t.Fatalf("expected token from file, got %q", vm.LastToken)
	}
	if got := os.Getenv(key); got != "value" {
		t.Fatalf("%s=%q want %q", key, got, "value")
	}
}

func TestNewProviderPrefersEnvToken(t *testing.T) {
	const key = "ENV_TEST_VAULT_ENV_TOKEN"
	unsetEnv(t, key)

	vm := newVaultSuccess(t, map[string]any{key: "value"})
	defer vm.Close()

	t.Setenv("VAULT_ADDR", vm.URL())
	t.Setenv("VAULT_TOKEN", "envtoken")
	home := t.TempDir()
	writeFile(t, home, ".vault-token", "filetoken")
	t.Setenv("HOME", home)

	p := NewProvider("v1/kv/data/app")
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if vm.LastToken != "envtoken" {
		t.Fatalf("expected env token, got %q", vm.LastToken)
	}
	if got := os.Getenv(key); got != "value" {
		t.Fatalf("%s=%q want %q", key, got, "value")
	}
}

func TestNewProviderTrimsTokenFile(t *testing.T) {
	const key = "ENV_TEST_VAULT_TRIM_TOKEN"
	unsetEnv(t, key)

	vm := newVaultSuccess(t, map[string]any{key: "value"})
	defer vm.Close()

	t.Setenv("VAULT_ADDR", vm.URL())
	t.Setenv("VAULT_TOKEN", "")
	home := t.TempDir()
	writeFile(t, home, ".vault-token", " spaced \n")
	t.Setenv("HOME", home)

	p := NewProvider("v1/kv/data/app")
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if vm.LastToken != "spaced" {
		t.Fatalf("expected trimmed token, got %q", vm.LastToken)
	}
}

func TestProviderInjectSetsNamespaceHeader(t *testing.T) {
	const key = "ENV_TEST_VAULT_NAMESPACE"
	unsetEnv(t, key)

	vm := newVaultSuccess(t, map[string]any{key: "value"})
	defer vm.Close()

	t.Setenv("VAULT_ADDR", vm.URL())
	t.Setenv("VAULT_TOKEN", "t")
	t.Setenv("VAULT_NAMESPACE", "team/space")
	t.Setenv("HOME", t.TempDir())

	p := NewProvider("v1/kv/data/app")
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if vm.LastNamespace != "team/space" {
		t.Fatalf("expected namespace header, got %q", vm.LastNamespace)
	}
	if got := os.Getenv(key); got != "value" {
		t.Fatalf("%s=%q want %q", key, got, "value")
	}
}

func TestProviderInjectURLJoin(t *testing.T) {
	vm := newVaultSuccess(t, map[string]any{"K": "V"})
	defer vm.Close()

	t.Setenv("VAULT_ADDR", vm.URL()+"/base")
	t.Setenv("VAULT_TOKEN", "t")
	t.Setenv("HOME", t.TempDir())

	p := NewProvider("v1/kv/data/app")
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if got, want := vm.LastPath, "/base/v1/kv/data/app"; got != want {
		t.Fatalf("LastPath=%q want %q", got, want)
	}
}

func TestProviderInjectShorthandKVV2WithAddrEndingInV1(t *testing.T) {
	vm := newVaultSuccess(t, map[string]any{"K": "V"})
	defer vm.Close()

	t.Setenv("VAULT_ADDR", vm.URL()+"/base/v1")
	t.Setenv("VAULT_TOKEN", "t")
	t.Setenv("HOME", t.TempDir())

	p := NewProvider("kvv2/testing-center/dev/env-vars")
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if got, want := vm.LastPath, "/base/v1/kvv2/data/testing-center/dev/env-vars"; got != want {
		t.Fatalf("LastPath=%q want %q", got, want)
	}
}

func TestProviderInjectShorthandKVV2WithAddrMissingV1(t *testing.T) {
	vm := newVaultSuccess(t, map[string]any{"K": "V"})
	defer vm.Close()

	t.Setenv("VAULT_ADDR", vm.URL()+"/base")
	t.Setenv("VAULT_TOKEN", "t")
	t.Setenv("HOME", t.TempDir())

	p := NewProvider("kvv2/testing-center/dev/env-vars")
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if got, want := vm.LastPath, "/base/v1/kvv2/data/testing-center/dev/env-vars"; got != want {
		t.Fatalf("LastPath=%q want %q", got, want)
	}
}

func TestProviderInjectFlattensNonStringValues(t *testing.T) {
	const num = "ENV_TEST_VAULT_NUM"
	const boolean = "ENV_TEST_VAULT_BOOL"
	const obj = "ENV_TEST_VAULT_OBJ"
	unsetEnv(t, num)
	unsetEnv(t, boolean)
	unsetEnv(t, obj)

	vm := newVaultSuccess(t, map[string]any{
		num:     42,
		boolean: true,
		obj:     map[string]any{"x": "y"},
	})
	defer vm.Close()

	t.Setenv("VAULT_ADDR", vm.URL())
	t.Setenv("VAULT_TOKEN", "t")
	t.Setenv("HOME", t.TempDir())

	p := NewProvider("v1/kv/data/app")
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if got := os.Getenv(num); got != "42" {
		t.Fatalf("%s=%q want %q", num, got, "42")
	}
	if got := os.Getenv(boolean); got != "true" {
		t.Fatalf("%s=%q want %q", boolean, got, "true")
	}
	if got := os.Getenv(obj); !strings.Contains(got, `"x":"y"`) {
		t.Fatalf("%s=%q missing JSON content", obj, got)
	}
}

func TestProviderInjectRespectsExistingValues(t *testing.T) {
	const keep = "ENV_TEST_VAULT_KEEP"
	const add = "ENV_TEST_VAULT_ADD"

	t.Setenv(keep, "original")
	unsetEnv(t, add)

	vm := newVaultSuccess(t, map[string]any{
		keep: "fromvault",
		add:  "new",
	})
	defer vm.Close()

	t.Setenv("VAULT_ADDR", vm.URL())
	t.Setenv("VAULT_TOKEN", "t")
	t.Setenv("HOME", t.TempDir())

	p := NewProvider("v1/kv/data/app")
	if err := p.Inject(); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if got := os.Getenv(keep); got != "original" {
		t.Fatalf("%s=%q want %q", keep, got, "original")
	}
	if got := os.Getenv(add); got != "new" {
		t.Fatalf("%s=%q want %q", add, got, "new")
	}
}

func TestProviderInjectHTTPError(t *testing.T) {
	vm := newVaultError(t, http.StatusForbidden, "nope")
	defer vm.Close()

	t.Setenv("VAULT_ADDR", vm.URL())
	t.Setenv("VAULT_TOKEN", "t")
	t.Setenv("HOME", t.TempDir())

	p := NewProvider("v1/kv/data/app")
	err := p.Inject()
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("want forbidden error, got %v", err)
	}
}

func TestProviderInjectInvalidJSON(t *testing.T) {
	vm := &vaultMock{}
	vm.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vm.LastPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{not-json"))
	}))
	defer vm.Close()

	t.Setenv("VAULT_ADDR", vm.URL())
	t.Setenv("VAULT_TOKEN", "t")
	t.Setenv("HOME", t.TempDir())

	p := NewProvider("v1/kv/data/app")
	err := p.Inject()
	if err == nil || !strings.Contains(err.Error(), "decode body") {
		t.Fatalf("want decode body error, got %v", err)
	}
}

func TestProviderInjectContextCancelled(t *testing.T) {
	vm := newVaultSuccess(t, map[string]any{"K": "V"})
	defer vm.Close()

	p := Provider{
		Address: vm.URL(),
		Token:   "t",
		Path:    "v1/kv/data/app",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := p.InjectContext(ctx)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("want context canceled error, got %v", err)
	}
}
