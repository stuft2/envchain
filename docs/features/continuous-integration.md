# Continuous Integration

## Goal

Automatically verify code quality on every pull request and push.

## Problem

Without CI, users and contributors cannot quickly trust that changes preserve behavior across environments.

## Requirements

- Run tests and static checks automatically.
- Fail fast on regressions.
- Keep feedback fast and deterministic.

## Proposed Baseline Workflow

- Trigger on `pull_request` and `push`.
- Use supported Go versions.
- Run:
  - `go test ./...`
  - `go vet ./...`

Optional follow-ups:

- `golangci-lint`
- race tests (`go test -race ./...`)
- coverage upload/reporting

## Implementation Tasks

1. Add `.github/workflows/ci.yml`.
2. Cache Go modules/build artifacts.
3. Add status badge to README.
4. Document local command parity with CI.
5. Configure required status checks so failing CI blocks merge.

## Acceptance Criteria

- CI runs automatically on PRs and main branch pushes.
- Merge is blocked when required checks fail.
- Commands in docs match CI behavior.
- Failures are visible and actionable from PR checks.

## Implementation Notes

- Workflows live under `.github/workflows/`.
