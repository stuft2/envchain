# Quickstart and Troubleshooting Docs

## Goal

Reduce first-run friction by giving users a short success path and clear fixes for common failures.

## Problem

Current README explains concepts well, but new users still need a fast copy/paste path and targeted diagnostics.

## Requirements

- Add a one-minute quickstart.
- Include common error messages and fixes.
- Keep examples minimal and environment-agnostic.

## Quickstart Scope

1. Install CLI.
2. Create sample `.env`.
3. Run `envault -- env` (or equivalent) and verify expected variables.
4. Optional Vault example with required env vars.

## Troubleshooting Scope

- Missing `VAULT_ADDR`
- Missing token (`VAULT_TOKEN` / `~/.vault-token`)
- Invalid Vault path format
- Command execution failures after `--`
- Expected precedence behavior (existing env wins)

## Implementation Tasks

1. Add Quickstart section near top of README.
2. Add Troubleshooting section with symptom -> cause -> fix format.
3. Include one explicit precedence example.
4. Add notes about verbose mode redaction once implemented.

## Acceptance Criteria

- A new user can run a successful local example in under five minutes.
- Most common support questions are answerable from docs.
- Troubleshooting entries map directly to real error strings.
