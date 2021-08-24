#!/bin/sh
set -e

TAR_FILE="/tmp/goreleaser.tar.gz"
RELEASES_URL="https://github.com/goreleaser/goreleaser/releases"
PINNED_VERSION="v0.173.2"

test -z "$TMPDIR" && TMPDIR="$(mktemp -d)"

pinned_version() {
  curl -sL -o /dev/null -w %{url_effective} "$RELEASES_URL/$PINNED_VERSION" |
    rev |
    cut -f1 -d'/'|
    rev
}

download() {
  test -z "$VERSION" && VERSION="$(pinned_version)"
  test -z "$VERSION" && {
    echo "Unable to get goreleaser version." >&2
    exit 1
  }
  rm -f "$TAR_FILE"
  curl -s -L -o "$TAR_FILE" \
    "$RELEASES_URL/download/$VERSION/goreleaser_$(uname -s)_$(uname -m).tar.gz"
}

download
tar -xf "$TAR_FILE" -C "$TMPDIR"
"${TMPDIR}/goreleaser" "$@"