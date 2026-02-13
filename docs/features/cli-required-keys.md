# CLI Required Keys Validation

## Summary
Add strict validation mode to ensure required environment variables exist before launching the child command.

## Problem
Injection may finish without error even when critical variables are still missing.

## Proposal
Add `--required` support (repeatable or CSV list). Validate keys after injection and before executing the command.

## Design
- Missing required keys produce a non-zero exit.
- Output includes key names only, never secret values.
- Validation should run regardless of provider combination.

## Acceptance Criteria
- Child command is not executed when required keys are missing.
- Works with `.env`, Vault, both, or neither.
- Includes tests for success and failure paths.

## Implementation Notes
- CLI entrypoint: `/Users/stuft2/Projects/envchain/cmd/envault/main.go`
