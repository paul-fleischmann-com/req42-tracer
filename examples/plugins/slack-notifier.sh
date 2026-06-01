#!/bin/sh
# Example notifier plugin — sends gap summary to Slack.
# stdin:  gaps/metrics JSON
# stdout: (nothing required)
# env:    SLACK_WEBHOOK_URL

WEBHOOK="${SLACK_WEBHOOK_URL:-}"
if [ -z "$WEBHOOK" ]; then
    echo "SLACK_WEBHOOK_URL not set" >&2
    exit 1
fi

payload=$(cat)
untraced=$(echo "$payload" | python3 -c "import json,sys; d=json.load(sys.stdin); print(len(d.get('untraced_test_results', [])))" 2>/dev/null || echo "?")
orphans=$(echo "$payload" | python3 -c "import json,sys; d=json.load(sys.stdin); print(len(d.get('orphan_requirements', [])))" 2>/dev/null || echo "?")

message=":warning: *req42-tracer Gap Report*\n• Untraced test results: ${untraced}\n• Orphan requirements: ${orphans}"

curl -s -X POST "$WEBHOOK" \
    -H "Content-Type: application/json" \
    -d "{\"text\": \"${message}\"}" \
    > /dev/null
