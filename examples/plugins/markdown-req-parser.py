#!/usr/bin/env python3
"""
Example requirement-parser plugin for Markdown files.

Protocol:
  stdin:  {"file": "path/to/reqs.md", "project": "software"}
  stdout: [{"id": "REQ-001", "title": "...", "priority": "medium", "status": "draft", "aspice": ""}]

Markdown format expected:
  ## REQ-001: Title of requirement
  - priority: high
  - status: approved
  - aspice: SWE.1
"""

import json, sys, re

def parse(path, project):
    results = []
    try:
        with open(path, encoding="utf-8") as f:
            content = f.read()
    except OSError as e:
        print(f"error reading {path}: {e}", file=sys.stderr)
        sys.exit(1)

    current = None
    for line in content.splitlines():
        m = re.match(r"^#{1,3}\s+(REQ-[\w-]+):\s*(.+)", line)
        if m:
            if current:
                results.append(current)
            current = {"id": m.group(1), "title": m.group(2).strip(),
                       "priority": "medium", "status": "draft", "aspice": "", "project": project}
            continue
        if current:
            for key in ("priority", "status", "aspice"):
                m2 = re.match(rf"^\s*[-*]\s+{key}:\s*(.+)", line, re.I)
                if m2:
                    current[key] = m2.group(1).strip()

    if current:
        results.append(current)
    return results

if __name__ == "__main__":
    req = json.load(sys.stdin)
    print(json.dumps(parse(req["file"], req.get("project", ""))))
