#!/usr/bin/env bash
# generate-diagrams.sh — Export all Bausteinsicht views as Mermaid diagrams.
#
# Views and titles are read directly from architecture.jsonc via the
# bausteinsicht tool — no hardcoded lists needed.
#
# Usage:
#   ./scripts/generate-diagrams.sh                   # .mmd + .adoc (mermaid block) + .md
#   ./scripts/generate-diagrams.sh --include adoc    # (default) AsciiDoc with inline [mermaid] block
#   ./scripts/generate-diagrams.sh --include svg     # AsciiDoc with image::view.svg[] (requires mmdc)
#   ./scripts/generate-diagrams.sh --include md      # Markdown target
#
# The --include svg mode additionally runs mmdc to render SVG files and writes
# adoc includes that reference the SVG images. Use this in CI before a
# docToolchain build to avoid needing mmdc inside the Docker container.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BAUSTEINSICHT="$REPO_ROOT/tools/bausteinsicht/bausteinsicht"
MODEL="$REPO_ROOT/architecture.jsonc"
OUT="$REPO_ROOT/docs/arc42/diagrams"
INCLUDE_FORMAT="adoc"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --include) INCLUDE_FORMAT="$2"; shift 2 ;;
    *) echo "Error: unknown argument '$1'" >&2; exit 1 ;;
  esac
done

if [[ "$INCLUDE_FORMAT" != "adoc" && "$INCLUDE_FORMAT" != "svg" && "$INCLUDE_FORMAT" != "md" ]]; then
  echo "Error: --include must be 'adoc', 'svg', or 'md'" >&2; exit 1
fi

if [[ "$INCLUDE_FORMAT" == "svg" ]] && ! command -v mmdc &>/dev/null; then
  echo "Error: --include svg requires mmdc (npm install -g @mermaid-js/mermaid-cli)" >&2
  exit 1
fi

mkdir -p "$OUT"

echo "Reading views from architecture.jsonc ..."
echo "Include format: $INCLUDE_FORMAT"
echo ""

TMPJSON=$(mktemp)
trap 'rm -f "$TMPJSON"' EXIT

"$BAUSTEINSICHT" --format json --model "$MODEL" \
  export-diagram --diagram-format mermaid > "$TMPJSON"

python3 - "$OUT" "$TMPJSON" "$INCLUDE_FORMAT" << 'PYEOF'
import json, sys, os, re

out_dir        = sys.argv[1]
json_path      = sys.argv[2]
include_format = sys.argv[3]

with open(json_path) as f:
    data = json.load(f)

count = 0
for item in data:
    view   = item["view"]
    source = item["source"].rstrip("\n")

    title_match = re.search(r'^\s+title\s+(.+)$', source, re.MULTILINE)
    title = title_match.group(1).strip() if title_match else view

    mmd_path  = os.path.join(out_dir, f"{view}.mmd")
    adoc_path = os.path.join(out_dir, f"{view}.adoc")
    md_path   = os.path.join(out_dir, f"{view}.md")

    with open(mmd_path, "w") as f:
        f.write(source + "\n")

    if include_format == "svg":
        # Reference pre-rendered SVG — no mmdc needed at asciidoctor render time
        with open(adoc_path, "w") as f:
            f.write(f".{title}\nimage::{view}.svg[{title},opts=inline]\n")
    else:
        with open(adoc_path, "w") as f:
            f.write(f".{title}\n[mermaid]\n....\n{source}\n....\n")

    with open(md_path, "w") as f:
        f.write(f"### {title}\n\n```mermaid\n{source}\n```\n")

    print(f"  ✓ {view}  ({title})")
    count += 1

print(f"\nDone — {count} views written ({include_format} format)")
PYEOF

# Render SVGs when --include svg is requested
if [[ "$INCLUDE_FORMAT" == "svg" ]]; then
  echo ""
  echo "Rendering SVGs with mmdc ..."
  for mmd in "$OUT"/*.mmd; do
    view="$(basename "$mmd" .mmd)"
    svg="$OUT/${view}.svg"
    mmdc -i "$mmd" -o "$svg" --backgroundColor transparent --quiet
    echo "  ✓ ${view}.svg"
  done
  echo ""
  echo "SVG rendering complete."
fi
