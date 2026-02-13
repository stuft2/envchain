# Additional Providers

## Summary
Expand beyond dotenv and Vault to support other secret/config backends.

## Problem
Many teams rely on cloud secret managers or SaaS secret products.

## Proposal
Add optional providers for:
- AWS SSM / Secrets Manager
- GCP Secret Manager
- 1Password / Doppler (as practical)

## Design
- Keep provider packages isolated under `providers/`.
- Avoid hard dependencies in core package.
- Start with one provider to validate extension patterns.

## Acceptance Criteria
- At least one new provider with docs and tests.
- Core `Inject` API remains unchanged.
- Provider behavior follows existing precedence semantics.

## Implementation Notes
- Existing provider pattern reference:
  - `/Users/stuft2/Projects/envchain/providers/dotenv/provider.go`
  - `/Users/stuft2/Projects/envchain/providers/vault/provider.go`
