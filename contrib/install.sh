#!/bin/sh
set -e

TAR_FILE="/tmp/ansible-requirements-lint.tar.gz"
RELEASES_URL="https://github.com/atosatto/ansible-requirements-lint/releases"
test -z "$TMPDIR" && TMPDIR="$(mktemp -d)"

last_version() {
  curl -sL -o /dev/null -w %{url_effective} "$RELEASES_URL/latest" | 
    rev | 
    cut -f1 -d'/'| 
    rev
}

download() {
  test -z "$VERSION" && VERSION="$(last_version)"
  test -z "$VERSION" && {
    echo "Unable to get ansible-requirements-lint version." >&2
    exit 1
  }
  rm -f "$TAR_FILE"
  curl -s -L -o "$TAR_FILE" \
    "$RELEASES_URL/download/$VERSION/ansible-requirements-lint_${VERSION#v}_$(uname -s)_$(uname -m).tar.gz"
}

download
tar -xf "$TAR_FILE" -C "$TMPDIR"
mv "${TMPDIR}/ansible-requirements-lint" .
