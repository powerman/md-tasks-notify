#!/bin/bash
set -euo pipefail

target_branch="${RELEASE_PLZ_TARGET_BRANCH:-master}"
pr_branch="${RELEASE_PLZ_PR_BRANCH:-release-plz}"
commit_prefix="${RELEASE_PLZ_COMMIT_PREFIX:-chore: release}"

{
    echo "TARGET_BRANCH=$target_branch"
    echo "PR_BRANCH=$pr_branch"
    echo "COMMIT_PREFIX=$commit_prefix"
} >>"$GITHUB_ENV"

echo
echo "Release-plz configuration:"
echo
echo "  RELEASE_PLZ_TARGET_BRANCH = '$target_branch'"
echo "  RELEASE_PLZ_PR_BRANCH     = '$pr_branch'"
echo "  RELEASE_PLZ_COMMIT_PREFIX = '$commit_prefix'"
echo
echo "Configure at: Settings → Secrets and variables → Actions → Variables"
echo
