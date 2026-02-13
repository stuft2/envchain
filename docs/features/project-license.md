# Project License

## Goal

Provide clear legal terms so organizations can evaluate and adopt `envault` without ambiguity.

## Problem

Without a license file, many companies and open source programs cannot use, redistribute, or contribute to the project.

## Requirements

- Add a standard SPDX-recognized open source license file.
- State license clearly in repository metadata and docs.
- Keep contribution and redistribution terms unambiguous.

## Recommended Choice

Use a permissive license such as MIT unless project constraints require stronger copyleft terms.

## Implementation Tasks

1. Add `LICENSE` at repository root.
2. Add a short License section in `README.md`.
3. If needed, add copyright holder and year.
4. Ensure package/release metadata references the same license.

## Acceptance Criteria

- `LICENSE` exists in repo root.
- README references the license.
- Scanners (e.g., GitHub license detection) identify the license correctly.
