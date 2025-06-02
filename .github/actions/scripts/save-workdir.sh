#!/bin/bash
set -euo pipefail

stash_name="${1:-release-plz-stash}"

# Save original ref and branch information
original_ref=$(git rev-parse HEAD)
original_branch=$(git branch --show-current 2>/dev/null || echo "")

# Export to environment for restore script
echo "ORIGINAL_REF=$original_ref" >> "$GITHUB_ENV"
echo "ORIGINAL_BRANCH=$original_branch" >> "$GITHUB_ENV"
echo "STASH_NAME=$stash_name" >> "$GITHUB_ENV"

# Stash any uncommitted changes (supports detached HEAD)
git stash push -u -m "$stash_name"

echo "Saved workdir state: ref=$original_ref, branch=$original_branch, stash=$stash_name"