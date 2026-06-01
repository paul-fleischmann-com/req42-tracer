#!/usr/bin/env python3
"""
Example linker plugin — links test results to requirements by keyword match.

Any test whose name contains a requirement ID is linked to that requirement.

stdin:  TraceabilityGraph JSON
stdout: []PluginTraceLink JSON
"""

import json, sys

def main():
    g = json.load(sys.stdin)
    req_ids = list(g.get("requirements", {}).keys())
    test_results = g.get("test_results", {})
    links = []

    for result_id, result in test_results.items():
        name = (result.get("test_name") or result_id).lower()
        for req_id in req_ids:
            if req_id.lower().replace("-", "").replace("_", "") in name.replace("-", "").replace("_", ""):
                links.append({
                    "from_id":   result_id,
                    "from_type": "test-result",
                    "to_id":     req_id,
                    "to_type":   "requirement",
                    "link_type": "verifies",
                    "reason":    f"keyword match: {req_id} in {result.get('test_name', result_id)}"
                })
                break  # one link per result

    print(json.dumps(links))

if __name__ == "__main__":
    main()
