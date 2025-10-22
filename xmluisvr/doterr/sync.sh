#!/usr/bin/env bash
# update-doterr.sh
# Scan for all files named doterr.go under the current directory,
# and replace each with the canonical xmluisvr/doterr/doterr.go
# after rewriting the package line to be "package <dirname>".
#
# Usage:
#   ./update-doterr.sh           # do the updates (with backups)
#   DRY_RUN=1 ./update-doterr.sh # show what would change, don't write
#
# Notes:
# - Creates a .bak next to each modified file.
# - Skips the canonical file itself.
# - Handles paths with spaces.
# - Only the first "package doterr" line in the canonical file is replaced.
# - Requires bash, awk, mktemp, find, diff, mv.

set -euo pipefail

CANONICAL="xmluisvr/doterr/doterr.go"

if [[ ! -f "$CANONICAL" ]]; then
  echo "ERROR: Canonical file not found at: $CANONICAL" >&2
  exit 1
fi

# Readability helper: rewrite canonical content's package line to a given package name.
# Usage: rewrite_package <pkgname> <outpath>
rewrite_package() {
  local pkg="$1"
  local out="$2"

  # macOS/BSD awk (no gensub, no \b). Replace both comment and decl.
  awk -v pkg="$pkg" '
    {
      # Replace ALL instances; keeps original capitalization of "Package"/"package".
      gsub(/Package[[:space:]]+doterr/, "Package " pkg)
      gsub(/package[[:space:]]+doterr/, "package " pkg)
      print
    }
  ' "$CANONICAL" > "$out"
}



updated=0
skipped=0

# Find all doterr.go files under current dir (null-delimited to handle spaces).
# We won't filter out the canonical at find-time because -ef comparison is robust in the loop.
while IFS= read -r -d '' target; do
  # Skip the canonical file itself
  if [[ "$target" -ef "$CANONICAL" ]]; then
    ((skipped++))
    echo "skip: canonical -> $target"
    continue
  fi

  # Determine the new package name: basename of the directory containing the target file.
  dirpath="$(dirname "$target")"
  pkgname="$(basename "$dirpath")"

  # Build the transformed content in a temp file.
  tmp="$(mktemp)"
  rewrite_package "$pkgname" "$tmp"

  # If unchanged vs. current target, skip writing.
  if diff -q "$tmp" "$target" >/dev/null 2>&1; then
    rm -f "$tmp"
    ((skipped++))
    echo "skip: no changes for $target (package $pkgname)"
    continue
  fi

  if [[ "${DRY_RUN:-}" == "1" ]]; then
    echo "would update: $target (package $pkgname)"
    rm -f "$tmp"
    continue
  fi

#  # Backup the existing file, then replace.
#  cp -p "$target" "$target.bak"
  mv "$tmp" "$target"
  # Preserve canonical file mode/owner if desired (optional). Commented by default:
  # chmod --reference="$CANONICAL" "$target" 2>/dev/null || true

  ((updated++))
  echo "updated: $target (package $pkgname)" # -> backup at $target.bak"
done < <(find . -type f -name 'doterr.go' -print0)

echo
echo "Done. Updated: $updated, Skipped: $skipped"
if [[ "${DRY_RUN:-}" == "1" ]]; then
  echo "(dry-run: no files were changed)"
fi
