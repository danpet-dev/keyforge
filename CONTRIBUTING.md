# Contributing to KeyForge

Thank you for your interest in contributing to KeyForge!

## Development Setup

1. **Prerequisites**
   - Go 1.21 or higher
   - Git
   - SOPS installed (`brew install sops` or download from GitHub)
   - GPG or Age for testing

2. **Clone and Build**
   ```bash
   git clone https://github.com/danpet-dev/keyforge
   cd keyforge
   go mod download
   go build -o keyforge ./cmd/keyforge
   ```

3. **Run Tests**
   ```bash
   go test ./...
   ```

## Code Structure

```
keyforge/
├── cmd/keyforge/       # CLI commands (Cobra)
├── pkg/
│   ├── sops/          # SOPS wrapper functions
│   ├── config/        # .sops.yaml handling
│   └── keys/          # PGP/Age key operations
└── docs/              # Documentation
```

## Guidelines

- **Code Style**: Follow standard Go conventions (`gofmt`, `golint`)
- **Tests**: Add tests for new features
- **Commits**: Use conventional commits (e.g., `feat:`, `fix:`, `docs:`)
- **Documentation**: Update README and docs/ when adding features

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Add tests
5. Run `go test ./...` and `gofmt -w .`
6. Commit your changes (`git commit -m 'feat: add my feature'`)
7. Push to your fork (`git push origin feature/my-feature`)
8. Open a Pull Request

## Reporting Issues

Use GitHub Issues to report bugs or request features. Please include:
- KeyForge version
- Operating system
- Steps to reproduce
- Expected vs. actual behavior

## Questions?

Open a GitHub Discussion or reach out via the issue tracker.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
