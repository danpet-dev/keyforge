# KeyForge Roadmap

This document outlines the planned features and improvements for KeyForge.

## Vision

KeyForge aims to be the **definitive CLI tool** for SOPS multi-key lifecycle management, making encrypted secrets management as simple as using Git.

---

## ✅ Released

### v0.1.0 - MVP (2026-05-04)
- ✅ `init` - Initialize .sops.yaml with best practices
- ✅ `validate` - Validate .sops.yaml configuration
- ✅ `add-member` - Add team members with key generation
- ✅ `update-all` - Bulk update encrypted files
- ✅ `edit` - Simplified SOPS file editing

### v0.2.0 - Age & Advanced Features (2026-05-04)
- ✅ Age encryption support (alongside PGP)
- ✅ `rotate-keys` - Zero-downtime key rotation
- ✅ `audit` - Access tracking and reporting
- ✅ Docker support with multi-arch images

### v0.2.1 - CI/CD Stability (2026-05-04)
- ✅ Go 1.26 compatibility
- ✅ Code quality improvements (errcheck, gofmt)
- ✅ Improved release pipeline

### v0.2.2 - Homebrew Support (2026-05-04)
- ✅ Homebrew tap for macOS and Linux
- ✅ Auto-installation of dependencies

### v0.3.0 - Essential Commands (2026-05-05)
- ✅ `keys list` - Central overview of all available keys
- ✅ `remove-member` - Team offboarding
- ✅ `status` - Quick overview of secrets state
- ✅ `decrypt` - Simplified decryption
- ✅ `encrypt` - Simplified encryption
- ✅ `diff` - Code reviews for encrypted files

### v0.3.1 - CI/CD Compatibility (2026-05-06)
- ✅ Fixed golangci-lint errcheck violations
- ✅ Improved error handling in defer statements
- ✅ CI badge passing (green)

---

## 🚀 In Progress

### v0.4.0 - Team Collaboration (Target: Q3 2026)

**Theme:** Making team-based secret management seamless.

- [ ] **`keyforge sync`** - Team key synchronization (#7)
  - Detect missing keys for full access
  - Generate import instructions for team members
  - Support key exchange workflows
  
- [ ] **`keyforge template`** - Secret templates (#8)
  - Create reusable secret templates
  - Fill templates with values
  - Ensure consistent secret structure across environments
  
- [ ] **`keyforge import/export`** - Migration helpers (#9)
  - Import from: .env, Kubernetes Secrets, Vault, AWS Secrets Manager
  - Export to: .env, Kubernetes manifests, JSON, YAML
  - Support for bulk migrations

---

## 📅 Planned

### v0.4.0 - Team Collaboration (Target: Q3 2026)

**Theme:** Making team-based secret management seamless.

- [ ] **`keyforge sync`** - Team key synchronization
  - Detect missing keys for full access
  - Generate import instructions for team members
  - Support key exchange workflows
  
- [ ] **`keyforge template`** - Secret templates
  - Create reusable secret templates
  - Fill templates with values
  - Ensure consistent secret structure across environments
  
- [ ] **`keyforge import/export`** - Migration helpers
  - Import from: .env, Kubernetes Secrets, Vault, AWS Secrets Manager
  - Export to: .env, Kubernetes manifests, JSON, YAML
  - Support for bulk migrations

---

## 📅 Planned

### v0.5.0 - CI/CD Integration (Target: Q4 2026)

**Theme:** First-class CI/CD pipeline support.

- [ ] **`keyforge check`** - Validation and linting
  - Validate .sops.yaml syntax and key availability
  - Check for plaintext secrets in repository
  - Warn about expiring keys
  - Strict mode for CI/CD (fail on warnings)
  - Pre-commit hook support
  
- [ ] **`keyforge ci setup`** - CI/CD environment setup
  - Auto-configure from environment variables
  - Support for GitHub Actions, GitLab CI, Jenkins
  - Validate CI environment has necessary keys
  
- [ ] **`keyforge ci decrypt-all`** - Bulk decrypt for builds
  - Decrypt all secrets for CI/CD pipeline
  - Output to temporary directory
  - Clean up after build
  
- [ ] **GitHub Actions Integration**
  - `danpet-dev/keyforge-action` for easy setup
  - Pre-built workflows for common patterns
  - Automatic secret validation in PRs

### v0.6.0 - Security & Compliance (Target: Q1 2027)

**Theme:** Enterprise-grade security features.

- [ ] **`keyforge report`** - Compliance reporting
  - Access matrix (who can decrypt what)
  - Key expiration tracking
  - Audit trail of changes
  - Export to Markdown, JSON, HTML
  
- [ ] **`keyforge backup/restore`** - Disaster recovery
  - Encrypted backup of all keys
  - Restore from backup
  - Verification and testing
  
- [ ] **`keyforge watch`** - Auto-encryption
  - Watch directories for changes
  - Auto-encrypt plaintext secrets
  - Git pre-commit hook integration
  - Prevent committing plaintext secrets

### v0.7.0 - Advanced Features (Target: Q2 2027)

**Theme:** Power user features and optimizations.

- [ ] **Interactive TUI Mode** - Terminal UI for visual workflows
- [ ] **Secrets Versioning** - Track secret history and rollback
- [ ] **Multi-Region Support** - Different keys per region/cluster
- [ ] **Secret Scanning** - Detect accidentally committed secrets
- [ ] **Performance Optimizations** - Parallel operations, caching

---

## 🔮 Future Ideas (No ETA)

Ideas under consideration for future releases:

- **Web UI** - Browser-based secret management
- **VS Code Extension** - Integrated secret editing in IDE
- **Secrets Rotation Automation** - Auto-rotate based on policies
- **Integration with Cloud KMS** - AWS KMS, GCP KMS, Azure Key Vault
- **Secret Sharing** - Secure one-time secret sharing
- **RBAC** - Role-based access control for teams
- **Compliance Frameworks** - SOC 2, GDPR, HIPAA helpers
- **Secrets Drift Detection** - Compare environments
- **AI-Powered Suggestions** - Detect weak patterns, suggest improvements

---

## 🤝 Contributing

Want to help build KeyForge? Check out [CONTRIBUTING.md](CONTRIBUTING.md) for:
- How to pick up a feature from the roadmap
- Development setup
- Testing guidelines
- PR process

**Looking for contributors for:**
- CI/CD integrations (GitHub Actions, GitLab CI)
- Cloud provider integrations (AWS, GCP, Azure)
- Documentation and examples
- Testing and bug reports

---

## 📢 Feedback

Have ideas or suggestions? We'd love to hear from you!

- **GitHub Issues**: https://github.com/danpet-dev/keyforge/issues
- **Discussions**: https://github.com/danpet-dev/keyforge/discussions
- **Email**: Create an issue instead for transparency

---

## 📝 Notes

**Versioning:**
- **Major (x.0.0)**: Breaking changes, major features
- **Minor (0.x.0)**: New features, backward compatible
- **Patch (0.0.x)**: Bug fixes, documentation

**Release Cadence:**
- Feature releases: ~6-8 weeks
- Patch releases: As needed
- Security fixes: Immediate

**Priorities:**
1. Security and correctness
2. User experience and documentation
3. Performance and optimization
4. Advanced features

---

Last updated: 2026-05-04
