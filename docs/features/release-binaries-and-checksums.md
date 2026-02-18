# Release Binaries and Checksums

## Goal

Allow users to install and use `envchain` without requiring a local Go toolchain.

## Problem

Current installation flow assumes Go is installed and configured. This is a barrier for operations users and mixed-language teams.

## Requirements

- Build binaries for common OS/architectures.
- Publish signed or at least checksummed artifacts.
- Make installation commands simple and scriptable.

## Suggested Targets

- macOS: `amd64`, `arm64`
- Linux: `amd64`, `arm64`
- Windows: `amd64`

## Implementation Tasks

1. Add a release workflow (e.g., GoReleaser or custom GitHub Actions).
2. Produce archives per platform with `envchain` binary.
3. Generate `checksums.txt` for each release.
4. Document install methods:
   - direct binary download
   - Homebrew tap (optional)
   - `go install` fallback

## Acceptance Criteria

- Tagged releases include multi-platform binaries and checksums.
- Install instructions work from a clean machine.
- Release process is reproducible and documented.
