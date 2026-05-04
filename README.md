# KeyForge

**Forge your encryption keys with confidence**

KeyForge is a CLI tool that simplifies SOPS multi-key lifecycle management. It automates common tasks like key rotation, team onboarding, and `.sops.yaml` configuration, making it easy to manage encrypted secrets across multiple environments.

## Features

- 🔑 **Simplified Key Management**: Add team members and rotate keys with single commands
- 🔄 **Automated Key Updates**: Update all encrypted files with `keyforge update-all`
- ✅ **Validation**: Check `.sops.yaml` syntax and key availability
- 🛠️ **Best Practices**: Generate `.sops.yaml` with proven multi-environment patterns
- 📝 **Smart Edit**: Wrapper for `sops` with auto-detection of file formats

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

## Quick Start

```bash
# Initialize .sops.yaml for your project
keyforge init --project my-project

# Add a team member
keyforge add-member --name "Alice" --email alice@example.com --environments dev,test,prod

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

🚧 **Alpha** - Under active development. Breaking changes may occur.

**Target:** v0.1.0 release (MVP commands: init, add-member, update-all, validate, edit)
