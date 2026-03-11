---
status: "Implemented"
implemented_in:
  - "[`inject.go`](../../inject.go)"
  - "[`internal/provider.go`](../../internal/provider.go)"
  - "[`providers/vault/provider.go`](../../providers/vault/provider.go)"
  - "[`inject_test.go`](../../inject_test.go)"
  - "[`providers/vault/provider_test.go`](../../providers/vault/provider_test.go)"
  - "[`README.md`](../../README.md)"
gaps: []
---

# Inject With Context

## Summary
Add a context-aware injection API so all providers can participate in request cancellation and deadlines.

## Problem
`Inject(...)` has no shared `context.Context`, so I/O-heavy providers cannot be canceled as a unit.

## Proposed API
```go
func InjectWithContext(ctx context.Context, providers ...internal.Provider) error
```

## Design
- Keep `Inject(...)` for backward compatibility.
- Add an optional context-aware provider contract (for example, `InjectContext(context.Context) error`).
- If a provider does not support context, fall back to `Inject()`.

## Acceptance Criteria
- Canceling context stops context-aware providers.
- Legacy providers still work unchanged.
- Tests cover mixed provider sets and cancellation paths.

## Implementation Notes
- Core package: `/Users/stuft2/Projects/envchain/inject.go`
- Provider contracts: `/Users/stuft2/Projects/envchain/internal/provider.go`
- Vault provider can be first adopter: `/Users/stuft2/Projects/envchain/providers/vault/provider.go`
