#!/usr/bin/env bash
# PortWatch installer for linux/darwin.
#
# Usage:
#   ./install.sh                       install the latest release
#   PORTWATCH_VERSION=v0.7.0 ./install.sh   install a pinned release tag
#   ./install.sh --uninstall           remove portwatch from ~/.local/bin
#
# Downloaded from GitHub releases of https://github.com/MY-Final/portWatch.
set -euo pipefail

REPO_URL="https://github.com/MY-Final/portWatch"
API_BASE="https://api.github.com/repos/MY-Final/portWatch"
VERSION="${PORTWATCH_VERSION:-latest}"
INSTALL_DIR="$HOME/.local/bin"
UNINSTALL=0

say() { printf '==> %s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

for arg in "$@"; do
  case "$arg" in
    --uninstall) UNINSTALL=1 ;;
    -h|--help)
      echo "usage: ./install.sh [--uninstall]   (PORTWATCH_VERSION=v0.7.0 pins a release)"
      exit 0
      ;;
    *) die "unknown argument: $arg (supported: --uninstall)" ;;
  esac
done

if [ "$UNINSTALL" -eq 1 ]; then
  removed_any=0
  for name in portwatch pw; do
    target="$INSTALL_DIR/$name"
    if [ ! -f "$target" ]; then
      continue
    fi
    say "Deleting $target"
    rm -f "$target" || die "failed to delete $target; close running portwatch processes and retry"
    removed_any=1
  done
  if [ "$removed_any" -eq 0 ]; then
    echo "portwatch not found in $INSTALL_DIR (already uninstalled)."
    exit 0
  fi
  if [ -z "$(ls -A "$INSTALL_DIR" 2>/dev/null)" ]; then
    # rmdir only removes empty directories, so other tools in ~/.local/bin
    # keep it alive and the failure is silently ignored.
    if rmdir "$INSTALL_DIR" 2>/dev/null; then
      echo "Removed empty directory $INSTALL_DIR."
    fi
  else
    echo "If $INSTALL_DIR was added to PATH in your shell profile, remove that line too."
  fi
  echo "portwatch uninstalled."
  exit 0
fi

command -v curl >/dev/null 2>&1 || die "curl is required but not installed"

OS=$(case "$(uname -s)" in Darwin*) echo darwin ;; *) echo linux ;; esac)
case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) die "unsupported architecture: $(uname -m)" ;;
esac

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  else
    die "no sha256 tool found (need sha256sum or shasum)"
  fi
}

# Print the download URL of the release asset whose name ends in $1.
asset_url() {
  printf '%s' "$RELEASE_JSON" \
    | grep -oE '"browser_download_url"[[:space:]]*:[[:space:]]*"[^"]+"' \
    | sed 's/.*"\(https[^"]*\)"$/\1/' \
    | grep -E "$1$" \
    | head -n 1
}

say "Resolving release ($VERSION)"
if [ "$VERSION" = "latest" ]; then
  RELEASE_JSON=$(curl -fsSL "$API_BASE/releases/latest") \
    || die "could not query $API_BASE/releases/latest"
else
  case "$VERSION" in
    v*) TAG="$VERSION" ;;
    *) TAG="v$VERSION" ;;
  esac
  RELEASE_JSON=$(curl -fsSL "$API_BASE/releases/tags/$TAG") \
    || die "could not query release $TAG"
fi

ASSET_URL=$(asset_url "${OS}_${ARCH}.tar.gz")
[ -n "$ASSET_URL" ] || die "release has no ${OS}_${ARCH}.tar.gz asset; check $REPO_URL/releases"
CHECKSUM_URL=$(asset_url 'checksums\.txt')

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

say "Downloading $ASSET_URL"
curl -fsSL -o "$TMP_DIR/portwatch.tar.gz" "$ASSET_URL" || die "download failed: $ASSET_URL"

if [ -n "$CHECKSUM_URL" ]; then
  curl -fsSL -o "$TMP_DIR/checksums.txt" "$CHECKSUM_URL" || die "download failed: $CHECKSUM_URL"
  expected=$(grep -E "[ _]${OS}_${ARCH}\.tar\.gz$" "$TMP_DIR/checksums.txt" | awk '{print $1}' | head -n 1)
  [ -n "$expected" ] || die "checksums.txt has no entry for the ${OS}_${ARCH} archive"
  actual=$(sha256_of "$TMP_DIR/portwatch.tar.gz")
  [ "$actual" = "$expected" ] || die "SHA256 mismatch: expected $expected, got $actual"
  echo "Checksum verified."
else
  echo "warning: release has no checksums.txt; skipping verification" >&2
fi

mkdir -p "$TMP_DIR/extract"
tar -xzf "$TMP_DIR/portwatch.tar.gz" -C "$TMP_DIR/extract"
BIN=$(find "$TMP_DIR/extract" -type f -name portwatch | head -n 1)
[ -n "$BIN" ] || die "archive contains no portwatch binary"

mkdir -p "$INSTALL_DIR"
mv -f "$BIN" "$INSTALL_DIR/portwatch" || die "failed to install into $INSTALL_DIR"
# Short alias: the binary picks its name from argv[0].
cp -f "$INSTALL_DIR/portwatch" "$INSTALL_DIR/pw"
chmod +x "$INSTALL_DIR/portwatch" "$INSTALL_DIR/pw"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo "note: $INSTALL_DIR is not in PATH; add it with:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    ;;
esac

say "Installed"
"$INSTALL_DIR/portwatch" --version
echo "Installed to $INSTALL_DIR/portwatch"
