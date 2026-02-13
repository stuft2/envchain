# Typed Config Helpers

## Summary
Add helpers to parse common env types with consistent error handling.

## Problem
Consumers repeatedly parse strings into ints, booleans, and durations, with inconsistent errors.

## Proposed API
- `GetIntOrDefault(key string, def int) (int, error)`
- `GetBoolOrDefault(key string, def bool) (bool, error)`
- `GetDurationOrDefault(key string, def time.Duration) (time.Duration, error)`

## Design
- Preserve current `GetEnvOrDefault` behavior.
- Return errors that include the env key and invalid value.
- Clearly document unset vs empty semantics.

## Acceptance Criteria
- Correct values when unset, set, and empty.
- Parse errors are actionable and consistent.
- Table-driven tests cover happy and failure paths.

## Implementation Notes
- Current helper location: `/Users/stuft2/Projects/envchain/env_default.go`
- Tests: `/Users/stuft2/Projects/envchain/env_default_test.go`
