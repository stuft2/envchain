# Versioning and Compatibility Policy

## Goal

Set user expectations for change safety, upgrade paths, and supported environments.

## Problem

Without a documented version policy, consumers cannot plan upgrades or assess breaking-change risk.

## Requirements

- Adopt semantic versioning for tags and releases.
- Define what constitutes a breaking change.
- Define supported Go versions and support window.

## Proposed Policy

- Use SemVer tags: `vMAJOR.MINOR.PATCH`.
- Breaking API or CLI contract change increments `MAJOR`.
- Backward-compatible features increment `MINOR`.
- Fixes only increment `PATCH`.
- Support latest two Go minor versions (for example, `1.22` and `1.23`).

## Implementation Tasks

1. Add a Versioning section in README.
2. Create `CHANGELOG.md` for release notes.
3. Enforce tagging and release notes in release process.
4. Ensure CI runs on all supported Go versions.

## Acceptance Criteria

- Release tags follow SemVer.
- Compatibility guarantees are documented.
- Consumers can identify upgrade risk from version numbers and changelog entries.
