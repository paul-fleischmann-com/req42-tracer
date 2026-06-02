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
IMG="$REPO_ROOT/docs/arc42/images"
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

mkdir -p "$IMG"

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
        # Reference pre-rendered SVG from images/ dir (asciidoctor-pdf default imagesdir)
        with open(adoc_path, "w") as f:
            f.write(f".{title}\nimage::{view}.svg[{title}]\n")
    else:
        with open(adoc_path, "w") as f:
            f.write(f".{title}\n[mermaid]\n....\n{source}\n....\n")

    with open(md_path, "w") as f:
        f.write(f"### {title}\n\n```mermaid\n{source}\n```\n")

    print(f"  ✓ {view}  ({title})")
    count += 1

print(f"\nDone — {count} views written ({include_format} format)")
PYEOF

# Generate .adoc includes for any .mmd files not owned by bausteinsicht
python3 - "$OUT" "$INCLUDE_FORMAT" << 'PYEOF'
import os, sys, re

out_dir        = sys.argv[1]
include_format = sys.argv[2]

# Titles for static diagrams (filename → human title)
STATIC_TITLES = {
    "runtime_trace":       "Trace Command Sequence",
    "runtime_analyzegaps": "AnalyzeGaps() Algorithm",
}

for view in sorted(STATIC_TITLES):
    adoc_path = os.path.join(out_dir, f"{view}.adoc")
    mmd_path  = os.path.join(out_dir, f"{view}.mmd")
    if not os.path.exists(mmd_path):
        continue
    with open(mmd_path) as f:
        source = f.read().rstrip("\n")
    title = STATIC_TITLES.get(view, view)
    if include_format == "svg":
        with open(adoc_path, "w") as f:
            f.write(f".{title}\nimage::{view}.svg[{title}]\n")
    else:
        with open(adoc_path, "w") as f:
            f.write(f".{title}\n[mermaid]\n....\n{source}\n....\n")
    print(f"  ✓ {view}  ({title})  [static]")
PYEOF

# Render SVGs when --include svg is requested
if [[ "$INCLUDE_FORMAT" == "svg" ]]; then
  echo ""
  echo "Rendering SVGs with mmdc ..."

  # Puppeteer needs --no-sandbox on GitHub Actions (Ubuntu AppArmor restrictions)
  PUPPETEER_CFG="$(mktemp --suffix=.json)"
  echo '{"args":["--no-sandbox","--disable-setuid-sandbox"]}' > "$PUPPETEER_CFG"
  trap 'rm -f "$TMPJSON" "$PUPPETEER_CFG"' EXIT

  for mmd in "$OUT"/*.mmd; do
    view="$(basename "$mmd" .mmd)"
    svg="$IMG/${view}.svg"
    mmdc -i "$mmd" -o "$svg" --backgroundColor transparent \
         -p "$PUPPETEER_CFG" --quiet
    echo "  ✓ ${view}.svg → images/"
  done
  echo ""
  echo "SVG rendering complete."
fi
