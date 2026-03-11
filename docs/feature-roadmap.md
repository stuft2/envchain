# Feature Roadmap

This document tracks implementation status for each feature specification.

Last Updated: March 11, 2026

## Status Snapshot

| Feature | Status | Evidence | Last Updated |
|---|---|---|---|
| [Inject With Context](./features/inject-with-context.md) | Implemented | [`inject.go`](../inject.go), [`internal/provider.go`](../internal/provider.go), [`inject_test.go`](../inject_test.go) | March 11, 2026 |
| [Typed Config Helpers](./features/typed-config-helpers.md) | Not Started | None yet | March 11, 2026 |
| [CLI Required Keys Validation](./features/cli-required-keys.md) | Not Started | None yet | March 11, 2026 |
| [Multiple Vault Paths](./features/multiple-vault-paths.md) | Not Started | None yet | March 11, 2026 |
| [Secret-Safe Diagnostics](./features/secret-safe-diagnostics.md) | Not Started | None yet | March 11, 2026 |
| [Vault Result Caching](./features/vault-result-caching.md) | Not Started | None yet | March 11, 2026 |
| [Additional Providers](./features/additional-providers.md) | Not Started | None yet | March 11, 2026 |
| [Key Filtering And Transform Rules](./features/key-filtering-and-transform.md) | Not Started | None yet | March 11, 2026 |
| [Quickstart And Troubleshooting Docs](./features/quickstart-and-troubleshooting-docs.md) | Implemented | [`README.md`](../README.md) | March 11, 2026 |
| [CLI Test Coverage](./features/cli-test-coverage.md) | Implemented | [`cmd/envchain/main_test.go`](../cmd/envchain/main_test.go) | March 11, 2026 |
| [Continuous Integration](./features/continuous-integration.md) | Partial | [`.github/workflows/ci.yml`](../.github/workflows/ci.yml), [`README.md`](../README.md) | March 11, 2026 |
| [Release Binaries and Checksums](./features/release-binaries-and-checksums.md) | Not Started | None yet | March 11, 2026 |
| [Project License](./features/project-license.md) | Not Started | None yet | March 11, 2026 |
| [Secret-Safe Verbose Logging](./features/secret-safe-verbose-logging.md) | Implemented | [`internal/provider.go`](../internal/provider.go), [`cmd/envchain/main_test.go`](../cmd/envchain/main_test.go), [`README.md`](../README.md) | March 11, 2026 |
| [Versioning and Compatibility Policy](./features/versioning-and-compatibility-policy.md) | Not Started | None yet | March 11, 2026 |

## Per-Feature Docs

1. [Inject With Context](./features/inject-with-context.md)
1. [Typed Config Helpers](./features/typed-config-helpers.md)
1. [CLI Required Keys Validation](./features/cli-required-keys.md)
1. [Multiple Vault Paths](./features/multiple-vault-paths.md)
1. [Secret-Safe Diagnostics](./features/secret-safe-diagnostics.md)
1. [Vault Result Caching](./features/vault-result-caching.md)
1. [Additional Providers](./features/additional-providers.md)
1. [Key Filtering And Transform Rules](./features/key-filtering-and-transform.md)
1. [Quickstart And Troubleshooting Docs](./features/quickstart-and-troubleshooting-docs.md)
1. [CLI Test Coverage](./features/cli-test-coverage.md)
1. [Continuous Integration](./features/continuous-integration.md)
1. [Release Binaries and Checksums](./features/release-binaries-and-checksums.md)
1. [Project License](./features/project-license.md)
1. [Secret-Safe Verbose Logging](./features/secret-safe-verbose-logging.md)
1. [Versioning and Compatibility Policy](./features/versioning-and-compatibility-policy.md)

## Suggested Delivery Order

1. Inject with context, required-key validation, and CLI tests.
1. Secret-safe diagnostics and multiple Vault paths.
1. Key filtering and result caching.
1. Additional providers.
1. CI hardening and release process improvements.
