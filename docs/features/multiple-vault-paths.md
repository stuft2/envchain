# Multiple Vault Paths

## Summary
Allow fetching secrets from more than one Vault path in a single run.

## Problem
Teams commonly split secrets across shared and app-specific paths.

## Proposal
Support repeated `--vault-path` flags (or a deterministic list format) and apply in order.

## Design
- First path fills unset keys.
- Later paths fill only remaining gaps.
- Existing process environment still has highest precedence.

## Acceptance Criteria
- Path precedence is documented and tested.
- Behavior is deterministic across repeated runs.
- Works with current "do not overwrite set env" behavior.

## Implementation Notes
- CLI provider assembly: `/Users/stuft2/Projects/envchain/cmd/envault/main.go`
- Vault provider behavior: `/Users/stuft2/Projects/envchain/providers/vault/provider.go`
