# Release Process

This document describes the release process for vibe.

## Prerequisites

1. **GitHub Personal Access Token**
   - Create a token with `repo` and `write:packages` scopes
   - Add as `GITHUB_TOKEN` (automatically available in GitHub Actions)

2. **Homebrew Tap Token** (for Homebrew releases)
   - Create a token with `repo` scope
   - Add as repository secret: `HOMEBREW_TAP_GITHUB_TOKEN`
   - Create a tap repository: `github.com/rithyhuot/homebrew-vibe`

3. **GoReleaser Installed Locally** (for testing)

   ```bash
   brew install goreleaser
   ```

## Automated Release Process

Releases are automated via GitHub Actions when a version tag is pushed.

### Step 1: Prepare the Release

1. **Update version references** (if needed)

   ```bash
   # Update CHANGELOG.md with release notes
   vim CHANGELOG.md
   ```

2. **Run tests locally**

   ```bash
   make test
   make lint
   ```

3. **Test GoReleaser locally** (optional)

   ```bash
   goreleaser release --snapshot --clean --skip=publish
   ```

### Step 2: Create and Push Tag

1. **Create version tag**

   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   ```

2. **Push tag to trigger release**

   ```bash
   git push origin v0.1.0
   ```

### Step 3: Verify Release

1. **Check GitHub Actions**
   - Go to: <https://github.com/rithyhuot/vibe/actions>
   - Verify the "Release" workflow completes successfully

2. **Check GitHub Release**
   - Go to: <https://github.com/rithyhuot/vibe/releases>
   - Verify release was created with:
     - Binary artifacts for all platforms
     - Checksums
     - Changelog
     - Docker images

3. **Check Homebrew Tap** (if configured)
   - Go to: <https://github.com/rithyhuot/homebrew-vibe>
   - Verify formula was updated

4. **Check Docker Images**
   - Go to: <https://github.com/rithyhuot/vibe/pkgs/container/vibe>
   - Verify images were pushed

## What Gets Released

### Binary Artifacts

GoReleaser builds binaries for:

- **Linux**: amd64, arm64, arm (v6, v7)
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

### Distribution Channels

1. **GitHub Releases**
   - Downloadable binary archives (.tar.gz for Unix, .zip for Windows)
   - Checksums file
   - Changelog

2. **Docker Images** (GitHub Container Registry)
   - `ghcr.io/rithyhuot/vibe:latest`
   - `ghcr.io/rithyhuot/vibe:v0.1.0`
   - Multi-arch support (amd64, arm64)

3. **Homebrew** (via tap)
   - `brew tap rithyhuot/vibe`
   - `brew install vibe`

## Manual Release (Emergency)

If automated release fails, you can release manually:

```bash
# Ensure you're on the tagged commit
git checkout v0.1.0

# Set required environment variables
export GITHUB_TOKEN="your_token_here"
export HOMEBREW_TAP_GITHUB_TOKEN="your_homebrew_token_here"

# Run GoReleaser
goreleaser release --clean
```

## Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality in a backwards compatible manner
- **PATCH**: Backwards compatible bug fixes

Examples:

- `v0.1.0` - Initial release
- `v0.1.1` - Bug fix release
- `v0.2.0` - New features
- `v1.0.0` - First stable release

## Pre-releases

For beta or release candidate versions:

```bash
git tag -a v0.2.0-beta.1 -m "Release v0.2.0-beta.1"
git push origin v0.2.0-beta.1
```

GoReleaser will automatically mark these as pre-releases on GitHub.

## Rollback

To rollback a release:

1. **Delete the GitHub release**
   - Go to: <https://github.com/rithyhuot/vibe/releases>
   - Delete the problematic release

2. **Delete the tag**

   ```bash
   git tag -d v0.1.0
   git push origin :refs/tags/v0.1.0
   ```

3. **Create a new patch release**

   ```bash
   git tag -a v0.1.1 -m "Release v0.1.1 (fixes v0.1.0)"
   git push origin v0.1.1
   ```

## Changelog Generation

GoReleaser automatically generates changelogs from commit messages. To ensure good changelogs:

### Commit Message Format

Use conventional commits:

```
<type>(<scope>): <subject>

<body>
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `perf`: Performance improvement
- `refactor`: Code refactoring
- `docs`: Documentation changes
- `test`: Test changes
- `chore`: Build process or tooling changes
- `ci`: CI/CD changes

Examples:

```
feat(pr): add AI-powered PR descriptions
fix(branch): handle special characters in branch names
perf(cache): improve sprint detection caching
docs(readme): update installation instructions
```

## Post-Release Checklist

After a successful release:

- [ ] Verify all binaries work on target platforms
- [ ] Test Homebrew installation
- [ ] Test Docker image
- [ ] Update documentation if needed
- [ ] Announce release (if applicable)
- [ ] Close milestone (if using GitHub milestones)

## Troubleshooting

### Release workflow fails

1. **Check GitHub Actions logs**
   - Look for specific error messages
   - Common issues:
     - Missing secrets
     - GoReleaser configuration errors
     - Docker build failures

2. **Test locally**

   ```bash
   goreleaser release --snapshot --clean --skip=publish
   ```

3. **Verify secrets**
   - Ensure `GITHUB_TOKEN` has correct permissions
   - Ensure `HOMEBREW_TAP_GITHUB_TOKEN` is set (if using Homebrew)

### Homebrew formula not updated

1. **Check tap repository exists**
   - Verify: <https://github.com/rithyhuot/homebrew-vibe>

2. **Check token permissions**
   - `HOMEBREW_TAP_GITHUB_TOKEN` needs `repo` scope

3. **Manually update formula** (if needed)
   - Go to tap repository
   - Update Formula/vibe.rb with new version and checksums

## Contact

For questions about the release process:

- Open an issue: <https://github.com/rithyhuot/vibe/issues>
- Review existing releases: <https://github.com/rithyhuot/vibe/releases>
