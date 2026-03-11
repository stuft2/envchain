---
status: "Implemented"
implemented_in:
  - "[`internal/inject/inject.go`](../../internal/inject/inject.go)"
  - "[`internal/provider.go`](../../internal/provider.go)"
  - "[`providers/vault/provider.go`](../../providers/vault/provider.go)"
  - "[`internal/inject/inject_test.go`](../../internal/inject/inject_test.go)"
  - "[`providers/vault/provider_test.go`](../../providers/vault/provider_test.go)"
  - "[`README.md`](../../README.md)"
gaps: []
---

# Inject With Context

## Summary
Add a context-aware injection API so all providers can participate in request cancellation and deadlines.

## Problem
The injection orchestration had no shared `context.Context`, so I/O-heavy providers could not be canceled as a unit.

## Proposed API
```go
func RunWithContext(ctx context.Context, providers ...internal.Provider) error
```

## Design
- Keep the non-context injection path available internally.
- Add an optional context-aware provider contract (for example, `InjectContext(context.Context) error`).
- If a provider does not support context, fall back to `Inject()`.

## Acceptance Criteria
- Canceling context stops context-aware providers.
- Legacy providers still work unchanged.
- Tests cover mixed provider sets and cancellation paths.

## Implementation Notes
- Internal injection package: `/Users/stuft2/Projects/envchain/internal/inject/inject.go`
- Provider contracts: `/Users/stuft2/Projects/envchain/internal/provider.go`
- Vault provider can be first adopter: `/Users/stuft2/Projects/envchain/providers/vault/provider.go`
