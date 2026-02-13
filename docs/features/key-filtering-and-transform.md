# Key Filtering And Transform Rules

## Summary
Add controls to limit and normalize which fetched keys become environment variables.

## Problem
Bulk secret payloads may include keys that should not be injected into application env.

## Proposal
Introduce configurable controls:
- allowlist explicit keys
- require key prefix
- strip prefix before setting env

## Design
- Implement as provider options or shared transform pipeline.
- Default behavior remains unchanged when options are not set.

## Acceptance Criteria
- Rules are deterministic and composable.
- Include/exclude/strip behavior is documented and tested.
- No regressions to existing precedence behavior.

## Implementation Notes
- Current env setting path: `/Users/stuft2/Projects/envchain/internal/provider.go`
- Provider-specific mapping points:
  - `/Users/stuft2/Projects/envchain/providers/dotenv/provider.go`
  - `/Users/stuft2/Projects/envchain/providers/vault/provider.go`
