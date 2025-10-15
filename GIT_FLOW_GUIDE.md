# Git Flow Guide for GitVersion-Go

This project uses Git Flow workflow with Makefile automation for streamlined development.

## Branch Structure

- **main**: Production-ready code, tagged releases
- **develop**: Integration branch for features
- **feature/**: Feature development branches
- **release/**: Release preparation branches  
- **hotfix/**: Emergency fixes for production

## Quick Start Commands

### Feature Development
```bash
# Start new feature
make git-feature-start FEATURE=awesome-feature

# Work on your feature...
# When ready to merge to develop:
make git-feature-finish FEATURE=awesome-feature
```

### Release Process
```bash
# Start release from develop
make git-release-start VERSION=1.2.0

# Test and prepare release...
# When ready to release:
make git-release-finish VERSION=1.2.0
```

### Hotfix Process
```bash
# Start hotfix from main
make git-hotfix-start VERSION=1.2.1

# Fix the issue...
# When ready to release:
make git-hotfix-finish VERSION=1.2.1
```

## Best Practices

### For Features
1. Always start features from `develop` branch
2. Use descriptive feature names: `user-authentication`, `api-improvements`
3. Run `make quality` before finishing features
4. Features get merged to `develop`, never directly to `main`

### For Releases
1. **Start from develop**: Releases are created from the `develop` branch
2. **Version format**: Use semantic versioning (e.g., `1.2.0`, `2.0.0-alpha`)
3. **Release branch purpose**: Final testing, documentation, version bumps
4. **Automatic version update**: The Makefile automatically updates `gitversion.yml`
5. **Dual merge**: Release gets merged to both `main` (with tag) and back to `develop`

### For Hotfixes
1. **Start from main**: Hotfixes are created from the `main` branch
2. **Critical fixes only**: Use for urgent production issues
3. **Version increment**: Typically increment patch version (e.g., `1.2.0` → `1.2.1`)
4. **Dual merge**: Hotfix gets merged to both `main` (with tag) and back to `develop`

## Development Workflow

### Daily Development
```bash
# Quick development checks
make dev

# Complete development workflow (includes pre-commit)
make dev-full

# Check current status
make git-status

# Sync with remote
make git-sync
```

### Pre-commit Integration
```bash
# Install pre-commit hooks (one-time setup)
make pre-commit-install

# Run pre-commit on all files
make pre-commit

# Update pre-commit hooks
make pre-commit-update
```

## Makefile Targets Overview

### Git Flow Commands
- `git-feature-start FEATURE=name` - Start new feature branch
- `git-feature-finish FEATURE=name` - Merge feature to develop
- `git-release-start VERSION=x.y.z` - Start release branch from develop
- `git-release-finish VERSION=x.y.z` - Finish release (merge to main+develop, tag)
- `git-hotfix-start VERSION=x.y.z` - Start hotfix branch from main
- `git-hotfix-finish VERSION=x.y.z` - Finish hotfix (merge to main+develop, tag)

### Utility Commands
- `git-status` - Show branches and status
- `git-sync` - Sync current branch with remote
- `git-merge-to-develop` - Merge current branch to develop
- `git-merge-to-main` - Merge current branch to main (use with caution)

### Quality Commands
- `pre-commit` - Run pre-commit hooks on all files
- `quality` - Run all quality checks (format, vet, lint, test)
- `dev` - Quick development workflow
- `dev-full` - Complete development workflow with pre-commit

## Example Workflows

### Adding a New Feature
```bash
# 1. Start feature
make git-feature-start FEATURE=json-output

# 2. Develop your feature...
# Make changes, commit regularly

# 3. Before finishing, run quality checks
make dev-full

# 4. Finish feature (merges to develop)
make git-feature-finish FEATURE=json-output

# 5. Clean up (optional)
git branch -d feature/json-output
git push origin --delete feature/json-output
```

### Creating a Release
```bash
# 1. Start release from develop
make git-release-start VERSION=1.0.0

# 2. Final testing and preparation
make quality
# Update documentation, changelog, etc.

# 3. Finish release
make git-release-finish VERSION=1.0.0

# 4. Clean up (optional)
git branch -d release/1.0.0
git push origin --delete release/1.0.0
```

### Emergency Hotfix
```bash
# 1. Start hotfix from main
make git-hotfix-start VERSION=1.0.1

# 2. Fix the critical issue
# Make minimal changes needed

# 3. Test thoroughly
make quality

# 4. Finish hotfix
make git-hotfix-finish VERSION=1.0.1

# 5. Clean up (optional)
git branch -d hotfix/1.0.1
git push origin --delete hotfix/1.0.1
```

## Version Management

The release process automatically updates the `next-version` field in `gitversion.yml`:

```yaml
next-version: "1.2.0"  # Updated automatically by make git-release-start
```

This ensures that the GitVersion tool uses the correct version during the release process.

## Safety Features

- **Quality gates**: All finish commands run `make quality` first
- **Confirmation prompts**: Critical operations like merging to main require confirmation
- **Automatic syncing**: Commands automatically pull latest changes before proceeding
- **Pre-commit integration**: Ensures code quality before commits

## Need Help?

Run `make help` to see all available commands with descriptions and examples.