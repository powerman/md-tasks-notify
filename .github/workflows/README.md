# Release PLZ Workflows

This directory contains reusable GitHub Actions composite actions for automated release
management using `release-plz` approach.

## Overview

The release-plz system provides automated version bumping, changelog generation, and release
creation based on conventional commits. It consists of composite actions that can be easily
integrated into any project's workflows.

## Quick Start

### 1. Copy Required Files

Copy these files to your project's `.github/workflows/` directory:

```
release-plz.yml
release-plz-pr.yml
release-plz-pr-update.yml
```

### 2. Update Configuration

Configure repository variables in **Settings → Secrets and variables → Actions → Variables**:

| Variable Name | Default Value | Description |
|---------------|---------------|-------------|
| `RELEASE_PLZ_TARGET_BRANCH` | `master` | Target branch for releases |
| `RELEASE_PLZ_PR_BRANCH` | `release-plz` | Technical branch for release PRs |
| `RELEASE_PLZ_COMMIT_PREFIX` | `chore: release` | Commit message prefix |

**Repository owner**: Update `github.repository_owner == 'powerman'` to your username in workflow `if:` conditions.

**Trigger branches**: Update `branches:` in workflow triggers to match your target branch:

```yaml
on:
  push:
    branches:
      - main  # Must match RELEASE_PLZ_TARGET_BRANCH variable
```

**Note**: If you don't set these variables, defaults will be used automatically.

### 3. Required Repository Settings

In **Settings → Actions → General**:

- ✅ Allow GitHub Actions to create and approve pull requests

### 4. Required Tools

Add to your project's tool configuration (e.g., `mise.toml`):

```toml
[tools]
git-cliff = 'latest' # Changelog generator
gh = 'latest'        # GitHub CLI
```

### 5. Changelog Configuration

