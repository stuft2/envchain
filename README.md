# envault

## Overview

`envault` is a tiny helper for Go services that **backfills** environment variables from one or more **providers** — without overwriting anything that’s already set. It also provides a small helper, `GetEnvOrDefault`, for reading environment variables with defaults.

#### Providers Available:

- **Dotenv**: read a local `.env` file
- **HashiCorp Vault** (KV v2): fetch secrets over HTTP using a Vault token

Order matters and establishes precedence: **existing process env** ➜ **first provider** ➜ **second provider** ➜ ...

Anything already set in the process wins. Missing keys are filled by the first provider you pass; still-missing keys are filled by the next provider, and so on.

---

## Quickstart (1 minute)

1. Install the CLI:
   ```bash
   go install github.com/stuft2/envault/cmd/envchain@latest
   ```
1. Create a sample `.env`:
   ```bash
   cat > .env <<'EOF'
   APP_NAME=envault-demo
   PORT=8080
   EOF
   ```
1. Run a command with backfilled env vars:
   ```bash
   envchain -- env | grep -E '^(APP_NAME|PORT)='
   ```
1. Verify output:
   ```text
   APP_NAME=envault-demo
   PORT=8080
   ```

Optional Vault quickstart (requires Vault access):

```bash
export VAULT_ADDR="https://vault.byu.edu"
export VAULT_TOKEN="<your-token>" # or use ~/.vault-token
envchain -vault-path "kvv2/<service>/dev/env-vars" -- env | grep '^YOUR_KEY='
```

If `YOUR_KEY` is unset locally, `envchain` backfills it from Vault.

## Troubleshooting

Symptom: `vault: VAULT_ADDR is required to inject environment variables`  
Cause: `-vault-path` is set, but `VAULT_ADDR` is missing.  
Fix: export `VAULT_ADDR` (for example, `https://vault.byu.edu`) or remove `-vault-path`.

Symptom: `vault: VAULT_ADDR set but no token found (VAULT_TOKEN or ~/.vault-token)`  
Cause: Vault address is configured, but auth token is unavailable.  
Fix: set `VAULT_TOKEN` or run `vault login` so `~/.vault-token` exists.

Symptom: `vault: invalid VAULT_ADDR "...": ...`  
Cause: `VAULT_ADDR` is malformed (missing scheme, invalid host, or invalid URL).  
Fix: use a valid URL such as `https://vault.byu.edu` (or `https://vault.byu.edu/v1`).

Symptom: `Usage: envault [flags] -- command [args...]`  
Cause: command was not passed after `--`.  
Fix: provide a command after separator, for example `envchain -- env`.

Symptom: `envault: failed to execute "..."`  
Cause: the command after `--` is missing from `PATH` or not executable.  
Fix: verify the executable name and run `which <command>` to confirm availability.

Symptom: expected value from `.env`/Vault is not applied  
Cause: existing process env takes precedence over providers.  
Fix: unset the key before running, for example:

```bash
export PORT=3000
envchain -- env | grep '^PORT='    # PORT=3000 (existing env wins)
unset PORT
envchain -- env | grep '^PORT='    # PORT from .env or Vault
```

Verbose mode note: `-verbose` logs provider flow and key names, not secret values. Secret-safe diagnostics and redaction guarantees will continue to improve as diagnostics features evolve.

---

## How to Use This

### Injecting Environment Variables

Construct the providers you want and pass them to `envault.Inject` in **precedence order**:

```go
package main

import (
    "log"

    "github.com/stuft2/envault"
    "github.com/stuft2/envault/providers/dotenv"
    "github.com/stuft2/envault/providers/vault"
)

func main() {
	// Precedence: keep existing env, then .env, then Vault. 
	if err := envault.Inject(dotenv.NewProvider(".env"), vault.NewProvider("kvv2/byuapi-persons-v4/dev/env-vars")); err != nil {
	    // could make provider errors fatal, but we're assuming that deployed environments
	    // will always have config and secrets injected before server start.
	    log.Printf("env injection warnings: %v", err)
	}

	// ... retrieve env vars with os.GetEnv or envault.GetEnvOrDefault
}
```

