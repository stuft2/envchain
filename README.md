# envchain

## Overview

`envchain` is a tiny helper for Go services that **backfills** environment variables from one or more **providers** — without overwriting anything that’s already set. It also provides a small helper, `GetEnv`, for reading environment variables with defaults.

#### Providers Available:

- **Dotenv**: read a local `.env` file
- **HashiCorp Vault** (KV v2): fetch secrets over HTTP using a Vault token

Order matters and establishes precedence: **existing process env** ➜ **first provider** ➜ **second provider** ➜ ...

Anything already set in the process wins. Missing keys are filled by the first provider you pass; still-missing keys are filled by the next provider, and so on.

---

## CLI Usage

Local dev:
1. Put a `.env` file beside your app.
2. Optionally set `VAULT_ADDR` and authenticate to backfill missing secrets from Vault:
   ```shell
   VAULT_ADDR=https://vault.byu.edu vault login -method=oidc -path=byu-sso
   ```

1. Install the CLI:
   ```bash
   go install github.com/stuft2/envchain/cmd/envchain@latest
   ```
1. Create a sample `.env`:
   ```bash
   cat > .env <<'EOF'
   APP_NAME=envchain-demo
   PORT=8080
   EOF
   ```
1. Run a command with backfilled env vars:
   ```bash
   envchain -- env | grep -E '^(APP_NAME|PORT)='
   ```
1. Verify output:
   ```text
   APP_NAME=envchain-demo
   PORT=8080
   ```

Optional Vault check:
```bash
envchain -vault-path "kvv2/<service>/dev/env-vars" -- env | grep '^YOUR_KEY='
```

CLI usage:

```bash
envchain [flags] -- <command> [args...]
```

Flags:

- `-dotenv` (default `.env`): path to a dotenv file to backfill from. Pass an empty string to skip it.
- `-vault-path`: KV v2 path to load from HashiCorp Vault. Leave empty to skip Vault.
- `-verbose`: emit debug logs detailing how each provider resolves environment variables.

Environment variables such as `VAULT_ADDR`, `VAULT_TOKEN`, and `VAULT_NAMESPACE` still control Vault behaviour.

CI/Prod: rely on process env only. If you don’t set `VAULT_ADDR`, the Vault provider is inert, and a missing `.env` is ignored.

## Modules

`envchain` now exposes configuration helpers separately from injection orchestration:

- The `envchain` CLI performs environment injection.
- The `config` package provides environment lookup and parsing helpers.

Injection orchestration is internal to this repository and is not exposed as a public Go API.

### Vault provider requirements

