#!/bin/bash
# Get revision of latest tag
LATEST_TAG_REV=$(git rev-list --tags --max-count=1)
# Get revision of latest commit
LATEST_COMMIT_REV=$(git rev-list HEAD --max-count=1)

# Get latest tag name, 0.0.0 if missing
if [ -n "$LATEST_TAG_REV" ]; then
    LATEST_TAG=$(git describe --tags "$LATEST_TAG_REV")
else
    LATEST_TAG="0.0.0"
fi

# Strip 'v' prefix
LATEST_TAG=${LATEST_TAG#v}

if [ "$LATEST_TAG_REV" != "$LATEST_COMMIT_REV" ]; then
    # If revision for latest commit doesn't match revision of latest tag,
    # it means we are creating a prerelease.

    # Strip any existing prerelease from tag
    LATEST_TAG=$(echo "$LATEST_TAG" | cut -d "-" -f1)

    # Increment the patch version (0.0.2 -> 0.0.3)
    LATEST_TAG=$(echo "$LATEST_TAG" | awk -F. -v OFS=. 'NF==1{print ++$NF}; NF>1{if(length($NF+1)>length($NF))$(NF-1)++; $NF=sprintf("%0*d", length($NF), ($NF+1)%(10^length($NF))); print}')

    # Append prerelease tag and commit number (0.0.3-pre.123)
    echo "$LATEST_TAG-pre.$(git rev-list --count HEAD)"
else
    # Otherwise, echo version in semver format (X.X.X)
    echo "$LATEST_TAG"
fi
