# Vault Result Caching

## Summary
Add optional caching to reduce repeated Vault network requests.

## Problem
Repeated local and CI runs can repeatedly fetch identical secret payloads.

## Proposal
Implement opt-in caching with short TTL:
- in-memory cache (process lifetime)
- optional file cache (cross-process)

## Design
- Cache key includes address, namespace, and secret path.
- TTL defaults to a conservative value.
- Cache can be disabled explicitly.

## Acceptance Criteria
- Cache hits avoid network calls within TTL.
- Expired entries trigger refresh.
- Tests verify hit/miss/expiry behavior.

## Implementation Notes
- Vault provider implementation: `/Users/stuft2/Projects/envchain/providers/vault/provider.go`
- Vault tests: `/Users/stuft2/Projects/envchain/providers/vault/provider_test.go`
