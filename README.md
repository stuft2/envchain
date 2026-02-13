# envault

## Overview

`envault` is a tiny helper for Go services that **backfills** environment variables from one or more **providers** — without overwriting anything that’s already set. It also provides a small helper, `GetEnvOrDefault`, for reading environment variables with defaults.

#### Providers Available:

- **Dotenv**: read a local `.env` file
- **HashiCorp Vault** (KV v2): fetch secrets over HTTP using a Vault token

Order matters and establishes precedence: **existing process env** ➜ **first provider** ➜ **second provider** ➜ ...

Anything already set in the process wins. Missing keys are filled by the first provider you pass; still-missing keys are filled by the next provider, and so on.

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
