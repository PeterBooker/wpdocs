#!/usr/bin/env bash
set -euo pipefail

# WordPress versions to generate docs for (newest first).
# Each entry is a git tag from https://github.com/WordPress/WordPress.
VERSIONS=(
  "6.8.1"
  "6.7.2"
  "6.6.2"
  "6.5.5"
  "6.4.5"
)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
WPDOCS="$SCRIPT_DIR/wpdocs"
OUTPUT="$SCRIPT_DIR/docs"
CACHE_DIR="$SCRIPT_DIR/.wp-cache"
GUIDES="$SCRIPT_DIR/content/guides"
OVERRIDES="$SCRIPT_DIR/content/overrides"

# ── Build wpdocs binary ─────────────────────────────────────────────────────
echo "==> Building wpdocs..."
(cd "$SCRIPT_DIR" && go build -o "$WPDOCS" ./cmd/wpdocs)
echo "    Built $WPDOCS"

# ── Create cache directory for WordPress source trees ────────────────────────
mkdir -p "$CACHE_DIR"

# ── Clone or update each WordPress version ───────────────────────────────────
clone_version() {
  local tag="$1"
  local dest="$CACHE_DIR/wordpress-$tag"

  if [[ -d "$dest/wp-includes" ]]; then
    echo "    Cached: $dest"
    return
  fi

  echo "    Cloning WordPress $tag..."
  rm -rf "$dest"
  git clone --depth 1 --branch "$tag" \
    https://github.com/WordPress/WordPress.git "$dest" 2>&1 | tail -1
}

# ── Generate docs for each version ──────────────────────────────────────────
for tag in "${VERSIONS[@]}"; do
  echo ""
  echo "==> WordPress $tag"

  clone_version "$tag"

  src="$CACHE_DIR/wordpress-$tag"

  echo "    Generating docs..."
  "$WPDOCS" \
    --source "$src" \
    --tag "$tag" \
    --output "$OUTPUT" \
    --guides "$GUIDES" \
    --overrides "$OVERRIDES" \
    --workers 8
done

# ── Build the Hugo site ─────────────────────────────────────────────────────
echo ""
echo "==> Building Hugo site..."
if command -v hugo &>/dev/null; then
  hugo --source "$OUTPUT" --minify
  echo ""
  echo "Done! Static site is in $OUTPUT/public/"
  echo "Run 'hugo server --source $OUTPUT' to preview."
else
  echo "Hugo not found. Install Hugo and run:"
  echo "  hugo --source $OUTPUT --minify"
fi