Create `cliff.toml` in your project root for changelog generation. See [git-cliff
documentation](https://git-cliff.org/docs/configuration) for details.

## Workflow Files

### `release-plz.yml`

Main release workflow that:

- Extracts version from commit messages
- Handles race conditions
- Creates draft releases
- Builds and uploads assets
- Signs assets with cosign (optional)
- Publishes final release

**Triggers**: Push to master branch with CHANGELOG.md changes

### `release-plz-pr.yml`

Creates/updates release pull requests based on conventional commits.

**Triggers**: Push to master branch (non-release commits)

### `release-plz-pr-update.yml`

Handles manual version edits in release PR titles.

**Triggers**: PR title edits on release-plz branch

## Composite Actions

### `release-plz-prepare`

Extracts version, checks for race conditions, and creates draft release.

**Inputs:**
- `version_cmd` (optional): Shell command for additional version updates

**Configuration**: Uses repository variables (set once via GITHUB_ENV):
- `RELEASE_PLZ_PR_BRANCH` (default: "release-plz")
- `RELEASE_PLZ_COMMIT_PREFIX` (default: "chore: release")  
- `RELEASE_PLZ_TARGET_BRANCH` (default: "master")

**Outputs:**

- `version`: Extracted version
- `changelog`: Generated changelog

**Requirements:**

- Repository must be checked out (any fetch-depth, action handles git history automatically)
- Git user must be configured for commits
- mise-action or equivalent tool setup

**Working Directory Preservation:**

- Original branch and uncommitted changes are preserved via git stash
- Workdir is restored to original state after action completion

**Version Command Execution:**

- Runs in clean workdir on release-plz branch during race condition handling
- Original repository state preserved via git stash

### `release-plz-finalize`

Uploads assets, optionally signs them, and finalizes the release.

**Inputs:**

- `version` (required): Release version
- `changelog` (required): Release changelog
- `assets_directory` (optional): Directory containing release assets
- `cosign` (optional): Enable cosign signing (default: false)

### `release-plz-pr`

Creates or updates release pull requests based on conventional commits.

**Inputs:**
- `version_cmd` (optional): Shell command to update additional files with version

**Configuration**: Uses repository variables (set once via GITHUB_ENV):
- `RELEASE_PLZ_PR_BRANCH` (default: "release-plz")
- `RELEASE_PLZ_COMMIT_PREFIX` (default: "chore: release")
- `RELEASE_PLZ_TARGET_BRANCH` (default: "master")

**Requirements:**

- Repository must be checked out (any fetch-depth, action handles git history automatically)
- Git user must be configured for commits and pushes

**Working Directory Preservation:**

- Original branch/ref and uncommitted changes are preserved via git stash
- Supports detached HEAD state
- Workdir is restored to original state after action completion

**Version Command Execution:**

- Runs in clean workdir on release-plz branch before CHANGELOG.md generation
- Ensures consistent version updates across all modified files

### `release-plz-pr-update`

Updates release-plz branch when PR title is manually edited.

**Inputs:**
- `version_cmd` (optional): Shell command to update additional files with version

**Configuration**: Uses repository variables (set once via GITHUB_ENV):
- `RELEASE_PLZ_PR_BRANCH` (default: "release-plz")  
- `RELEASE_PLZ_COMMIT_PREFIX` (default: "chore: release")

**Requirements:**

- Repository must be checked out (action handles branch switching and git history automatically)
- Git user must be configured for commits and pushes

**Working Directory Preservation:**

- Original branch/ref and uncommitted changes are preserved via git stash
- Supports detached HEAD state
- Workdir is restored to original state after action completion

**Automatic Context Detection:**

- PR title and number are automatically extracted from `github.event`
- Automatically switches to correct branch if needed
- No need to pass PR details as inputs

**Version Command Execution:**

- Runs in clean workdir when version changes are detected
- Executes before CHANGELOG.md regeneration and commit amendment

## Customization Examples

### Custom Version Command

Update additional files with the new version:

```yaml
# Configure repository variables in Settings → Secrets and variables → Actions → Variables:
# RELEASE_PLZ_TARGET_BRANCH = main
# RELEASE_PLZ_PR_BRANCH = release-plz  
# RELEASE_PLZ_COMMIT_PREFIX = chore: release

# In workflow if condition (uses vars):
if: >-
  ${{
    github.repository_owner == 'myorg' &&
    ! contains(github.event.head_commit.message, vars.RELEASE_PLZ_COMMIT_PREFIX) &&
    ! (contains(github.event.head_commit.message, 'Merge pull request') &&
       contains(github.event.head_commit.message, format('from {0}', vars.RELEASE_PLZ_PR_BRANCH)))
  }}

# In workflow steps (no inputs needed - uses vars automatically):
- uses: ./.github/actions/release-plz-pr
  with:
    version_cmd: |
      sed -i "s/version = \".*\"/version = \"${RELEASE_PLZ_VERSION#v}\"/" Cargo.toml
      sed -i "s/__version__ = \".*\"/__version__ = \"${RELEASE_PLZ_VERSION#v}\"/" src/__init__.py
```

**Important**: The `version_cmd` is executed in the proper context:

- **Working directory**: Clean workdir on `release-plz` branch (or new branch being created)
- **Timing**: Before CHANGELOG.md is generated and committed
- **Environment**: `$RELEASE_PLZ_VERSION` contains the full version (e.g., "v1.2.3")
- **State**: Original repository state is preserved in git stash

This ensures that:

1. Version files are updated consistently across all workflows
2. CHANGELOG.md generation includes the version updates
3. No conflicts with existing working directory changes
4. Race condition detection works correctly

### Custom Build Steps

Replace the build section in `release-plz.yml`:

```yaml
# For Rust projects
- name: Build release binaries
  run: |
    cargo build --release
    mkdir -p ./dist
    cp target/release/myapp ./dist/

# For Node.js projects
- name: Build and package
  run: |
    npm ci
    npm run build
    npm pack --pack-destination ./dist
```

### Different Programming Languages

#### Python Project Example

```yaml
# No fetch-depth needed - action handles git history automatically
- uses: actions/checkout@v4

- uses: actions/setup-python@v5
  with:
    python-version: "3.11"

- name: Build Python package
  run: |
    pip install build
    python -m build --outdir ./dist

- uses: ./.github/actions/release-plz-prepare
  id: prepare

- uses: ./.github/actions/release-plz-finalize
  with:
    version: ${{ steps.prepare.outputs.version }}
    changelog: ${{ steps.prepare.outputs.changelog }}
    assets_directory: ./dist
```

#### Rust Project Example

```yaml
# No fetch-depth needed - action handles git history automatically
- uses: actions/checkout@v4

- uses: dtolnay/rust-toolchain@stable

- name: Build Rust binary
  run: |
    cargo build --release
    mkdir -p ./dist
    cp target/release/myapp ./dist/

- uses: ./.github/actions/release-plz-prepare
  id: prepare
  with:
    version_cmd: 'sed -i "s/^version = \".*\"/version = \"${RELEASE_PLZ_VERSION#v}\"/" Cargo.toml'

- uses: ./.github/actions/release-plz-finalize
  with:
    version: ${{ steps.prepare.outputs.version }}
    changelog: ${{ steps.prepare.outputs.changelog }}
    assets_directory: ./dist
    cosign: true
```

### Enable Asset Signing

```yaml
- uses: ./.github/actions/release-plz-finalize
  with:
    version: ${{ needs.release.outputs.version }}
    changelog: ${{ needs.release.outputs.changelog }}
    assets_directory: ./dist
    cosign: true
```

## Testing with Act

Test workflows locally using [act](https://github.com/nektos/act):

```bash
# Test release PR creation
act push -e test-events/push.json

# Test release workflow
act push -e test-events/release-push.json
```

The workflows automatically handle git remote setup for act when `$ACT` environment variable
is detected.

## Race Condition Handling

The system automatically handles race conditions where:

1. User manually changes version in release PR title
2. New commits are pushed while release PR is pending

In such cases, a new release PR is created instead of proceeding with potentially incorrect
release.

## Repository Variables Benefits

Using GitHub repository variables (`vars.*`) instead of workflow `env:` provides several advantages:

### **Centralized Configuration**
- Variables are set once in **Settings → Secrets and variables → Actions → Variables**
- All workflows automatically use the same values
- No risk of inconsistency between different workflow files

### **True Defaults**
- If variables are not set, built-in defaults are used automatically
- No need to configure anything for basic usage
- Variables only need to be set when customization is required

**Runtime Visibility**
- Each action shows current configuration with exact variable names
- Easy to verify what settings are being used (e.g., "RELEASE_PLZ_TARGET_BRANCH: master")
- Clear indication of where to change values if needed
- Variables are set once via GITHUB_ENV and available to all subsequent steps

**Simplified Maintenance**
- No `env:` sections to maintain in workflow files
- No action inputs to pass configuration
- Variables set once per action via GITHUB_ENV
- Less boilerplate code in workflows

This approach reduces setup complexity while providing better visibility and consistency across all release-plz workflows.