You can include any number of providers; only unset keys are written.

### Vault provider requirements

The Vault provider is active when these are satisfied:
 - VAULT_ADDR — base URL for Vault (e.g., https://vault.byu.edu or https://vault.byu.edu/v1)
 - A token:
   - VAULT_TOKEN, or
   - ~/.vault-token (created by vault login)
 - Secret path you pass to vault.NewProvider(...) (KV v2 path, e.g. `kvv2/<service-name>/dev/env-vars`)
 - Optional: VAULT_NAMESPACE → sent as X-Vault-Namespace

Timeouts: Vault HTTP requests use a 10s timeout.

Context: The Vault provider uses a background context by default. To override:

```go
p := vault.NewProvider("/app/web")
p.Context = ctx // set deadlines, cancellation, etc.
if err := envault.Inject(p); err != nil { /* ... */ }
```

Or pass one shared context across all context-aware providers:

```go
if err := envault.InjectWithContext(ctx, dotenv.NewProvider(".env"), vault.NewProvider("/app/web")); err != nil {
	// handle joined provider errors
}
```

### Reading Environment Variables with Defaults

Instead of manually checking for missing values, use `GetEnvOrDefault`:

```go
// GetEnvOrDefault returns the value of an environment variable if set,
// otherwise it returns the provided default.
func GetEnvOrDefault(key, def string) string {
    if val, ok := os.LookupEnv(key); ok {
        return val
    }
    return def
}
```

This lets you simplify configuration code:

```go
package main

import (
	"fmt"
	"os"

	"github.com/stuft2/envault"
)

func main() {
	// Example: PORT will default to 8080 if not set.
	port := envault.GetEnvOrDefault("PORT", "8080")
	addr := envault.GetEnvOrDefault("ADDR", ":http")

	fmt.Println("Starting server on", addr, "port", port)

	// Example with empty string (treated as set):
	_ = os.Setenv("DEBUG", "")
	debug := envault.GetEnvOrDefault("DEBUG", "false")
	fmt.Println("Debug mode =", debug)
}
```

---

## Recommended usage
 - Local dev:
   1. Put a `.env` beside your app
   2. Set `VAULT_ADDR` and authenticate (`vault login`) to populate missing secrets.
     ```shell
     VAULT_ADDR=https://vault.byu.edu vault login -method=oidc -path=byu-sso
     ```
 - CI/Prod: rely on process env only. If you don’t set `VAULT_ADDR`, the Vault provider effectively becomes inert, and a missing `.env` is ignored.

### CLI helper

Instead of wiring `envault.Inject` into your code, you can wrap any command with the CLI:

```bash
go run ./cmd/envchain -- <command> [args...]
```

Install it once and reuse it anywhere:

```bash
# recommended
go install ./cmd/envchain
# or without cloning the repo
go install github.com/stuft2/envault/cmd/envchain@latest
```

Flags:

- `-dotenv` (default `.env`): path to a dotenv file to backfill from. Pass an empty string to skip it.
- `-vault-path`: KV v2 path to load from HashiCorp Vault. Leave empty to skip Vault.
- `-verbose`: emit debug logs detailing how each provider resolves environment variables.

Environment variables such as `VAULT_ADDR`, `VAULT_TOKEN`, and `VAULT_NAMESPACE` still control Vault behaviour.

---

## External Dependencies

* [joho/godotenv](https://github.com/joho/godotenv) — parse `.env` files.

## CI / Local Checks

CI runs on push and pull request via `.github/workflows/ci.yml`.

Run the same checks locally:

```bash
task fmt
task lint
task test
```

## Project Docs

- Usability feature docs: [`docs/README.md`](docs/README.md)
