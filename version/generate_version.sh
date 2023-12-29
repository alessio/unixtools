#!/bin/sh
SELF_DIR=$(SELF=$(dirname "$0") && sh -c "cd \"$SELF\" && pwd")
GIT_VERSION=$(git describe --abbrev=6 2>/dev/null || echo "v0.0.0-UNKNOWN")
echo "${GIT_VERSION}" > "${SELF_DIR}/version.txt"
