# GitHub Publishing Guide for KeyForge v0.2.0

## Prerequisites

You need a GitHub account. If you don't have one:
1. Go to https://github.com/signup
2. Create an account with username `danpet-dev` (or adjust all commands below)

## Step 1: Create GitHub Repository

### Via GitHub Web UI (Recommended)

1. Go to https://github.com/new
2. Fill in the form:
   - **Repository name:** `keyforge`
   - **Description:** `KeyForge - SOPS multi-key lifecycle management CLI`
   - **Visibility:** Public
   - **DO NOT** initialize with README, .gitignore, or license (we already have these)
3. Click "Create repository"
4. **Keep the page open** - you'll need the URLs shown

### Via GitHub CLI (Alternative)

If you install `gh`:
```bash
# Install gh (Ubuntu/Debian)
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install gh

# Authenticate
gh auth login

# Create repo
cd ~/repos/keyforge
gh repo create danpet-dev/keyforge --public --source=. --remote=origin
```

## Step 2: Add GitHub as Remote

```bash
cd ~/repos/keyforge

# Add remote (replace YOUR_USERNAME if not danpet-dev)
git remote add origin https://github.com/danpet-dev/keyforge.git

# Verify
git remote -v
```

## Step 3: Push Code and Tags

```bash
cd ~/repos/keyforge

# Push main branch
git push -u origin main

# Push all tags
git push origin --tags

# Verify
git log --oneline --graph --decorate -5
```

## Step 4: Verify GitHub Actions

After pushing, GitHub Actions will automatically:
1. Run tests (`.github/workflows/test.yml`)
2. Run linter (`.github/workflows/lint.yml`)
3. Build binaries on tag push (`.github/workflows/release.yml`)

Check status at: https://github.com/danpet-dev/keyforge/actions

## Step 5: Create GitHub Release

### Automatic (via Goreleaser - if v0.2.0 tag triggers workflow)

The release workflow should automatically create a release with:
- Compiled binaries for Linux, macOS, Windows (amd64, arm64)
- Checksums
- CHANGELOG

### Manual (if automatic fails)

1. Go to https://github.com/danpet-dev/keyforge/releases/new
2. Select tag: `v0.2.0`
3. Release title: `KeyForge v0.2.0 - Age Support, Key Rotation, Audit`
4. Copy content from `V0.2-RELEASE-NOTES.md`
5. Attach any pre-built binaries (optional)
6. Click "Publish release"

## Step 6: Setup Docker Image Publishing (Optional)

For Docker images to publish to GitHub Container Registry:

1. Go to https://github.com/settings/tokens
2. Create a Personal Access Token with scopes:
   - `write:packages`
   - `read:packages`
3. Add secret to repository:
   - Go to: https://github.com/danpet-dev/keyforge/settings/secrets/actions
   - Name: `GHCR_TOKEN`
   - Value: Your PAT

Then Docker images will publish on release.

## Step 7: Verify Everything Works

```bash
# Test installation from GitHub
go install github.com/danpet-dev/keyforge/cmd/keyforge@v0.2.0

# Test Docker image (after release)
docker pull ghcr.io/danpet-dev/keyforge:v0.2.0
docker run --rm ghcr.io/danpet-dev/keyforge:v0.2.0 --help
```

## Expected Repository State After Push

```
https://github.com/danpet-dev/keyforge
├── 📁 .github/workflows/    (CI/CD workflows)
├── 📁 cmd/keyforge/          (CLI source)
├── 📁 pkg/                   (Packages: config, keys, sops)
├── 📄 .goreleaser.yml        (Release automation)
├── 📄 Dockerfile             (Docker image)
├── 📄 README.md              (Documentation)
├── 📄 CHANGELOG.md           (Version history)
├── 📄 V0.2-RELEASE-NOTES.md  (Release notes)
├── 📄 LICENSE                (MIT)
└── 🏷️  Tags: v0.1.0, v0.2.0
```

## Troubleshooting

### Problem: "remote origin already exists"
```bash
git remote remove origin
git remote add origin https://github.com/danpet-dev/keyforge.git
```

### Problem: Authentication failed
```bash
# Use SSH instead
git remote set-url origin git@github.com:danpet-dev/keyforge.git

# Or configure Git credentials
git config --global credential.helper store
```

### Problem: Goreleaser fails
Check GitHub Actions logs at:
https://github.com/danpet-dev/keyforge/actions

Common issues:
- Missing `GITHUB_TOKEN` (auto-provided by GitHub Actions)
- Invalid `.goreleaser.yml` syntax
- Docker login issues (need `GHCR_TOKEN`)

## Next Steps After Publishing

1. ✅ Verify GitHub Actions pass
2. ✅ Download and test release binaries
3. ✅ Test Docker image
4. 📢 Announce in NornForge documentation
5. 🔄 Update NornForge to use KeyForge from GitHub

---

**Ready to publish?** Run the commands in Step 2 and Step 3!
