#!/bin/bash
set -euo pipefail

# Restore original ref/branch
if [ -n "${ORIGINAL_BRANCH:-}" ]; then
    git checkout "$ORIGINAL_BRANCH"
else
    git checkout "${ORIGINAL_REF:-HEAD}"
fi

# Restore stashed changes if they exist
stash_name="${STASH_NAME:-release-plz-stash}"
if git stash list | grep -q "$stash_name"; then
    git stash pop
fi

echo "Restored workdir state: ref=${ORIGINAL_REF:-HEAD}, branch=${ORIGINAL_BRANCH:-}, stash=$stash_name"