# CLI Test Coverage

## Goal

Validate `cmd/envault` behavior as a stable interface for end users and automation.

## Problem

Core providers are tested, but CLI command wiring has no direct tests. This risks regressions in argument parsing, exit codes, and subprocess behavior.

## Requirements

- Test flag parsing and usage behavior.
- Test exit code passthrough from wrapped command.
- Test provider error handling and failure exit paths.

## High-Value Test Cases

1. `-h` returns exit code `0` and prints usage.
2. Missing command returns exit code `2`.
3. Wrapped command non-zero exit propagates correctly.
4. Non-existent command returns clear execution error.
5. `-dotenv` empty disables dotenv provider.
6. `-vault-path` wiring triggers Vault provider creation.
7. `-verbose` enables debug logging without secret leakage.

## Implementation Tasks

1. Add tests around `run(args []string)` in `cmd/envault/main_test.go`.
2. Isolate process execution for testability (inject command runner if needed).
3. Capture stderr/stdout in tests for deterministic assertions.
4. Add test fixtures for provider failure paths.

## Acceptance Criteria

- CLI tests cover success and failure paths.
- Exit code semantics are locked by tests.
- Tests run in CI with no network dependency.
