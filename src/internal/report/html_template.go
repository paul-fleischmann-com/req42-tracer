package report

// HTMLTemplate is the embedded HTML template for graph visualization with D3.js.
const HTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>req42-tracer: Traceability Report</title>
    <script src="https://d3js.org/d3.v7.min.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: #f5f5f5;
            color: #333;
        }

        .container {
            display: flex;
            height: 100vh;
        }

        .sidebar {
            width: 250px;
            background: #2c3e50;
            color: white;
            padding: 20px;
            overflow-y: auto;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }

        .sidebar h2 {
            font-size: 18px;
            margin-bottom: 20px;
            text-transform: uppercase;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }

        .sidebar h3 {
            font-size: 14px;
            margin-top: 15px;
            margin-bottom: 10px;
            color: #ecf0f1;
        }

        .legend {
            display: flex;
            flex-direction: column;
            gap: 8px;
            margin-bottom: 20px;
        }

        .legend-item {
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 13px;
        }

        .legend-dot {
            width: 12px;
            height: 12px;
            border-radius: 50%;
        }

        .legend-dot.requirement { background: #4a90e2; }
        .legend-dot.arch { background: #7ed321; }
        .legend-dot.dsn { background: #9b59b6; }
        .legend-dot.test-spec { background: #e74c3c; }
        .legend-dot.test-code { background: #e67e22; }
        .legend-dot.test-result { background: #999; }

        .stats {
            background: rgba(255,255,255,0.1);
            border-radius: 4px;
            padding: 12px;
            font-size: 12px;
            margin-bottom: 20px;
        }

        .stats-item {
            display: flex;
            justify-content: space-between;
            margin: 4px 0;
        }

        .controls {
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .control-group {
            display: flex;
            flex-direction: column;
            gap: 6px;
        }

        .control-group label {
            font-size: 12px;
            text-transform: uppercase;
            color: #bdc3c7;
        }

        button {
            background: #3498db;
            color: white;
            border: none;
            padding: 8px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 12px;
            text-transform: uppercase;
            transition: background 0.3s;
        }

        button:hover {
            background: #2980b9;
        }

        button.active {
            background: #e74c3c;
        }

        .main {
            flex: 1;
            display: flex;
            flex-direction: column;
        }

        .header {
            background: white;
            padding: 20px;
            border-bottom: 1px solid #e0e0e0;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
        }

        .header h1 {
            font-size: 24px;
            margin-bottom: 5px;
        }

        .header p {
            color: #666;
            font-size: 14px;
        }

        .tabs {
            display: flex;
            gap: 0;
            border-bottom: 2px solid #e0e0e0;
            margin-top: 15px;
        }

        .tab-button {
            background: transparent;
            border: none;
            padding: 12px 24px;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            color: #666;
            border-bottom: 3px solid transparent;
            transition: all 0.3s;
        }

        .tab-button:hover {
            color: #333;
        }

        .tab-button.active {
            color: #3498db;
            border-bottom-color: #3498db;
        }

        .tab-content {
            display: none;
            flex: 1;
            background: #fff;
            position: relative;
            overflow: auto;
        }

        .tab-content.active {
            display: flex;
        }

        #graph {
            flex: 1;
            background: #fff;
            position: relative;
        }

        #matrix {
            flex: 1;
            flex-direction: column;
        }

        .matrix-controls {
            padding: 20px;
            background: #f8f9fa;
            border-bottom: 1px solid #e0e0e0;
            display: flex;
            gap: 15px;
            flex-wrap: wrap;
            align-items: center;
        }

        .matrix-controls input,
        .matrix-controls select {
            padding: 8px 12px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 13px;
        }

        .matrix-table-wrapper {
            flex: 1;
            overflow: auto;
            padding: 20px;
        }

        .matrix-table {
            border-collapse: collapse;
            font-size: 12px;
            background: white;
        }

        .matrix-table th,
        .matrix-table td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: center;
        }

        .matrix-table th {
            background: #f0f0f0;
            font-weight: 600;
            position: sticky;
            top: 0;
        }

        .matrix-table td.req-id {
            text-align: left;
            background: #f8f9fa;
            font-weight: 500;
            min-width: 150px;
        }

        .matrix-cell {
            cursor: pointer;
            transition: background 0.2s;
        }

        .matrix-cell.covered {
            background: #d4edda;
            color: #155724;
        }

        .matrix-cell.missing {
            background: #f8d7da;
            color: #721c24;
        }

        .matrix-cell.stale {
            background: #fff3cd;
            color: #856404;
        }

        .impl-cell code {
            font-size: 11px;
            background: #f0f0f0;
            padding: 1px 4px;
            border-radius: 3px;
            white-space: nowrap;
        }

        .missing-impl {
            color: #aaa;
        }

        .coverage-badge {
            font-size: 16px;
            text-align: center;
        }

        .matrix-stats {
            padding: 20px;
            background: #f8f9fa;
            border-top: 1px solid #e0e0e0;
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
        }

        .stat-item {
            background: white;
            padding: 12px;
            border-radius: 4px;
            border-left: 4px solid #3498db;
        }

        .stat-label {
            font-size: 12px;
            color: #666;
            text-transform: uppercase;
        }

        .stat-value {
            font-size: 18px;
            font-weight: bold;
            margin-top: 4px;
        }

        svg {
            width: 100%;
            height: 100%;
        }

        .node {
            stroke: #fff;
            stroke-width: 2px;
            cursor: pointer;
        }

        .node:hover {
            stroke: #000;
            stroke-width: 3px;
        }

        .node.selected {
            stroke: #e74c3c;
            stroke-width: 3px;
            filter: drop-shadow(0 0 6px rgba(231, 76, 60, 0.8));
        }

        .node-label {
            font-size: 11px;
            pointer-events: none;
            text-anchor: middle;
            dominant-baseline: central;
            font-weight: 500;
        }

        .link {
            stroke: #bbb;
            stroke-width: 1.5px;
            stroke-dasharray: 0;
            opacity: 0.6;
        }

        .link.faded {
            opacity: 0.1;
        }

        /* BFS focus: dimmed = not in neighborhood */
        .node.dimmed {
            opacity: 0.06;
            pointer-events: none;
            transition: opacity 0.3s;
        }
        .node-label.dimmed {
            opacity: 0.06;
            transition: opacity 0.3s;
        }
        .link.dimmed {
            opacity: 0.04;
            transition: opacity 0.3s;
        }
        .node { transition: opacity 0.3s; }
        .link { transition: opacity 0.3s; }

        /* Sidebar element selector */
        .selector-section {
            border-top: 1px solid #e0e0e0;
            padding: 10px 0 0;
        }
        .selector-section h3 {
            font-size: 11px;
            color: #999;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin: 0 0 6px;
            padding: 0 16px;
        }
        .selector-search {
            display: block;
            width: calc(100% - 32px);
            margin: 0 16px 6px;
            padding: 5px 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 12px;
            box-sizing: border-box;
        }
        .selector-list {
            max-height: 220px;
            overflow-y: auto;
            padding: 0 8px;
        }
        .selector-item {
            display: flex;
            align-items: center;
            gap: 7px;
            padding: 4px 8px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 12px;
            font-family: monospace;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
        .selector-item:hover { background: #f0f0f0; }
        .selector-item.active { background: #fff3cd; font-weight: 600; }
        .selector-dot {
            width: 8px; height: 8px;
            border-radius: 50%;
            flex-shrink: 0;
        }
        .selector-show-all {
            display: block;
            width: calc(100% - 32px);
            margin: 6px 16px 0;
            padding: 4px 8px;
            font-size: 12px;
            border: 1px solid #ddd;
            border-radius: 4px;
            background: #f8f9fa;
            cursor: pointer;
            text-align: center;
        }
        .selector-show-all:hover { background: #e9ecef; }

        /* Hop controls bar in graph view */
        .hop-controls {
            position: absolute;
            top: 20px;
            left: 20px;
            display: flex;
            align-items: center;
            gap: 6px;
            background: white;
            border: 1px solid #ddd;
            border-radius: 4px;
            padding: 5px 10px;
            font-size: 12px;
            z-index: 100;
            box-shadow: 0 1px 3px rgba(0,0,0,0.08);
        }
        .hop-controls label { color: #666; margin-right: 2px; }
        .hop-btn {
            width: 24px; height: 24px;
            border: 1px solid #ddd;
            border-radius: 3px;
            background: #f8f9fa;
            cursor: pointer;
            font-size: 15px;
            font-weight: bold;
            line-height: 1;
            padding: 0;
            display: flex; align-items: center; justify-content: center;
        }
        .hop-btn:hover { background: #e9ecef; }
        .hop-value {
            min-width: 18px;
            text-align: center;
            font-weight: 700;
            color: #333;
        }
        .hop-full-btn {
            padding: 3px 7px;
            border: 1px solid #ddd;
            border-radius: 3px;
            background: #f8f9fa;
            cursor: pointer;
            font-size: 11px;
            margin-left: 4px;
        }
        .hop-full-btn:hover { background: #e9ecef; }
        .hop-controls.inactive { opacity: 0.4; pointer-events: none; }

        .link-label {
            font-size: 10px;
            pointer-events: none;
            fill: #666;
            background: white;
        }

        .tooltip {
            position: absolute;
            background: rgba(0, 0, 0, 0.9);
            color: white;
            padding: 12px;
            border-radius: 4px;
            font-size: 12px;
            pointer-events: none;
            z-index: 1000;
            max-width: 300px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.3);
            display: none;
        }

        .tooltip.visible {
            display: block;
        }

        .tooltip-title {
            font-weight: bold;
            margin-bottom: 6px;
            word-break: break-word;
        }

        .tooltip-content {
            font-size: 11px;
            line-height: 1.5;
        }

        .tooltip-meta {
            margin-top: 6px;
            padding-top: 6px;
            border-top: 1px solid rgba(255,255,255,0.2);
            font-style: italic;
        }

        .zoom-controls {
            position: absolute;
            top: 20px;
            right: 20px;
            display: flex;
            gap: 8px;
            z-index: 100;
        }

        .zoom-btn {
            background: white;
            color: #333;
            border: 1px solid #ddd;
            width: 36px;
            height: 36px;
            padding: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            border-radius: 4px;
            cursor: pointer;
            font-size: 18px;
            font-weight: bold;
            transition: all 0.2s;
        }

        .zoom-btn:hover {
            background: #f0f0f0;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        @media (max-width: 1024px) {
            .sidebar {
                width: 200px;
            }
        }

        @media (max-width: 768px) {
            .container {
                flex-direction: column;
            }

            .sidebar {
                width: 100%;
                max-height: 200px;
            }
        }

        /* ASPICE Dashboard */
        #aspice {
            flex: 1;
            flex-direction: column;
            overflow-y: auto;
            background: #f8f9fa;
        }

        .aspice-overview {
            background: white;
            border-bottom: 1px solid #e0e0e0;
            padding: 24px 30px;
        }

        .aspice-overview h2 {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 16px;
            color: #333;
        }

        .overall-coverage {
            display: flex;
            align-items: center;
            gap: 24px;
        }

        .overall-pct {
            font-size: 42px;
            font-weight: bold;
            min-width: 90px;
        }

        .overall-pct.good { color: #27ae60; }
        .overall-pct.warning { color: #f39c12; }
        .overall-pct.danger { color: #e74c3c; }

        .overall-bar-wrap { flex: 1; }

        .overall-bar-label {
            font-size: 12px;
            color: #888;
            text-transform: uppercase;
            margin-bottom: 6px;
        }

        .overall-bar-track {
            height: 10px;
            background: #e0e0e0;
            border-radius: 5px;
            overflow: hidden;
        }

        .overall-bar-fill {
            height: 100%;
            border-radius: 5px;
            transition: width 0.6s ease;
        }

        .overall-bar-fill.good { background: #27ae60; }
        .overall-bar-fill.warning { background: #f39c12; }
        .overall-bar-fill.danger { background: #e74c3c; }

        .aspice-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
            gap: 16px;
            padding: 20px;
        }

        .process-card {
            background: white;
            border-radius: 8px;
            border: 1px solid #e0e0e0;
            overflow: hidden;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
        }

        .process-card-header {
            padding: 14px 16px;
            display: flex;
            align-items: center;
            gap: 12px;
            border-left: 4px solid #3498db;
        }

        .process-card-header.good { border-left-color: #27ae60; }
        .process-card-header.warning { border-left-color: #f39c12; }
        .process-card-header.danger { border-left-color: #e74c3c; }

        .process-id-badge {
            background: #2c3e50;
            color: white;
            padding: 3px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: bold;
            white-space: nowrap;
        }

        .process-card-name {
            flex: 1;
            font-weight: 600;
            font-size: 13px;
            color: #333;
        }

        .process-card-pct {
            font-size: 16px;
            font-weight: bold;
        }

        .process-card-pct.good { color: #27ae60; }
        .process-card-pct.warning { color: #f39c12; }
        .process-card-pct.danger { color: #e74c3c; }

        .process-progress {
            height: 4px;
            background: #f0f0f0;
        }

        .process-progress-fill {
            height: 100%;
            transition: width 0.5s ease;
        }

        .process-progress-fill.good { background: #27ae60; }
        .process-progress-fill.warning { background: #f39c12; }
        .process-progress-fill.danger { background: #e74c3c; }

        .bp-list { padding: 4px 0; }

        .bp-item {
            padding: 8px 16px;
            border-bottom: 1px solid #f5f5f5;
        }

        .bp-item:last-child { border-bottom: none; }

        .bp-row {
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .bp-id {
            font-size: 11px;
            font-weight: bold;
            color: #888;
            min-width: 72px;
        }

        .bp-title {
            flex: 1;
            font-size: 12px;
            color: #444;
        }

        .bp-badge {
            font-size: 11px;
            font-weight: bold;
            padding: 2px 6px;
            border-radius: 3px;
        }

        .bp-badge.good { background: #d5f5e3; color: #1e8449; }
        .bp-badge.warning { background: #fef9e7; color: #b7950b; }
        .bp-badge.danger { background: #fdedec; color: #c0392b; }

        .bp-gaps {
            margin: 4px 0 0 80px;
            padding: 0;
            list-style: none;
        }

        .bp-gaps li {
            font-size: 11px;
            color: #e74c3c;
            padding: 1px 0;
        }

        .bp-gaps li::before {
            content: "→ ";
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="sidebar">
            <h2>Traceability Graph</h2>

            <div class="legend">
                <h3>Node Types</h3>
                <div class="legend-item">
                    <div class="legend-dot requirement"></div>
                    <span>Requirement</span>
                </div>
                <div class="legend-item">
                    <div class="legend-dot arch"></div>
                    <span>Architecture</span>
                </div>
                <div class="legend-item">
                    <div class="legend-dot dsn"></div>
                    <span>Design Element (SWE.3)</span>
                </div>
                <div class="legend-item">
                    <div class="legend-dot test-spec"></div>
                    <span>Test Spec</span>
                </div>
                <div class="legend-item">
                    <div class="legend-dot test-code"></div>
                    <span>Test Code</span>
                </div>
                <div class="legend-item">
                    <div class="legend-dot test-result"></div>
                    <span>Test Result</span>
                </div>
            </div>

            <div class="legend">
                <h3>Link Types</h3>
                <div class="legend-item">→ satisfies</div>
                <div class="legend-item">→ implements</div>
                <div class="legend-item">→ verifies</div>
                <div class="legend-item">→ derives</div>
                <div class="legend-item">→ covers</div>
            </div>

            <div class="stats" id="stats">
                <div class="stats-item">
                    <span>Requirements:</span>
                    <span id="stat-reqs">0</span>
                </div>
                <div class="stats-item">
                    <span>Architecture:</span>
                    <span id="stat-arch">0</span>
                </div>
                <div class="stats-item">
                    <span>Test Specs:</span>
                    <span id="stat-specs">0</span>
                </div>
                <div class="stats-item">
                    <span>Design Elements:</span>
                    <span id="stat-dsn">0</span>
                </div>
                <div class="stats-item">
                    <span>Test Results:</span>
                    <span id="stat-results">0</span>
                </div>
                <div class="stats-item">
                    <span>Total Nodes:</span>
                    <span id="stat-total">0</span>
                </div>
                <div class="stats-item">
                    <span>Total Links:</span>
                    <span id="stat-links">0</span>
                </div>
            </div>

            <div class="selector-section">
                <h3>Focus Element</h3>
                <input id="selector-search" class="selector-search" type="text" placeholder="Filter by ID…">
                <div id="selector-list" class="selector-list"></div>
                <button id="selector-show-all" class="selector-show-all">Show All</button>
            </div>

            <div class="controls">
                <div class="control-group">
                    <label>Filters</label>
                    <button id="btn-reset">Reset View</button>
                </div>
                <div class="control-group">
                    <label>Node Types</label>
                    <button id="btn-filter-all" class="active">All</button>
                    <button id="btn-filter-req">Requirements</button>
                    <button id="btn-filter-arch">Architecture</button>
                    <button id="btn-filter-dsn">Design</button>
                    <button id="btn-filter-tests">Tests</button>
                </div>
                <div class="control-group">
                    <label>Layout</label>
                    <button id="btn-layout-force" class="active">Force</button>
                    <button id="btn-layout-radial">Radial</button>
                </div>
            </div>
        </div>

        <div class="main">
            <div class="header">
                <div>
                    <h1>Traceability Report</h1>
                    <p>Interactive dependency graph and traceability matrix</p>
                </div>
                <div class="tabs">
                    <button class="tab-button active" onclick="switchTab('graph')">Graph View</button>
                    <button class="tab-button" onclick="switchTab('matrix')">Matrix View</button>
                    <button class="tab-button" onclick="switchTab('aspice')">ASPICE Dashboard</button>
                    <button class="tab-button" onclick="switchTab('gaps')" id="btn-tab-gaps">Gaps</button>
                    <button class="tab-button" onclick="switchTab('elements')">Elements</button>
                    <button class="tab-button" onclick="switchTab('coverage')">Coverage</button>
                </div>
            </div>

            <div id="graph" class="tab-content active">
                <div class="zoom-controls">
                    <button class="zoom-btn" id="btn-zoom-in">+</button>
                    <button class="zoom-btn" id="btn-zoom-out">−</button>
                    <button class="zoom-btn" id="btn-zoom-fit">⊙</button>
                </div>
                <div class="hop-controls inactive" id="hop-controls">
                    <label>Hops:</label>
                    <button class="hop-btn" id="hop-minus">−</button>
                    <span class="hop-value" id="hop-value">2</span>
                    <button class="hop-btn" id="hop-plus">+</button>
                    <button class="hop-full-btn" id="hop-full" title="Show full chain (10 hops)">Full chain</button>
                </div>
                <div class="tooltip" id="tooltip"></div>
            </div>

            <div id="matrix" class="tab-content">
                <div class="matrix-controls">
                    <input type="text" id="search-req" placeholder="Search requirements...">
                    <select id="filter-priority">
                        <option value="">All Priorities</option>
                        <option value="high">High</option>
                        <option value="medium">Medium</option>
                        <option value="low">Low</option>
                    </select>
                    <select id="filter-status">
                        <option value="">All Status</option>
                        <option value="approved">Approved</option>
                        <option value="draft">Draft</option>
                        <option value="deprecated">Deprecated</option>
                    </select>
                    <button onclick="exportMatrixCSV()" style="margin-left: auto;">📥 Export CSV</button>
                </div>
                <div class="matrix-table-wrapper" id="matrix-table-container">
                    <!-- Matrix table will be generated by JavaScript -->
                </div>
                <div class="matrix-stats">
                    <div class="stat-item">
                        <div class="stat-label">Total Requirements</div>
                        <div class="stat-value" id="matrix-total">0</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Covered</div>
                        <div class="stat-value" id="matrix-covered">0</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Missing</div>
                        <div class="stat-value" id="matrix-missing">0</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Coverage</div>
                        <div class="stat-value" id="matrix-coverage">0%</div>
                    </div>
                </div>
            </div>

            <div id="aspice" class="tab-content">
                <div class="aspice-overview">
                    <h2>ASPICE PAM 4.0 Compliance</h2>
                    <div class="overall-coverage">
                        <div class="overall-pct" id="aspice-overall-pct">-</div>
                        <div class="overall-bar-wrap">
                            <div class="overall-bar-label">Overall Coverage</div>
                            <div class="overall-bar-track">
                                <div class="overall-bar-fill" id="aspice-overall-fill" style="width:0%"></div>
                            </div>
                        </div>
                    </div>
                    <div style="margin-top:12px;display:flex;align-items:center;gap:8px;">
                        <label for="filter-process" style="font-size:13px;color:#666;">Filter process:</label>
                        <select id="filter-process" style="padding:4px 8px;border:1px solid #ddd;border-radius:4px;font-size:13px;" onchange="renderASPICE()">
                            <option value="">All processes</option>
                        </select>
                    </div>
                </div>
                <div class="aspice-grid" id="aspice-grid"></div>
            </div>

            <div id="gaps" class="tab-content">
                <div style="padding:20px">
                    <h2 style="margin-bottom:16px">Gap Analysis</h2>
                    <div id="gaps-content"></div>
                </div>
            </div>

            <div id="elements" class="tab-content" style="flex-direction:column">
                <div style="padding:16px 20px;background:#f8f9fa;border-bottom:1px solid #e0e0e0;display:flex;gap:12px;flex-wrap:wrap;align-items:center">
                    <strong>Elements</strong>
                    <select id="el-filter-type" style="padding:6px 10px;border:1px solid #ddd;border-radius:4px;font-size:13px">
                        <option value="">All Types</option>
                        <option value="req">Requirement</option>
                        <option value="arch">Architecture</option>
                        <option value="dsn">Design Element</option>
                        <option value="test-spec">Test Spec</option>
                        <option value="test-result">Test Result</option>
                    </select>
                    <input id="el-search" type="text" placeholder="Search ID or title…"
                        style="padding:6px 10px;border:1px solid #ddd;border-radius:4px;font-size:13px;flex:1;min-width:150px">
                    <span id="el-count" style="font-size:13px;color:#666"></span>
                </div>
                <div style="flex:1;overflow:auto;padding:0 20px 20px">
                    <table id="el-table" style="width:100%;border-collapse:collapse;font-size:13px;margin-top:12px">
                        <thead>
                            <tr style="background:#f0f0f0;position:sticky;top:0">
                                <th style="padding:8px;text-align:left;cursor:pointer;border:1px solid #ddd" onclick="elSort('id')">ID ↕</th>
                                <th style="padding:8px;text-align:left;border:1px solid #ddd">Type</th>
                                <th style="padding:8px;text-align:left;cursor:pointer;border:1px solid #ddd" onclick="elSort('title')">Title ↕</th>
                                <th style="padding:8px;text-align:left;border:1px solid #ddd">Impl / Status</th>
                                <th style="padding:8px;text-align:center;border:1px solid #ddd" title="Elements linking TO this">Trace ↑</th>
                                <th style="padding:8px;text-align:center;border:1px solid #ddd" title="Elements this links TO">Trace ↓</th>
                            </tr>
                        </thead>
                        <tbody id="el-tbody"></tbody>
                    </table>
                </div>
            </div>
            <div id="coverage" class="tab-content" style="flex-direction:column">
                <div style="padding:16px 20px;background:#f8f9fa;border-bottom:1px solid #e0e0e0;display:flex;gap:12px;flex-wrap:wrap;align-items:center">
                    <strong>Coverage Dashboard</strong>
                    <select id="cov-filter-level" style="padding:6px 10px;border:1px solid #ddd;border-radius:4px;font-size:13px">
                        <option value="">All levels</option>
                        <option value="danger">Danger (&lt;70%)</option>
                        <option value="warning">Warning (70–80%)</option>
                        <option value="good">Good (≥80%)</option>
                    </select>
                    <input id="cov-search" type="text" placeholder="Search package or arch…"
                        style="padding:6px 10px;border:1px solid #ddd;border-radius:4px;font-size:13px;flex:1;min-width:150px">
                    <span id="cov-count" style="font-size:13px;color:#666"></span>
                </div>
                <div style="padding:12px 20px;display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:12px" id="cov-cards"></div>
                <div style="flex:1;overflow:auto;padding:0 20px 20px">
                    <table style="width:100%;border-collapse:collapse;font-size:13px">
                        <thead>
                            <tr style="background:#f0f0f0;position:sticky;top:0">
                                <th style="padding:8px;text-align:left;border:1px solid #ddd;cursor:pointer" onclick="covSort('arch_id')">Arch ↕</th>
                                <th style="padding:8px;text-align:left;border:1px solid #ddd;cursor:pointer;font-family:monospace" onclick="covSort('package')">Package ↕</th>
                                <th style="padding:8px;text-align:right;border:1px solid #ddd;cursor:pointer" onclick="covSort('statements')">Stmts ↕</th>
                                <th style="padding:8px;text-align:right;border:1px solid #ddd;cursor:pointer" onclick="covSort('covered')">Covered ↕</th>
                                <th style="padding:8px;border:1px solid #ddd;cursor:pointer" onclick="covSort('pct')">Coverage ↕</th>
                            </tr>
                        </thead>
                        <tbody id="cov-tbody"></tbody>
                        <tfoot id="cov-tfoot"></tfoot>
                    </table>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Graph data will be injected here: <!--GRAPH_DATA-->
        const graphData = <!--GRAPH_DATA_JSON-->;
        const matrixData = <!--MATRIX_DATA_JSON-->;
        const aspiceData = <!--ASPICE_DATA_JSON-->;
        const gapsData = <!--GAPS_DATA_JSON-->;
        const elementsData = <!--ELEMENTS_DATA_JSON-->;
        const coverageData = <!--COVERAGE_DATA_JSON-->;
        const SOURCE_BASE_URL = "<!--SOURCE_BASE_URL-->";

        // Returns an HTML string for an impl= value: clickable link when SOURCE_BASE_URL is set.
        function renderImplCell(implVal) {
            if (!implVal) return '<span class="missing-impl">—</span>';
            const parts = implVal.split(':');
            const isFile = parts[0].includes('.');
            if (isFile && SOURCE_BASE_URL) {
                const file = parts[0];
                const line = parts[2] ? parseInt(parts[2], 10) : 0;
                const href = SOURCE_BASE_URL.replace(/\/$/, '') + '/' + file + (line ? '#L' + line : '');
                return '<a href="' + escHtml(href) + '" target="_blank" rel="noopener"><code>' + escHtml(implVal) + '</code></a>';
            }
            return '<code>' + escHtml(implVal) + '</code>';
        }

        // Global variables for filtering
        let globalNode = null;
        let globalLink = null;
        let globalLabel = null;
        let globalEdgeLabel = null;
        let globalGraphData = null;

        // Focus state
        let focusNodeId = null;
        let focusHops = 2;
        const HOP_MIN = 1, HOP_MAX = 10;

        // BFS: returns Set of node IDs reachable within N steps (bidirectional)
        function bfsNeighbors(startId, hops, data) {
            const visited = new Set([startId]);
            let frontier = [startId];
            for (let h = 0; h < hops; h++) {
                const next = [];
                frontier.forEach(id => {
                    data.edges.forEach(e => {
                        const sid = typeof e.source === 'object' ? e.source.id : e.source;
                        const tid = typeof e.target === 'object' ? e.target.id : e.target;
                        if (sid === id && !visited.has(tid)) { visited.add(tid); next.push(tid); }
                        if (tid === id && !visited.has(sid)) { visited.add(sid); next.push(sid); }
                    });
                });
                frontier = next;
            }
            return visited;
        }

        function applyFocus(nodeId, hops, data) {
            if (!globalNode) return;
            const d = data || globalGraphData;
            if (!d) return;
            focusNodeId = nodeId;
            focusHops = Math.max(HOP_MIN, Math.min(HOP_MAX, hops));

            // Update hop display
            document.getElementById('hop-value').textContent = focusHops;
            document.getElementById('hop-controls').classList.remove('inactive');

            const visible = bfsNeighbors(nodeId, focusHops, d);

            globalNode
                .classed('selected', n => n.id === nodeId)
                .classed('faded', false)
                .classed('dimmed', n => !visible.has(n.id));
            globalLabel.classed('dimmed', n => !visible.has(n.id));
            globalLink.classed('faded', false)
                .classed('dimmed', e => {
                    const sid = typeof e.source === 'object' ? e.source.id : e.source;
                    const tid = typeof e.target === 'object' ? e.target.id : e.target;
                    return !visible.has(sid) || !visible.has(tid);
                });

            // Highlight sidebar item
            document.querySelectorAll('.selector-item').forEach(el => {
                el.classList.toggle('active', el.dataset.id === nodeId);
            });

            // URL hash
            window.location.hash = 'focus=' + encodeURIComponent(nodeId) + '&hops=' + focusHops;

            // Switch to graph tab if not already there
            if (!document.getElementById('graph').classList.contains('active')) {
                switchTab('graph');
            }
        }

        function clearFocus() {
            focusNodeId = null;
            if (globalNode) {
                globalNode.classed('selected', false).classed('faded', false).classed('dimmed', false);
            }
            if (globalLabel) globalLabel.classed('dimmed', false);
            if (globalLink) globalLink.classed('faded', false).classed('dimmed', false);
            document.querySelectorAll('.selector-item').forEach(el => el.classList.remove('active'));
            document.getElementById('hop-controls').classList.add('inactive');
            window.location.hash = '';
        }

        function renderSidebarSelector(data) {
            const colorMap = {
                'requirement': '#4a90e2', 'arch': '#7ed321', 'dsn': '#9b59b6',
                'test-spec': '#e74c3c', 'test-code': '#e67e22', 'test-result': '#999'
            };
            const list = document.getElementById('selector-list');
            if (!list || !data || !data.nodes) return;

            function rebuild(filter) {
                const q = filter.toLowerCase();
                const items = data.nodes
                    .filter(n => !q || n.id.toLowerCase().includes(q) || (n.label||'').toLowerCase().includes(q))
                    .sort((a,b) => a.id.localeCompare(b.id));
                list.innerHTML = items.map(n =>
                    '<div class="selector-item' + (n.id === focusNodeId ? ' active' : '') + '" data-id="' + escHtml(n.id) + '">' +
                    '<div class="selector-dot" style="background:' + (colorMap[n.type]||'#999') + '"></div>' +
                    escHtml(n.id) + '</div>'
                ).join('');
                list.querySelectorAll('.selector-item').forEach(el => {
                    el.addEventListener('click', () => applyFocus(el.dataset.id, focusHops, data));
                });
            }

            rebuild('');
            document.getElementById('selector-search').addEventListener('input', function() {
                rebuild(this.value);
            });
            document.getElementById('selector-show-all').addEventListener('click', clearFocus);
        }

        function initHopControls(data) {
            document.getElementById('hop-minus').addEventListener('click', () => {
                if (focusNodeId && focusHops > HOP_MIN) applyFocus(focusNodeId, focusHops - 1, data);
            });
            document.getElementById('hop-plus').addEventListener('click', () => {
                if (focusNodeId && focusHops < HOP_MAX) applyFocus(focusNodeId, focusHops + 1, data);
            });
            document.getElementById('hop-full').addEventListener('click', () => {
                if (focusNodeId) applyFocus(focusNodeId, HOP_MAX, data);
            });
        }

        function restoreHashState(data) {
            const hash = window.location.hash.slice(1);
            if (!hash) return;
            const params = Object.fromEntries(hash.split('&').map(p => p.split('=')));
            const id = params.focus ? decodeURIComponent(params.focus) : null;
            const hops = parseInt(params.hops) || 2;
            if (id && data.nodes.some(n => n.id === id)) {
                applyFocus(id, hops, data);
            }
        }

        function initializeGraph(data) {
            globalGraphData = data;
            // Update statistics
            updateStats(data);

            // Set up SVG
            const container = document.getElementById('graph');
            const width = container.clientWidth;
            const height = container.clientHeight;

            const svg = d3.select('#graph').append('svg');

            // Create groups for zooming
            const g = svg.append('g');
            const linkGroup = g.append('g').attr('class', 'links');
            const nodeGroup = g.append('g').attr('class', 'nodes');
            const labelGroup = g.append('g').attr('class', 'labels');

            // Color mapping
            const colorMap = {
                'requirement': '#4a90e2',
                'arch': '#7ed321',
                'dsn': '#9b59b6',
                'test-spec': '#e74c3c',
                'test-code': '#e67e22',
                'test-result': '#999'
            };

            // Simulation
            const simulation = d3.forceSimulation(data.nodes)
                .force('link', d3.forceLink(data.edges)
                    .id(d => d.id)
                    .distance(d => 100 + (d.value * 20)))
                .force('charge', d3.forceManyBody().strength(-300))
                .force('center', d3.forceCenter(width / 2, height / 2))
                .force('collision', d3.forceCollide(25));

            // Create links (store globally for filtering)
            globalLink = linkGroup.selectAll('line')
                .data(data.edges)
                .enter()
                .append('line')
                .attr('class', 'link')
                .attr('stroke-width', d => Math.sqrt(d.value))
                .on('mouseover', function() {
                    d3.select(this).classed('active', true);
                })
                .on('mouseout', function() {
                    d3.select(this).classed('active', false);
                });

            // Create nodes (store globally for filtering)
            globalNode = nodeGroup.selectAll('circle')
                .data(data.nodes)
                .enter()
                .append('circle')
                .attr('class', 'node')
                .attr('r', 12)
                .attr('fill', d => {
                    // TestSpec nodes colored by result status
                    if (d.type === 'test-spec' && d.metadata && d.metadata.result_status) {
                        if (d.metadata.result_status === 'pass') return '#28a745';
                        if (d.metadata.result_status === 'fail') return '#dc3545';
                    }
                    return colorMap[d.type] || '#999';
                })
                .call(drag(simulation))
                .on('mouseover', showTooltip)
                .on('mouseout', hideTooltip)
                .on('click', function(event, d) {
                    event.stopPropagation();
                    applyFocus(d.id, focusHops, data);
                });

            // Create labels (store globally for filtering)
            globalLabel = labelGroup.selectAll('text')
                .data(data.nodes)
                .enter()
                .append('text')
                .attr('class', 'node-label')
                .attr('dy', '0.31em')
                .text(d => d.id)
                .style('font-size', '11px');

            // Create edge labels (shown on hover)
            globalEdgeLabel = labelGroup.selectAll('.link-label')
                .data(data.edges)
                .enter()
                .append('text')
                .attr('class', 'link-label')
                .attr('text-anchor', 'middle')
                .text(d => d.label || '')
                .style('pointer-events', 'none')
                .style('opacity', '0');

            // Show edge label on link hover
            globalLink
                .on('mouseover.label', function(event, d) {
                    globalEdgeLabel.filter(e => e === d).style('opacity', '1');
                })
                .on('mouseout.label', function(event, d) {
                    globalEdgeLabel.filter(e => e === d).style('opacity', '0');
                });

            // Zoom behavior
            const zoom = d3.zoom()
                .on('zoom', event => {
                    g.attr('transform', event.transform);
                });

            svg.call(zoom);

            // Zoom buttons
            document.getElementById('btn-zoom-in').addEventListener('click', () => {
                svg.transition().duration(750).call(zoom.scaleBy, 1.3);
            });

            document.getElementById('btn-zoom-out').addEventListener('click', () => {
                svg.transition().duration(750).call(zoom.scaleBy, 0.7);
            });

            document.getElementById('btn-zoom-fit').addEventListener('click', () => {
                const bounds = g.node().getBBox();
                const fullWidth = width;
                const fullHeight = height;
                const midX = bounds.x + bounds.width / 2;
                const midY = bounds.y + bounds.height / 2;

                if (bounds.width > 0 && bounds.height > 0) {
                    const scale = 0.8 / Math.max(bounds.width / fullWidth, bounds.height / fullHeight);
                    const translate = [fullWidth / 2 - scale * midX, fullHeight / 2 - scale * midY];

                    svg.transition()
                        .duration(750)
                        .call(zoom.transform, d3.zoomIdentity.translate(translate[0], translate[1]).scale(scale));
                }
            });

            // Update simulation
            simulation.on('tick', () => {
                globalLink
                    .attr('x1', d => d.source.x)
                    .attr('y1', d => d.source.y)
                    .attr('x2', d => d.target.x)
                    .attr('y2', d => d.target.y);

                globalNode
                    .attr('cx', d => d.x)
                    .attr('cy', d => d.y);

                globalLabel
                    .attr('x', d => d.x)
                    .attr('y', d => d.y);

                if (globalEdgeLabel) {
                    globalEdgeLabel
                        .attr('x', d => (d.source.x + d.target.x) / 2)
                        .attr('y', d => (d.source.y + d.target.y) / 2);
                }
            });

            // Reset view button
            document.getElementById('btn-reset').addEventListener('click', () => clearFocus());

            // Filter buttons
            document.getElementById('btn-filter-all').addEventListener('click', () => {
                updateFilterButtons('all');
                filterByType('all');
            });

            document.getElementById('btn-filter-req').addEventListener('click', () => {
                updateFilterButtons('requirement');
                filterByType('requirement');
            });

            document.getElementById('btn-filter-arch').addEventListener('click', () => {
                updateFilterButtons('arch');
                filterByType('arch');
            });

            document.getElementById('btn-filter-dsn').addEventListener('click', () => {
                updateFilterButtons('dsn');
                filterByType('dsn');
            });

            document.getElementById('btn-filter-tests').addEventListener('click', () => {
                updateFilterButtons('tests');
                filterByType('test-spec', 'test-code', 'test-result');
            });

            function filterByType(...types) {
                if (types.length === 1 && types[0] === 'all') {
                    globalNode.style('display', 'block');
                    globalLabel.style('display', 'block');
                    globalLink.style('display', 'line');
                    if (globalEdgeLabel) globalEdgeLabel.style('opacity', '0');
                    return;
                }
                const typeMap = {};
                types.forEach(t => typeMap[t] = true);

                globalNode.style('display', d => typeMap[d.type] ? 'block' : 'none');
                globalLabel.style('display', d => typeMap[d.type] ? 'block' : 'none');
                globalLink.style('display', d => typeMap[d.source.type] && typeMap[d.target.type] ? 'line' : 'none');
                if (globalEdgeLabel) {
                    globalEdgeLabel.style('opacity', '0');
                }
            }

            function updateFilterButtons(active) {
                document.getElementById('btn-filter-all').classList.remove('active');
                document.getElementById('btn-filter-req').classList.remove('active');
                document.getElementById('btn-filter-arch').classList.remove('active');
                document.getElementById('btn-filter-dsn').classList.remove('active');
                document.getElementById('btn-filter-tests').classList.remove('active');

                if (active === 'all') document.getElementById('btn-filter-all').classList.add('active');
                else if (active === 'requirement') document.getElementById('btn-filter-req').classList.add('active');
                else if (active === 'arch') document.getElementById('btn-filter-arch').classList.add('active');
                else if (active === 'dsn') document.getElementById('btn-filter-dsn').classList.add('active');
                else if (active === 'tests') document.getElementById('btn-filter-tests').classList.add('active');
            }

            function showTooltip(event, d) {
                const tooltip = document.getElementById('tooltip');
                let metaHTML = '';
                const meta = d.metadata || {};
                // Show impl= for requirement and arch nodes
                if (meta.impl) metaHTML += '<strong>impl:</strong> ' + renderImplCell(meta.impl) + '<br>';
                // Show result status for test-spec nodes
                if (d.type === 'test-spec' && meta.result_status) {
                    const badge = meta.result_status === 'pass' ? '🟢 pass' : (meta.result_status === 'fail' ? '🔴 fail' : '🟡 missing');
                    metaHTML += '<strong>result:</strong> ' + badge + '<br>';
                }
                // Show test-result details
                if (d.type === 'test-result') {
                    const statusBadge = meta.status === 'passed' ? '🟢' : (meta.status === 'failed' ? '🔴' : '🟡');
                    metaHTML += '<strong>status:</strong> ' + statusBadge + ' ' + escHtml(meta.status || '') + '<br>';
                    if (meta.duration) metaHTML += '<strong>duration:</strong> ' + escHtml(String(meta.duration)) + 's<br>';
                    if (meta.platform) metaHTML += '<strong>platform:</strong> ' + escHtml(meta.platform) + '<br>';
                    if (meta.error) metaHTML += '<strong>error:</strong> <pre style="margin:4px 0;font-size:11px;white-space:pre-wrap;max-width:300px;color:#c0392b">' + escHtml(meta.error.substring(0, 300)) + '</pre>';
                } else {
                    Object.entries(meta).forEach(([k, v]) => {
                        if (['impl', 'result_status', 'status', 'duration', 'platform', 'error', 'stdout', 'linked_spec', 'linked_code'].includes(k)) return;
                        if (v) metaHTML += '<strong>' + escHtml(k) + ':</strong> ' + escHtml(String(v)) + '<br>';
                    });
                }

                tooltip.innerHTML =
                    '<div class="tooltip-title">' + escHtml(d.label) + '</div>' +
                    '<div class="tooltip-content">' +
                    '<strong>Type:</strong> ' + escHtml(d.type) + '<br>' +
                    '<strong>ID:</strong> ' + escHtml(d.id) +
                    '</div>' +
                    '<div class="tooltip-meta">' + metaHTML + '</div>';

                tooltip.classList.add('visible');
                tooltip.style.left = (event.pageX + 10) + 'px';
                tooltip.style.top = (event.pageY + 10) + 'px';
            }

            function hideTooltip() {
                document.getElementById('tooltip').classList.remove('visible');
            }


            function drag(simulation) {
                function dragstarted(event) {
                    if (!event.active) simulation.alphaTarget(0.3).restart();
                    event.subject.fx = event.subject.x;
                    event.subject.fy = event.subject.y;
                }

                function dragged(event) {
                    event.subject.fx = event.x;
                    event.subject.fy = event.y;
                }

                function dragended(event) {
                    if (!event.active) simulation.alphaTarget(0);
                    event.subject.fx = null;
                    event.subject.fy = null;
                }

                return d3.drag()
                    .on('start', dragstarted)
                    .on('drag', dragged)
                    .on('end', dragended);
            }
        }

        function updateStats(data) {
            if (!data || !data.nodes) {
                console.error('No data or nodes in updateStats:', data);
                return;
            }

            const stats = {
                requirement: 0,
                arch: 0,
                dsn: 0,
                'test-spec': 0,
                'test-code': 0,
                'test-result': 0
            };

            data.nodes.forEach(node => {
                if (node.type in stats) {
                    stats[node.type]++;
                }
            });

            const reqEl = document.getElementById('stat-reqs');
            const archEl = document.getElementById('stat-arch');
            const dsnEl = document.getElementById('stat-dsn');
            const specsEl = document.getElementById('stat-specs');
            const resultsEl = document.getElementById('stat-results');
            const totalEl = document.getElementById('stat-total');
            const linksEl = document.getElementById('stat-links');

            if (reqEl) reqEl.textContent = stats.requirement;
            if (archEl) archEl.textContent = stats.arch;
            if (dsnEl) dsnEl.textContent = stats.dsn;
            if (specsEl) specsEl.textContent = stats['test-spec'];
            if (resultsEl) resultsEl.textContent = stats['test-result'];
            if (totalEl) totalEl.textContent = data.nodes.length;
            if (linksEl) linksEl.textContent = data.edges.length;
        }

        // Tab switching
        function switchTab(tabName) {
            document.querySelectorAll('.tab-content').forEach(el => {
                el.classList.remove('active');
            });
            document.querySelectorAll('.tab-button').forEach(el => {
                el.classList.remove('active');
            });

            if (tabName === 'graph') {
                document.getElementById('graph').classList.add('active');
                document.querySelector('button[onclick="switchTab(\'graph\')"]').classList.add('active');
            } else if (tabName === 'matrix') {
                document.getElementById('matrix').classList.add('active');
                document.querySelector('button[onclick="switchTab(\'matrix\')"]').classList.add('active');
                renderMatrix();
            } else if (tabName === 'aspice') {
                document.getElementById('aspice').classList.add('active');
                document.querySelector('button[onclick="switchTab(\'aspice\')"]').classList.add('active');
                renderASPICE();
            } else if (tabName === 'gaps') {
                document.getElementById('gaps').classList.add('active');
                document.getElementById('btn-tab-gaps').classList.add('active');
                renderGaps();
            } else if (tabName === 'elements') {
                document.getElementById('elements').classList.add('active');
                document.querySelector('button[onclick="switchTab(\'elements\')"]').classList.add('active');
                renderElements();
            } else if (tabName === 'coverage') {
                document.getElementById('coverage').classList.add('active');
                document.querySelector('button[onclick="switchTab(\'coverage\')"]').classList.add('active');
                renderCoverage();
            }
        }

        // ── Coverage Tab ────────────────────────────────────────────────
        let covSortCol = 'pct', covSortAsc = true;

        function covSort(col) {
            covSortAsc = (covSortCol === col) ? !covSortAsc : true;
            covSortCol = col;
            renderCoverage();
        }

        function renderCoverage() {
            if (!coverageData || !coverageData.rows || coverageData.rows.length === 0) {
                document.getElementById('cov-tbody').innerHTML =
                    '<tr><td colspan="5" style="padding:20px;text-align:center;color:#999">' +
                    'No coverage data. Run <code>req42-tracer coverage --coverage coverage.out</code> to include coverage data.' +
                    '</td></tr>';
                document.getElementById('cov-count').textContent = '';
                return;
            }

            const level  = document.getElementById('cov-filter-level').value;
            const search = (document.getElementById('cov-search').value || '').toLowerCase();
            const levelColor = { good: '#27ae60', warning: '#f39c12', danger: '#e74c3c' };

            let rows = coverageData.rows.filter(r => {
                if (level  && r.level !== level) return false;
                if (search && !(r.package || '').toLowerCase().includes(search) &&
                             !(r.arch_id || '').toLowerCase().includes(search) &&
                             !(r.arch_title || '').toLowerCase().includes(search)) return false;
                return true;
            });

            rows.sort((a, b) => {
                const va = a[covSortCol] ?? '', vb = b[covSortCol] ?? '';
                if (typeof va === 'number') return covSortAsc ? va - vb : vb - va;
                return covSortAsc ? String(va).localeCompare(String(vb)) : String(vb).localeCompare(String(va));
            });

            document.getElementById('cov-count').textContent = rows.length + ' package(s)';

            document.getElementById('cov-tbody').innerHTML = rows.map(r => {
                const color = levelColor[r.level] || '#999';
                const pct = (r.pct || 0).toFixed(1);
                const bar = '<div style="display:inline-block;width:80px;height:8px;background:#eee;border-radius:3px;vertical-align:middle">' +
                            '<div style="width:' + Math.min(r.pct||0,100).toFixed(1) + '%;height:8px;background:' + color + ';border-radius:3px"></div></div>';
                const arch = r.arch_id
                    ? '<span style="background:#7ed321;color:#fff;border-radius:3px;padding:1px 5px;font-size:11px">' + escHtml(r.arch_id) + '</span>'
                    : '<span style="color:#ddd;font-size:11px">—</span>';
                return '<tr style="border-bottom:1px solid #f0f0f0">' +
                    '<td style="padding:7px 8px;border:1px solid #ddd">' + arch + (r.arch_title ? ' <small style="color:#888">' + escHtml(r.arch_title) + '</small>' : '') + '</td>' +
                    '<td style="padding:7px 8px;border:1px solid #ddd;font-family:monospace;font-size:12px">' + escHtml(r.package || '') + '</td>' +
                    '<td style="padding:7px 8px;border:1px solid #ddd;text-align:right">' + (r.statements || 0) + '</td>' +
                    '<td style="padding:7px 8px;border:1px solid #ddd;text-align:right">' + (r.covered || 0) + '</td>' +
                    '<td style="padding:7px 8px;border:1px solid #ddd">' +
                        '<span style="font-weight:600;color:' + color + ';margin-right:8px">' + pct + '%</span>' + bar +
                    '</td></tr>';
            }).join('');

            const totStmts = rows.reduce((s,r) => s + (r.statements||0), 0);
            const totCov   = rows.reduce((s,r) => s + (r.covered||0), 0);
            const totPct   = totStmts > 0 ? (totCov/totStmts*100).toFixed(1) : '0.0';
            document.getElementById('cov-tfoot').innerHTML =
                '<tr style="background:#f8f9fa;font-weight:700;border-top:2px solid #ddd">' +
                '<td colspan="2" style="padding:8px 12px;border:1px solid #ddd">Total (' + rows.length + ' packages)</td>' +
                '<td style="padding:8px;border:1px solid #ddd;text-align:right">' + totStmts + '</td>' +
                '<td style="padding:8px;border:1px solid #ddd;text-align:right">' + totCov + '</td>' +
                '<td style="padding:8px;border:1px solid #ddd">' + totPct + '%</td></tr>';

            // Summary cards
            const lvl = { good: 0, warning: 0, danger: 0 };
            rows.forEach(r => { if (lvl[r.level] !== undefined) lvl[r.level]++; });
            const ovPct = (coverageData.overall_pct || 0).toFixed(1);
            const ovColor = levelColor[coverageData.overall_level] || '#667eea';
            document.getElementById('cov-cards').innerHTML =
                covCard(ovColor, 'Overall', ovPct + '%') +
                covCard(levelColor.good, '≥80% packages', lvl.good) +
                covCard(levelColor.warning, '70–80% packages', lvl.warning) +
                covCard(levelColor.danger, '<70% packages', lvl.danger);
        }

        function covCard(color, label, value) {
            return '<div style="background:white;border-radius:4px;padding:12px;box-shadow:0 1px 3px rgba(0,0,0,.08);border-left:4px solid ' + color + '">' +
                '<div style="font-size:11px;text-transform:uppercase;color:#888;margin-bottom:4px">' + label + '</div>' +
                '<div style="font-size:24px;font-weight:700;color:' + color + '">' + value + '</div></div>';
        }

        document.getElementById('cov-filter-level').addEventListener('change', renderCoverage);
        document.getElementById('cov-search').addEventListener('input', renderCoverage);

        // ── Elements Tab ────────────────────────────────────────────────
        let elSortCol = 'id', elSortAsc = true;

        function elSort(col) {
            if (elSortCol === col) { elSortAsc = !elSortAsc; } else { elSortCol = col; elSortAsc = true; }
            renderElements();
        }

        function renderElements() {
            const typeFilter = document.getElementById('el-filter-type').value;
            const search = (document.getElementById('el-search').value || '').toLowerCase();
            const typeColor = { req: '#4a90e2', arch: '#7ed321', dsn: '#9b59b6', 'test-spec': '#e74c3c', 'test-result': '#999' };

            let items = (elementsData.items || []).filter(it => {
                if (typeFilter && it.type !== typeFilter) return false;
                if (search && !it.id.toLowerCase().includes(search) && !it.title.toLowerCase().includes(search)) return false;
                return true;
            });

            items = items.slice().sort((a, b) => {
                const va = (elSortCol === 'id' ? a.id : a.title) || '';
                const vb = (elSortCol === 'id' ? b.id : b.title) || '';
                return elSortAsc ? va.localeCompare(vb) : vb.localeCompare(va);
            });

            document.getElementById('el-count').textContent = items.length + ' element(s)';

            const traceCell = (ids) => {
                if (!ids || ids.length === 0) return '<td style="text-align:center;border:1px solid #ddd;color:#ccc">—</td>';
                const tip = ids.slice(0, 5).join(', ') + (ids.length > 5 ? ', …' : '');
                return '<td style="text-align:center;border:1px solid #ddd;cursor:default" title="' + escHtml(tip) + '">' +
                       '<span style="background:#e8f4fd;border-radius:10px;padding:2px 8px;font-size:12px">' + ids.length + '</span></td>';
            };

            const rows = items.map(it => {
                const color = typeColor[it.type] || '#999';
                const badge = '<span style="background:' + color + ';color:#fff;border-radius:3px;padding:1px 6px;font-size:11px">' + escHtml(it.type) + '</span>';
                const meta = escHtml(it.impl || it.status || '');
                const traceUpMissing = (!it.trace_up || it.trace_up.length === 0) && it.type !== 'req';
                const rowStyle = traceUpMissing ? 'background:#fff8f8' : '';
                return '<tr style="border-bottom:1px solid #f0f0f0;' + rowStyle + '">' +
                    '<td style="padding:7px 8px;border:1px solid #ddd;font-family:monospace;font-size:12px">' + escHtml(it.id) + '</td>' +
                    '<td style="padding:7px 8px;border:1px solid #ddd">' + badge + '</td>' +
                    '<td style="padding:7px 8px;border:1px solid #ddd">' + escHtml(it.title || '') + '</td>' +
                    '<td style="padding:7px 8px;border:1px solid #ddd;font-size:12px;color:#666">' + meta + '</td>' +
                    traceCell(it.trace_up) +
                    traceCell(it.trace_down) +
                    '</tr>';
            }).join('');

            document.getElementById('el-tbody').innerHTML = rows || '<tr><td colspan="6" style="padding:20px;text-align:center;color:#999">No elements match filter.</td></tr>';
        }

        document.getElementById('el-filter-type').addEventListener('change', renderElements);
        document.getElementById('el-search').addEventListener('input', renderElements);

        function renderGaps() {
            const container = document.getElementById('gaps-content');
            if (!container) return;

            function gapSection(title, items, emptyMsg) {
                if (items.length === 0) {
                    return '<div style="margin-bottom:24px"><h3 style="color:#27ae60;margin-bottom:8px">✓ ' + escHtml(title) + ' (0)</h3><p style="color:#999;font-size:13px">' + escHtml(emptyMsg) + '</p></div>';
                }
                let rows = items.map(item =>
                    '<tr><td style="padding:6px 12px;font-family:monospace;font-size:13px">' + escHtml(item.id) +
                    '</td><td style="padding:6px 12px;font-size:13px">' + escHtml(item.title) +
                    '</td><td style="padding:6px 12px;font-size:12px;color:#888">' + escHtml(item.info || '') + '</td></tr>'
                ).join('');
                return '<div style="margin-bottom:24px"><h3 style="color:#e74c3c;margin-bottom:8px">✗ ' + escHtml(title) + ' (' + items.length + ')</h3>' +
                    '<table style="width:100%;border-collapse:collapse;background:#fff;border:1px solid #e0e0e0;border-radius:4px">' +
                    '<thead><tr style="background:#f8f9fa"><th style="padding:6px 12px;text-align:left;font-size:12px;color:#666">ID</th>' +
                    '<th style="padding:6px 12px;text-align:left;font-size:12px;color:#666">Title</th>' +
                    '<th style="padding:6px 12px;text-align:left;font-size:12px;color:#666">Info</th></tr></thead>' +
                    '<tbody>' + rows + '</tbody></table></div>';
            }

            const html = (!gapsData.has_gaps)
                ? '<div style="color:#27ae60;font-size:16px;padding:20px 0">✓ No gaps found — full traceability achieved.</div>'
                : gapSection('Orphan Requirements (no arch/test coverage)', gapsData.orphan_requirements, 'All requirements are covered.') +
                  gapSection('Orphan Architecture Elements (no requirement)', gapsData.orphan_arch_elements, 'All arch elements are traced.') +
                  gapSection('Orphan Design Elements — SWE.3 (no arch parent)', gapsData.orphan_design_elements, 'All design elements have arch parents.') +
                  gapSection('Untested Architecture Elements — SWE.2 (no integration test)', gapsData.untested_arch_elements, 'All arch elements have integration tests.') +
                  gapSection('Untested Design Elements — SWE.4 (no unit test)', gapsData.untested_design_elements, 'All design elements have unit tests.') +
                  gapSection('Orphan Test Specs (no requirement link)', gapsData.orphan_test_specs, 'All test specs are linked.') +
                  gapSection('Untraced Test Results (no linked spec)', gapsData.untraced_test_results, 'All test results are traced.');

            container.innerHTML = html;
        }

        function escHtml(s) {
            return String(s)
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/"/g, '&quot;')
                .replace(/'/g, '&#39;');
        }

        function coverageClass(pct) {
            if (pct >= 80) return 'good';
            if (pct >= 50) return 'warning';
            return 'danger';
        }

        function renderASPICE() {
            if (!aspiceData) return;

            var overall = aspiceData.overall || 0;
            var overallPct = Math.round(overall);
            var cls = coverageClass(overall);

            var pctEl = document.getElementById('aspice-overall-pct');
            var fillEl = document.getElementById('aspice-overall-fill');
            if (pctEl) {
                pctEl.textContent = overallPct + '%';
                pctEl.className = 'overall-pct ' + cls;
            }
            if (fillEl) {
                fillEl.style.width = overallPct + '%';
                fillEl.className = 'overall-bar-fill ' + cls;
            }

            var grid = document.getElementById('aspice-grid');
            if (!grid || !aspiceData.processes) return;

            var filterEl = document.getElementById('filter-process');
            var filterVal = filterEl ? filterEl.value : '';

            // Populate dropdown on first render
            if (filterEl && filterEl.options.length <= 1) {
                aspiceData.processes.forEach(function(proc) {
                    var opt = document.createElement('option');
                    opt.value = proc.id;
                    opt.textContent = proc.id + ' — ' + proc.name;
                    filterEl.appendChild(opt);
                });
            }

            var html = '';
            aspiceData.processes.forEach(function(proc) {
                if (filterVal && proc.id !== filterVal) return;
                var covPct = Math.round(proc.coverage || 0);
                var pcls = coverageClass(proc.coverage || 0);

                var bpHTML = '';
                if (proc.bps) {
                    proc.bps.forEach(function(bp) {
                        var bpPct = Math.round(bp.coverage || 0);
                        var bcls = coverageClass(bp.coverage || 0);
                        var gapsHTML = '';
                        if (bp.gaps && bp.gaps.length > 0) {
                            gapsHTML = '<ul class="bp-gaps">';
                            bp.gaps.forEach(function(g) {
                                gapsHTML += '<li>' + escHtml(g) + '</li>';
                            });
                            gapsHTML += '</ul>';
                        }
                        bpHTML += '<div class="bp-item">' +
                            '<div class="bp-row">' +
                            '<span class="bp-id">' + escHtml(bp.id) + '</span>' +
                            '<span class="bp-title">' + escHtml(bp.title) + '</span>' +
                            '<span class="bp-badge ' + bcls + '">' + bpPct + '%</span>' +
                            '</div>' +
                            gapsHTML +
                            '</div>';
                    });
                }

                html += '<div class="process-card">' +
                    '<div class="process-card-header ' + pcls + '">' +
                    '<span class="process-id-badge">' + escHtml(proc.id) + '</span>' +
                    '<span class="process-card-name">' + escHtml(proc.name) + '</span>' +
                    '<span class="process-card-pct ' + pcls + '">' + covPct + '%</span>' +
                    '</div>' +
                    '<div class="process-progress">' +
                    '<div class="process-progress-fill ' + pcls + '" style="width:' + covPct + '%"></div>' +
                    '</div>' +
                    '<div class="bp-list">' + bpHTML + '</div>' +
                    '</div>';
            });

            grid.innerHTML = html || '<p style="padding:20px;color:#999;">No ASPICE process data available.</p>';
        }

        let currentMatrixData = matrixData;
        let matrixSortCol = null;
        let matrixSortAsc = true;

        function sortMatrix(col) {
            if (matrixSortCol === col) {
                matrixSortAsc = !matrixSortAsc;
            } else {
                matrixSortCol = col;
                matrixSortAsc = true;
            }
            renderMatrix();
        }

        function renderMatrix() {
            const container = document.getElementById('matrix-table-container');
            const priority = document.getElementById('filter-priority').value;
            const status = document.getElementById('filter-status').value;
            const search = document.getElementById('search-req').value.toLowerCase();

            let filtered = matrixData.rows.filter(row => {
                const matchesPriority = !priority || row.Priority === priority;
                const matchesStatus = !status || row.Status === status;
                const matchesSearch = !search || row.RequirementID.toLowerCase().includes(search) || row.Title.toLowerCase().includes(search);
                return matchesPriority && matchesStatus && matchesSearch;
            });

            if (matrixSortCol) {
                filtered = filtered.slice().sort((a, b) => {
                    let va = '', vb = '';
                    if (matrixSortCol === 'req') { va = a.RequirementID; vb = b.RequirementID; }
                    else if (matrixSortCol === 'priority') { va = a.Priority; vb = b.Priority; }
                    else if (matrixSortCol === 'status') { va = a.Status; vb = b.Status; }
                    else {
                        const ca = a.Cells[matrixSortCol], cb = b.Cells[matrixSortCol];
                        va = ca ? ca.Status : 'missing'; vb = cb ? cb.Status : 'missing';
                    }
                    return matrixSortAsc ? va.localeCompare(vb) : vb.localeCompare(va);
                });
            }

            currentMatrixData = { ...matrixData, rows: filtered };

            const sortArrow = (col) => {
                if (matrixSortCol !== col) return ' ↕';
                return matrixSortAsc ? ' ↑' : ' ↓';
            };

            let html = '<table class="matrix-table"><thead><tr>';
            html += '<th style="cursor:pointer" onclick="sortMatrix(\'req\')">Requirement' + sortArrow('req') + '</th>';
            html += '<th style="cursor:pointer" onclick="sortMatrix(\'priority\')">Priority' + sortArrow('priority') + '</th>';
            html += '<th style="cursor:pointer" onclick="sortMatrix(\'status\')">Status' + sortArrow('status') + '</th>';
            html += '<th title="Implementation package (impl= attribute)">Impl</th>';
            html += '<th title="Overall coverage: Arch + TestSpec + TestResult">Coverage</th>';
            matrixData.columns.forEach(col => {
                const typeColor = col.Type === 'arch' ? '#7ed321' : (col.Type === 'dsn' ? '#9b59b6' : (col.Type === 'test-code' ? '#e67e22' : '#e74c3c'));
                html += '<th style="cursor:pointer;border-top:3px solid ' + typeColor + '" title="[' + escHtml(col.Type) + '] ' + escHtml(col.Title) + '" onclick="sortMatrix(\'' + escHtml(col.ID) + '\')">' + escHtml(col.ID) + sortArrow(col.ID) + '</th>';
            });
            html += '</tr></thead><tbody>';

            filtered.forEach(row => {
                html += '<tr>';
                html += '<td class="req-id" title="' + escHtml(row.Title) + '">' + escHtml(row.RequirementID) + '</td>';
                html += '<td>' + escHtml(row.Priority) + '</td>';
                html += '<td>' + escHtml(row.Status) + '</td>';

                // impl= column
                const implVal = row.Impl || '';
                html += '<td class="impl-cell" title="' + escHtml(implVal) + '">' + renderImplCell(implVal) + '</td>';

                // Coverage badge: 🟢 pass / 🟡 missing / 🔴 fail
                const trStatus = row.TestResultStatus || 'missing';
                const badge = trStatus === 'pass' ? '🟢' : (trStatus === 'fail' ? '🔴' : '🟡');
                const badgeTitle = trStatus === 'pass' ? 'All tests passing' : (trStatus === 'fail' ? 'Test failure detected' : 'No test results found');
                html += '<td class="coverage-badge tr-' + trStatus + '" title="' + badgeTitle + '">' + badge + '</td>';

                matrixData.columns.forEach(col => {
                    const cell = row.Cells[col.ID];
                    const status = cell ? cell.Status : 'missing';
                    const symbol = status === 'covered' ? '✓' : (status === 'stale' ? '⚠' : '✗');
                    html += '<td class="matrix-cell ' + status + '" title="' + escHtml(cell ? cell.Evidence : 'Not covered') + '">' + symbol + '</td>';
                });

                html += '</tr>';
            });

            html += '</tbody></table>';
            container.innerHTML = html;

            // Update statistics
            const stats = matrixData.statistics;
            document.getElementById('matrix-total').textContent = stats.TotalRequirements;
            document.getElementById('matrix-covered').textContent = stats.CoveredRequirements;
            document.getElementById('matrix-missing').textContent = stats.MissingRequirements;
            document.getElementById('matrix-coverage').textContent = Math.round(stats.CoveragePercentage) + '%';
        }

        function csvQuote(s) {
            const str = String(s);
            if (str.includes(',') || str.includes('"') || str.includes('\n')) {
                return '"' + str.replace(/"/g, '""') + '"';
            }
            return str;
        }

        function exportMatrixCSV() {
            let csv = 'Requirement,Priority,Status';
            matrixData.columns.forEach(col => {
                csv += ',' + csvQuote(col.ID);
            });
            csv += '\n';

            currentMatrixData.rows.forEach(row => {
                csv += csvQuote(row.RequirementID) + ',' + csvQuote(row.Priority) + ',' + csvQuote(row.Status);
                matrixData.columns.forEach(col => {
                    const cell = row.Cells[col.ID];
                    const symbol = cell && cell.Status === 'covered' ? '✓' : (cell && cell.Status === 'stale' ? '⚠' : '✗');
                    csv += ',' + symbol;
                });
                csv += '\n';
            });

            const blob = new Blob([csv], { type: 'text/csv' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'traceability-matrix.csv';
            a.click();
            window.URL.revokeObjectURL(url);
        }

        // Wire up filter controls
        document.getElementById('filter-priority').addEventListener('change', renderMatrix);
        document.getElementById('filter-status').addEventListener('change', renderMatrix);
        document.getElementById('search-req').addEventListener('input', renderMatrix);

        // Initialize graph visualization after all functions are defined
        function bootGraph() {
            if (graphData && graphData.nodes) {
                initializeGraph(graphData);
                renderSidebarSelector(graphData);
                initHopControls(graphData);
                restoreHashState(graphData);
            }
            if (gapsData && gapsData.has_gaps) {
                const btn = document.getElementById('btn-tab-gaps');
                if (btn) btn.style.color = '#e74c3c';
            }
        }

        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', bootGraph);
        } else {
            bootGraph();
        }
    </script>
</body>
</html>
`
