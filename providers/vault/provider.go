package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/stuft2/envault/internal"
)

type Provider struct {
	Context   context.Context
	Address   string
	Token     string
	Path      string
	Namespace string
}

func NewProvider(vaultPath string) Provider {
	p := Provider{}

	// First, we check for the vault address. If this is provided, then we assume the user wants us to pull from vault.
	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		return p
	}
	p.Address = addr

	// To pull from vault, we need a vault token
	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		if home, _ := os.UserHomeDir(); home != "" {
			if b, err := os.ReadFile(filepath.Join(home, ".vault-token")); err == nil {
				token = strings.TrimSpace(string(b))
			}
		}
	}
	if token == "" {
		return p
	}
	p.Token = token

	// The last required piece is the path that it should look at to pull secrets
	if vaultPath == "" {
		return p
	}
	p.Path = normalizeSecretPath(addr, vaultPath)

	// Optionally, a user may not be using the default namespace for their secrets so we will check for that
	p.Namespace = os.Getenv("VAULT_NAMESPACE")

	// Set the background context
	p.Context = context.Background()

	return p
}

func (p Provider) Inject() error {
	ctx := p.Context
	if ctx == nil {
		ctx = context.Background()
	}
	return p.injectContext(ctx)
}

func (p Provider) InjectContext(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return p.injectContext(ctx)
}

func (p Provider) injectContext(ctx context.Context) error {
	internal.Debugf("vault: starting injection (addr=%q, path=%q, namespace=%q)", p.Address, p.Path, p.Namespace)
	if p.Address == "" {
		return fmt.Errorf("vault: VAULT_ADDR is required to inject environment variables")
	}
	if p.Token == "" {
		return fmt.Errorf("vault: VAULT_ADDR set but no token found (VAULT_TOKEN or ~/.vault-token)")
	}
	if p.Path == "" {
		return fmt.Errorf("vault: VAULT_ADDR set but no secret path provided (vault.Provider.Path is empty)")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	u, err := url.Parse(p.Address)
	if err != nil {
		return fmt.Errorf("vault: invalid VAULT_ADDR %q: %w", p.Address, err)
	}
	joined := path.Join(u.Path, p.Path) // append the secret path
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	u.Path = joined
	fullURL := u.String()
	internal.Debugf("vault: requesting %s", fullURL)

	// NewRequestWithContext fails only if the method is empty or the URL is invalid.
	// (e.g. malformed scheme, missing scheme, bad host/port). A canceled context
	// or a nil body will not cause an error at construction time.
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	req.Header.Set("X-Vault-Token", p.Token)
	if p.Namespace != "" {
		// support vault namespacing, otherwise uses vault's default p.Namespace
		req.Header.Set("X-Vault-Namespace", p.Namespace)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("vault: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("vault: %s (failed to read response body: %v)", resp.Status, err)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("vault: %s\n%s", resp.Status, string(body))
	}

	var out struct {
		Data struct {
			Data map[string]any `json:"data"`
		} `json:"data"`
	}
	if err = json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("vault: decode body: %w", err)
	}

	flat := make(map[string]string, len(out.Data.Data))
	for k, v := range out.Data.Data {
		if s, ok := v.(string); ok {
			flat[k] = s
		} else {
			b, _ := json.Marshal(v)
			flat[k] = string(b)
		}
	}
	internal.Debugf("vault: received %d variables from %s", len(flat), fullURL)

	err = internal.SetEnvMap(flat)
	if err != nil {
		return fmt.Errorf("cannot set env vars provided by vault (%q): %w", fullURL, err)
	}
	internal.Debugf("vault: finished applying variables from %s", fullURL)

	return nil
}

var _ internal.Provider = (*Provider)(nil)
var _ internal.ContextProvider = (*Provider)(nil)

// normalizeSecretPath lets users provide a shorthand kvv2/<service>/... path and
// ensures Vault is still queried at v1/kvv2/data/... depending on VAULT_ADDR.
func normalizeSecretPath(addr, requested string) string {
	if requested == "" {
		return requested
	}
	originalHadSlash := strings.HasPrefix(requested, "/")
	sanitized := strings.TrimPrefix(requested, "/")
	if !strings.HasPrefix(sanitized, "kvv2/") {
		return requested
	}
	suffix := strings.TrimPrefix(sanitized, "kvv2/")
	base := path.Join("v1", "kvv2", "data")
	trimmedAddr := strings.TrimRight(addr, "/")
	if strings.HasSuffix(trimmedAddr, "/v1") {
		base = path.Join("kvv2", "data")
	}
	rewritten := path.Join(base, suffix)
	if originalHadSlash {
		return "/" + rewritten
	}
	return rewritten
}
