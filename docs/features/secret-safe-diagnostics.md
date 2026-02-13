# Secret-Safe Diagnostics

## Summary
Add a planning mode to show what will change without exposing sensitive values.

## Problem
Users need visibility for debugging and CI, but logs must avoid secret leakage.

## Proposal
Add `--dry-run` and/or `--plan` mode that reports:
- active providers
- keys that would be set
- keys skipped due to existing env

## Design
- Never print secret values.
- Dry-run performs reads and evaluation, but no writes.
- Keep output format stable for automation.

## Acceptance Criteria
- Diagnostics contain keys only.
- No writes happen in dry-run mode.
- Tests verify redaction and behavior.

## Implementation Notes
- Logging utility: `/Users/stuft2/Projects/envchain/internal/logging.go`
- Env mutation path: `/Users/stuft2/Projects/envchain/internal/provider.go`
- CLI flags/output: `/Users/stuft2/Projects/envchain/cmd/envchain/main.go`