The Vault provider is active when these are satisfied:
 - VAULT_ADDR — base URL for Vault (e.g., https://vault.byu.edu or https://vault.byu.edu/v1)
 - A token:
   - VAULT_TOKEN, or
   - ~/.vault-token (created by vault login)
 - Secret path you pass to vault.NewProvider(...) (KV v2 path, e.g. `kvv2/<service-name>/dev/env-vars`)
 - Optional: VAULT_NAMESPACE → sent as X-Vault-Namespace

Timeouts: Vault HTTP requests use a 10s timeout.

Context: The Vault provider uses a background context by default. To override, set `Provider.Context` before the provider is used by the CLI or internal injection flow.

## Loading Config From Environment Tags

Define a config struct and let `config.Load` populate it from `env` tags:

```go
package main

import (
	"fmt"
	"net/url"
	"time"

	"github.com/stuft2/envchain/config"
)

type appConfig struct {
	Port    int           `env:"PORT,default=8080"`
	Addr    string        `env:"ADDR,default=:http"`
	Debug   bool          `env:"DEBUG,default=false"`
	Timeout time.Duration `env:"TIMEOUT,default=5s"`
	BaseURL *url.URL      `env:"BASE_URL,required"`
}

func main() {
	var cfg appConfig
	if err := config.Load(&cfg); err != nil {
		panic(err)
	}

	fmt.Println("Starting server on", cfg.Addr, "port", cfg.Port)
	fmt.Println("Debug mode =", cfg.Debug)
}
```

`config.Load` reads exported struct fields with an `env` tag. It walks nested structs, applies defaults, and returns a single aggregated error if any required values are missing or any values fail to parse.

Tag format:

```go
`env:"ENV_KEY,option,option=value"`
```

General rules:

- The first tag segment is always the environment variable name.
- Omit the tag to ignore a field entirely.
- Use `env:"-"` to explicitly ignore an exported field.
- If the env var is set, its value wins over `default=...`.
- An empty string still counts as "set". Defaults only apply when the variable is unset.
- Nested structs are traversed automatically. Tagged fields inside them are loaded the same way as top-level fields.
- Errors are aggregated and returned together so you can fix all invalid or missing variables in one pass.

Supported field types:

- `string`
- `bool`
- signed integers and unsigned integers
- `float32` and `float64`
- `time.Duration`
- `time.Time`
- `url.URL` and `*url.URL`
- `[]string`
- `map[string]string`

### `required`

Marks a field as mandatory. If the environment variable is unset and no default is provided, `Load` reports an error.

```go
type config struct {
	BaseURL *url.URL `env:"BASE_URL,required"`
}
```

Notes:

- `required` only checks for "unset", not "empty".
- If both `required` and `default=...` are present, the default satisfies the requirement when the env var is unset.

### `default=...`

Provides a fallback value used only when the environment variable is unset.

```go
type config struct {
	Port    int           `env:"PORT,default=8080"`
	Timeout time.Duration `env:"TIMEOUT,default=5s"`
	Mode    string        `env:"MODE,default=dev"`
}
```

Notes:

- The default string is parsed with the same rules as a real environment value.
- Invalid defaults fail during `Load` just like invalid env values.
- Defaults for slices and maps use the same separators as parsed env values.

### `sep=...`

Overrides the separator for `[]string` fields. The default separator is `,`.

```go
type config struct {
	Hosts []string `env:"HOSTS,sep=|"`
}
```

Examples:

- `HOSTS=api|worker|admin` with `sep=|` becomes `[]string{"api", "worker", "admin"}`
- `HOSTS=api, worker` without `sep=...` becomes `[]string{"api", "worker"}`

Notes:

- Whitespace around entries is trimmed.
- Empty entries are skipped.

### `entrysep=...` and `kvsep=...`

Override separators for `map[string]string` fields. Defaults are `entrysep=,` and `kvsep==`.

```go
type config struct {
	Labels map[string]string `env:"LABELS,entrysep=;,kvsep=:"`
}
```

Examples:

- `LABELS=team=platform,service=envchain` becomes `map[string]string{"team": "platform", "service": "envchain"}`
- `LABELS=team:platform;service:envchain` with `entrysep=;` and `kvsep=:` parses the same data

Notes:

- Keys are trimmed and must not be empty.
- Values are trimmed.
- Invalid entries such as `broken` or `=value` produce an error.

### `layout=...`

Defines the parse layout for `time.Time` fields. This option is required for every `time.Time` field.

```go
type config struct {
	StartedAt time.Time `env:"STARTED_AT,layout=2006-01-02"`
}
```

Notes:

- Layouts use Go's `time.Parse` reference time format.
- If a `time.Time` field is tagged without `layout=...`, `Load` returns an error.

### `oneof=...`

Constrains `string` fields to an allowed set of values.

```go
type config struct {
	Mode string `env:"MODE,default=dev,oneof=dev|staging|prod"`
}
```

Notes:

- `oneof` currently applies to `string` fields.
- Comparison is exact and case-sensitive.
- Defaults are also validated against the allowed set.

### `format=bytes`

Enables byte-size parsing for signed integer fields.

```go
type config struct {
	MaxBytes int64 `env:"MAX_BYTES,default=256MiB,format=bytes"`
}
```

Examples:

- `42` => `42`
- `2KB` => `2048`
- `4MiB` => `4194304`

Supported units:

- `B`
- `K`, `KB`, `KiB`
- `M`, `MB`, `MiB`
- `G`, `GB`, `GiB`
- `T`, `TB`, `TiB`

Notes:

- Units are case-insensitive.
- Negative values are rejected.
- Overflow for the target integer type returns an error.

### Full Example

```go
type appConfig struct {
	Port      int               `env:"PORT,default=8080"`
	Debug     bool              `env:"DEBUG,default=false"`
	Timeout   time.Duration     `env:"TIMEOUT,default=5s"`
	BaseURL   *url.URL          `env:"BASE_URL,required"`
	Aliases   []string          `env:"ALIASES,default=api,worker"`
	Labels    map[string]string `env:"LABELS,default=team=platform,service=envchain"`
	StartedAt time.Time         `env:"STARTED_AT,layout=2006-01-02"`
	Mode      string            `env:"MODE,default=dev,oneof=dev|prod"`
	MaxBytes  int64             `env:"MAX_BYTES,default=256MiB,format=bytes"`
}
```

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

---

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

Symptom: `Usage: envchain [flags] -- command [args...]`  
Cause: command was not passed after `--`.  
Fix: provide a command after separator, for example `envchain -- env`.

Symptom: `envchain: failed to execute "..."`  
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

## Project Docs

- Usability feature docs: [`docs/README.md`](docs/README.md)
