#!/bin/sh
MARKER_PREFIX="##"
VERSION="$1"

IFS=''
found=0

# if version is a pre release, print Unreleased changelog
if [[ $VERSION == *-pre\.* ]]; then
    VERSION="Unreleased"
fi

# This script prints a subset of changelog matching the version entry
cat CHANGELOG.md | while read "line"; do

    # If not found and matching heading
    if [ $found -eq 0 ] && echo "$line" | grep -q "^$MARKER_PREFIX \[$VERSION\]"; then
        found=1
        continue
    fi

    # If needed version if found, and reaching next delimiter - stop
    if [ $found -eq 1 ] && echo "$line" | grep -q -E "^$MARKER_PREFIX \[[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]\]+"; then
        found=0
        break
    fi

    # Keep printing out lines as no other version delimiter found
    if [ $found -eq 1 ]; then
        echo "$line"
    fi
done
