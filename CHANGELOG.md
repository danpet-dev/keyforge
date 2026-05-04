# Changelog

All notable changes to KeyForge will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-05-04

### Added
- **Age Key Support**: Full support for Age encryption keys alongside PGP
  - `keyforge init --key-type age --generate-key` for Age-based projects
  - `keyforge add-member --key-type age` for adding Age keys
  - Automatic Age key listing from `~/.config/sops/age/keys.txt`
  - Age key generation and secure storage
  
- **Key Rotation Command**: `keyforge rotate-keys` for secure key rotation
  - Zero-downtime re-encryption workflow
  - Supports both PGP and Age key types
  - Automatic backup recommendations
  - Environment-specific rotation

- **Security Audit Command**: `keyforge audit` for access tracking
  - Shows who has access to which secrets
  - Lists encrypted files accessible by each key
  - Identifies unavailable/missing keys
  - JSON output format for automation

- **Docker Support**: Official Docker image for CI/CD pipelines
  - Multi-architecture support (amd64, arm64)
  - Includes SOPS, GPG, and Age
  - Non-root user for security
  - Available at `ghcr.io/danpet-dev/keyforge`

### Changed
- Validate command now checks both PGP and Age keys
- Improved error messages for missing keys
- Better handling of mixed PGP/Age configurations

### Fixed
- Flexible .sops.yaml parsing for both string and array key formats
- SOPS updatekeys compatibility (removed unsupported --output-type flag)

## [0.1.0] - 2026-05-04

### Added
- **Initial MVP Release**
- `keyforge init`: Initialize .sops.yaml with best-practice templates
  - Multi-environment setup (dev/test/prod)
  - Optional PGP key generation
  - Service-specific and database secret rules
  
- `keyforge validate`: Validate .sops.yaml configuration
  - Syntax validation
  - Key availability checks
  - Expiration warnings for PGP keys
  
- `keyforge add-member`: Add team members to .sops.yaml
  - Automatic PGP key generation
  - Environment-specific access
  - Public key export for sharing
  - Automatic file re-encryption
  
- `keyforge update-all`: Bulk update all encrypted files
  - Auto-detection of .sops, .yaml.sops, .json.sops files
  - Format-aware updatekeys execution
  
- `keyforge edit`: Simplified SOPS file editing
  - Auto-detection of file format (yaml/json/env)
  - Direct wrapper for `sops` command

- **Core Infrastructure**
  - Multi-PGP key support with .sops.yaml parsing
  - GitHub Actions CI/CD (test, lint, build, release)
  - Goreleaser configuration for multi-platform builds
  - Comprehensive unit tests for config and sops packages

### Documentation
- README with quickstart guide
- CONTRIBUTING.md for contributors
- MIT License
- Integration with NornForge project documentation

---

## Release Notes

### v0.2.0 Highlights

This release adds **Age encryption support** for faster encryption and simpler key management compared to PGP. Age keys are now first-class citizens in KeyForge.

**Key rotation** is now fully automated with the new `rotate-keys` command, following security best practices for zero-downtime re-encryption.

The **audit command** provides visibility into who has access to your secrets - critical for compliance and security reviews.

**Docker support** makes it easy to integrate KeyForge into CI/CD pipelines (Tekton, GitLab CI, GitHub Actions, etc).

### v0.1.0 Highlights

First stable release of KeyForge! This MVP includes all essential commands for managing SOPS multi-key configurations. The focus was on automating the most painful parts of SOPS: key rotation, team onboarding, and `.sops.yaml` management.

---

## Upgrade Guide

### From v0.1.0 to v0.2.0

No breaking changes. All v0.1.0 commands continue to work.

**New features to try:**
```bash
# Switch to Age for faster encryption
keyforge init --project my-new-project --key-type age --generate-key

# Audit your current access
keyforge audit

# Rotate keys (PGP or Age)
keyforge rotate-keys --key-type age
```

**Docker users:**
```bash
docker pull ghcr.io/danpet-dev/keyforge:0.2.0
```

[0.2.0]: https://github.com/danpet-dev/keyforge/releases/tag/v0.2.0
[0.1.0]: https://github.com/danpet-dev/keyforge/releases/tag/v0.1.0
