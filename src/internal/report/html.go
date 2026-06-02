package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paulefl/req42-tracer/src/internal/graph"
	"github.com/paulefl/req42-tracer/src/internal/model"
)

// HTMLReporter generates interactive HTML reports with D3.js graph visualization.
type HTMLReporter struct {
	analyzer   *graph.Analyzer
	config     *model.Config
	outputPath string
}

// NewHTMLReporter creates a new HTML reporter.
func NewHTMLReporter(analyzer *graph.Analyzer, config *model.Config, outputPath string) *HTMLReporter {
	return &HTMLReporter{
		analyzer:   analyzer,
		config:     config,
		outputPath: outputPath,
	}
}

// GenerateReport creates an interactive HTML report with graph visualization and matrix.
func (hr *HTMLReporter) GenerateReport() error {
	// Get the traceability graph
	g := hr.analyzer.GetGraph()

	// Export graph data for D3.js
	graphData := ExportGraphData(g)

	// Build matrix data
	matrixData := BuildMatrixData(g)

	// Serialize graph data to JSON
	graphJSON, err := json.MarshalIndent(graphData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal graph data: %w", err)
	}

	// Serialize matrix data to JSON
	matrixJSON, err := json.MarshalIndent(matrixData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal matrix data: %w", err)
	}

	// Build and serialize ASPICE dashboard data
	aspiceData := BuildASPICEDashboardData(hr.analyzer, hr.config)
	aspiceJSON, err := json.MarshalIndent(aspiceData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal aspice data: %w", err)
	}

	// Build and serialize gaps data
	gapsData := BuildGapsData(hr.analyzer.AnalyzeGaps())
	gapsJSON, err := json.MarshalIndent(gapsData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal gaps data: %w", err)
	}

	// Build and serialize elements data
	elementsData := BuildElementsData(g)
	elementsJSON, err := json.MarshalIndent(elementsData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal elements data: %w", err)
	}

	// Coverage tab: empty by default (populated via `req42-tracer coverage`)
	coverageData := &CoverageData{Rows: []CoverageRow{}}
	coverageJSON, err := json.MarshalIndent(coverageData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal coverage data: %w", err)
	}

	// Create output directory if needed
	dir := filepath.Dir(hr.outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Inject source-base-url as a JS constant so impl= links become clickable.
	sourceBaseURL := hr.config.Reports.HTML.SourceBaseURL
	sourceBaseURLJS := "\"" + strings.ReplaceAll(sourceBaseURL, `"`, `\"`) + "\""

	// Generate HTML by replacing placeholders in template
	htmlContent := strings.ReplaceAll(HTMLTemplate, "<!--GRAPH_DATA_JSON-->", string(graphJSON))
	htmlContent = strings.ReplaceAll(htmlContent, "<!--MATRIX_DATA_JSON-->", string(matrixJSON))
	htmlContent = strings.ReplaceAll(htmlContent, "<!--ASPICE_DATA_JSON-->", string(aspiceJSON))
	htmlContent = strings.ReplaceAll(htmlContent, "<!--GAPS_DATA_JSON-->", string(gapsJSON))
	htmlContent = strings.ReplaceAll(htmlContent, "<!--ELEMENTS_DATA_JSON-->", string(elementsJSON))
	htmlContent = strings.ReplaceAll(htmlContent, "<!--COVERAGE_DATA_JSON-->", string(coverageJSON))
	htmlContent = strings.ReplaceAll(htmlContent, "\"<!--SOURCE_BASE_URL-->\"", sourceBaseURLJS)

	// Write HTML file
	if err := os.WriteFile(hr.outputPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write HTML report: %w", err)
	}

	return nil
}

// GenerateSummaryReport creates an additional summary HTML with statistics.
func (hr *HTMLReporter) GenerateSummaryReport(summaryPath string) error {
	g := hr.analyzer.GetGraph()

	// Calculate statistics
	totalReqs := len(g.Requirements)
	totalArch := len(g.ArchElements)
	totalTests := len(g.TestSpecs)
	totalResults := len(g.TestResults)

	coveredReqs := 0
	for _, req := range g.Requirements {
		for _, link := range g.Links {
			if link.FromID == req.ID && link.FromType == "requirement" {
				coveredReqs++
				break
			}
		}
	}

	coveredArch := 0
	for _, arch := range g.ArchElements {
		for _, link := range g.Links {
			if link.FromID == arch.ID && link.FromType == "arch" {
				coveredArch++
				break
			}
		}
	}

	testedSpecs := 0
	for _, spec := range g.TestSpecs {
		for _, link := range g.Links {
			if link.FromID == spec.ID && link.FromType == "test-spec" {
				testedSpecs++
				break
			}
		}
	}

	// Create summary HTML
	html := generateSummaryHTML(
		totalReqs, coveredReqs,
		totalArch, coveredArch,
		totalTests, testedSpecs,
		totalResults,
		len(g.Links),
		filepath.Base(hr.outputPath),
	)

	if err := os.WriteFile(summaryPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write summary report: %w", err)
	}

	return nil
}

func generateSummaryHTML(totalReqs, coveredReqs, totalArch, coveredArch,
	totalTests, testedSpecs, totalResults, totalLinks int, graphFilename string) string {
	pct := func(covered, total int) float64 {
		if total == 0 {
			return 0
		}
		return float64(covered) / float64(total) * 100
	}
	var b strings.Builder
	b.WriteString(summaryHTMLHeader())
	b.WriteString(summaryMetricsGrid(totalReqs, coveredReqs, totalArch, coveredArch,
		totalTests, testedSpecs, totalResults, totalLinks, pct))
	b.WriteString(summaryCoverageDetails(totalReqs, coveredReqs, totalArch, coveredArch,
		totalTests, testedSpecs, totalLinks))
	b.WriteString(summaryRecommendations(totalReqs-coveredReqs, totalTests-testedSpecs))
	b.WriteString(summaryHTMLFooter(graphFilename))
	return b.String()
}

func summaryHTMLHeader() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>req42-tracer: Coverage Summary</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
               background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
               color: #333; min-height: 100vh; padding: 20px; }
        .container { max-width: 1000px; margin: 0 auto; background: white;
                     border-radius: 8px; box-shadow: 0 10px 40px rgba(0,0,0,0.2); overflow: hidden; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                  color: white; padding: 40px 20px; text-align: center; }
        .header h1 { font-size: 32px; margin-bottom: 10px; }
        .header p  { font-size: 16px; opacity: 0.9; }
        .content { padding: 40px; }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
                        gap: 20px; margin-bottom: 40px; }
        .metric { background: #f8f9fa; border-left: 4px solid #667eea; padding: 20px; border-radius: 4px; }
        .metric.good { border-left-color: #7ed321; }
        .metric.warning { border-left-color: #f5a623; }
        .metric.danger { border-left-color: #e74c3c; }
        .metric-label { font-size: 12px; color: #666; text-transform: uppercase; margin-bottom: 8px; }
        .metric-value { font-size: 28px; font-weight: bold; color: #333; margin-bottom: 4px; }
        .metric-subtext { font-size: 13px; color: #999; }
        .coverage-bar { width: 100%; height: 8px; background: #e0e0e0;
                        border-radius: 4px; margin-top: 10px; overflow: hidden; }
        .coverage-fill { height: 100%; background: #7ed321; transition: width 0.3s; }
        .summary-section { margin-bottom: 40px; }
        .summary-section h2 { font-size: 20px; margin-bottom: 20px; color: #333;
                              border-bottom: 2px solid #667eea; padding-bottom: 10px; }
        .summary-table { width: 100%; border-collapse: collapse; }
        .summary-table tr { border-bottom: 1px solid #e0e0e0; }
        .summary-table tr:hover { background: #f8f9fa; }
        .summary-table td { padding: 12px; text-align: left; }
        .summary-table td:first-child { font-weight: 500; width: 40%; }
        .summary-table td:last-child { text-align: right; font-weight: bold; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center;
                  color: #666; font-size: 12px; border-top: 1px solid #e0e0e0; }
        .footer a { color: #667eea; text-decoration: none; }
        .footer a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>req42-tracer Coverage Summary</h1>
            <p>Traceability and test coverage metrics</p>
        </div>
        <div class="content">
`
}

func summaryMetricsGrid(totalReqs, coveredReqs, totalArch, coveredArch,
	totalTests, testedSpecs, totalResults, totalLinks int,
	pct func(int, int) float64) string {
	card := func(level, label, subtext string, value float64) string {
		return fmt.Sprintf(`            <div class="metric %s">
                <div class="metric-label">%s</div>
                <div class="metric-value">%.1f%%</div>
                <div class="metric-subtext">%s</div>
                <div class="coverage-bar"><div class="coverage-fill" style="width: %.1f%%"></div></div>
            </div>
`, level, label, value, subtext, value)
	}
	stat := func(label, subtext string, value int) string {
		return fmt.Sprintf(`            <div class="metric">
                <div class="metric-label">%s</div>
                <div class="metric-value">%d</div>
                <div class="metric-subtext">%s</div>
            </div>
`, label, value, subtext)
	}
	var b strings.Builder
	b.WriteString(`        <div class="metrics-grid">` + "\n")
	b.WriteString(card(getCoverageLevel(pct(coveredReqs, totalReqs)), "Requirement Coverage",
		fmt.Sprintf("%d of %d requirements", coveredReqs, totalReqs), pct(coveredReqs, totalReqs)))
	b.WriteString(card(getCoverageLevel(pct(coveredArch, totalArch)), "Architecture Coverage",
		fmt.Sprintf("%d of %d components", coveredArch, totalArch), pct(coveredArch, totalArch)))
	b.WriteString(card(getCoverageLevel(pct(testedSpecs, totalTests)), "Test Coverage",
		fmt.Sprintf("%d of %d specs", testedSpecs, totalTests), pct(testedSpecs, totalTests)))
	b.WriteString(stat("Total Artifacts", "across all types", totalReqs+totalArch+totalTests+totalResults))
	b.WriteString(stat("Trace Links", "req → arch → test", totalLinks))
	b.WriteString(stat("Test Results", "from CI/CD", totalResults))
	b.WriteString("        </div>\n")
	return b.String()
}

func summaryCoverageDetails(totalReqs, coveredReqs, totalArch, coveredArch,
	totalTests, testedSpecs, totalLinks int) string {
	return fmt.Sprintf(`        <div class="summary-section">
            <h2>Coverage Details</h2>
            <table class="summary-table">
                <tr><td>Requirements Covered</td><td>%d / %d</td></tr>
                <tr><td>Architecture Components</td><td>%d / %d</td></tr>
                <tr><td>Test Specifications Tested</td><td>%d / %d</td></tr>
                <tr><td>Total Trace Links</td><td>%d</td></tr>
            </table>
        </div>
`, coveredReqs, totalReqs, coveredArch, totalArch, testedSpecs, totalTests, totalLinks)
}

func summaryRecommendations(uncoveredReqs, untestedSpecs int) string {
	return fmt.Sprintf(`        <div class="summary-section">
            <h2>Recommendations</h2>
            <table class="summary-table">
                <tr><td>Uncovered Requirements</td><td>%d</td></tr>
                <tr><td>Untested Specifications</td><td>%d</td></tr>
            </table>
        </div>
`, uncoveredReqs, untestedSpecs)
}

func summaryHTMLFooter(graphFilename string) string {
	return fmt.Sprintf(`        </div>
        <div class="footer">
            Generated by <a href="https://github.com/paulefl/req42-tracer">req42-tracer</a> •
            <a href="%s">View Interactive Graph</a>
        </div>
    </div>
</body>
</html>`, graphFilename)
}

func getCoverageLevel(percentage float64) string {
	if percentage >= 80 {
		return "good"
	} else if percentage >= 50 {
		return "warning"
	}
	return "danger"
}

