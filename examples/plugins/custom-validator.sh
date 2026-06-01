#!/bin/sh
# Example validator plugin — enforces REQ-* naming convention.
# stdin:  TraceabilityGraph JSON
# stdout: []ValidationIssue JSON

graph=$(cat)
issues="[]"

# Extract requirement IDs and check naming convention
req_ids=$(echo "$graph" | python3 -c "
import json, sys
g = json.load(sys.stdin)
for rid in g.get('requirements', {}).keys():
    print(rid)
" 2>/dev/null)

violations=""
for id in $req_ids; do
    case "$id" in
        REQ-*) ;;  # valid
        *)
            if [ -n "$violations" ]; then violations="$violations,"; fi
            violations="${violations}{\"rule\":\"req-naming\",\"severity\":\"warning\",\"element_id\":\"${id}\",\"message\":\"Requirement ID should start with REQ-\"}"
            ;;
    esac
done

echo "[${violations}]"
