[![Tests](https://github.com/danpet-dev/keyforge/actions/workflows/test.yml/badge.svg)](https://github.com/danpet-dev/keyforge/actions/workflows/test.yml)

# KeyForge

**Forge your encryption keys with confidence**

KeyForge is a CLI tool that simplifies SOPS multi-key lifecycle management. It automates common tasks like key rotation, team onboarding, and `.sops.yaml` configuration, making it easy to manage encrypted secrets across multiple environments.

## Features

- 🔑 **Multi-Key Support**: PGP and Age key management
- 🔄 **Automated Key Rotation**: Secure key rotation with zero-downtime re-encryption
- 👥 **Team Onboarding**: Add team members and rotate keys with single commands
- 📊 **Security Audit**: Track who has access to which secrets
- ✅ **Validation**: Check `.sops.yaml` syntax and key availability
- 🛠️ **Best Practices**: Generate `.sops.yaml` with proven multi-environment patterns
- 📝 **Smart Edit**: Wrapper for `sops` with auto-detection of file formats
- 🐳 **CI/CD Ready**: Docker image for pipeline integration

## Installation

### From Source

```bash
git clone https://github.com/danpet-dev/keyforge
cd keyforge
go build -o keyforge ./cmd/keyforge
sudo mv keyforge /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/danpet-dev/keyforge/cmd/keyforge@latest
```

### Docker

```bash
# Pull image
docker pull ghcr.io/danpet-dev/keyforge:latest

# Run command
docker run --rm -v $(pwd):/workspace ghcr.io/danpet-dev/keyforge:latest validate

# With Age keys
docker run --rm \
  -v $(pwd):/workspace \
  -v ~/.config/sops/age:/home/keyforge/.config/sops/age:ro \
  ghcr.io/danpet-dev/keyforge:latest audit
```

## Quick Start

```bash
# Initialize .sops.yaml with PGP keys (default)
keyforge init --project my-project --generate-key

# Or use Age keys for faster encryption
keyforge init --project my-project --key-type age --generate-key

# Add a team member
keyforge add-member --name "Alice" --email alice@example.com --key-type age --generate-key

# Audit access permissions
keyforge audit

# Rotate encryption keys (recommended every 6-12 months)
keyforge rotate-keys --key-type age

# Update all encrypted files with new keys
keyforge update-all

# Validate configuration
keyforge validate

# Edit encrypted file
keyforge edit secrets/production.yaml.sops
```

## Why KeyForge?

SOPS multi-key management can be complex and error-prone:

- `sops updatekeys` requires manual `--input-type yaml` flags
- Key rotation across multiple files is tedious
- Team onboarding involves multiple manual steps
- Placeholder keys and misconfigurations cause failures

**KeyForge solves these problems** by automating SOPS best practices.

## Documentation

- [Installation Guide](docs/installation.md) (Coming soon)
- [Command Reference](docs/commands.md) (Coming soon)
- [Quick Start Tutorial](docs/quickstart.md) (Coming soon)

## Use Cases

KeyForge is used in production by [NornForge](https://github.com/danpet-dev/NornForge) - a local, autonomous, trusted agent platform.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Status

**Current Version:** v0.2.0

- ✅ **MVP Complete** (v0.1.0): init, add-member, update-all, validate, edit
- ✅ **Enhanced Features** (v0.2.0): Age key support, rotate-keys, audit, Docker image
- 🚧 **In Development**: Comprehensive documentation, integration tests

See [CHANGELOG.md](CHANGELOG.md) for full release history.
