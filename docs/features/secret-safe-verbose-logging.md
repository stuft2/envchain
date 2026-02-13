# Secret-Safe Verbose Logging

## Goal

Keep verbose mode useful for debugging provider behavior without exposing secret values in logs.

## Problem

Current debug logging includes raw environment variable values when keys are set. This can leak credentials, API tokens, or private data into local terminals, CI logs, and centralized log systems.

## Requirements

- Never log secret values.
- Continue logging enough context to debug precedence and provider behavior.
- Keep the `-verbose` flag behavior stable for users.

## Proposed Behavior

- Log key names and source provider only.
- For set operations, log a redacted marker like `[REDACTED]` instead of value.
- Optionally include value length to help diagnose empty/non-empty values.

## Implementation Tasks

1. Update `internal.SetEnvMap` logging to remove `%v` value output.
2. Ensure all provider debug logs avoid printing token/secret values.
3. Add tests that assert logs never contain known secret fixtures.
4. Document redaction behavior in README and CLI help notes.

## Acceptance Criteria

- Running with `-verbose` does not emit plaintext secret values.
- Existing injection behavior remains unchanged.
- Regression tests fail if a secret fixture appears in logs.
