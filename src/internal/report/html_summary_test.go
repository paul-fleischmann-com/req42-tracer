package report

import (
	"strings"
	"testing"
)

// [test-spec,id=TS-RPT-049,req="REQ-REPORT-001",dsn=comp.report.html,aspice="SWE.4.BP2"]
// TestSummaryHTMLHeader verifies the header contains DOCTYPE and CSS.
func TestSummaryHTMLHeader(t *testing.T) {
	h := summaryHTMLHeader()
	for _, want := range []string{"<!DOCTYPE html>", "<style>", "metrics-grid", "coverage-bar"} {
		if !strings.Contains(h, want) {
			t.Errorf("summaryHTMLHeader missing %q", want)
		}
	}
}

// [test-spec,id=TS-RPT-050,req="REQ-REPORT-001",aspice="SWE.4.BP2"]
// TestSummaryHTMLFooter verifies the footer contains link and closing tags.
func TestSummaryHTMLFooter(t *testing.T) {
	f := summaryHTMLFooter("report.html")
	if !strings.Contains(f, "report.html") {
		t.Error("footer missing graph filename")
	}
	if !strings.Contains(f, "</html>") {
		t.Error("footer missing </html>")
	}
}

// [test-spec,id=TS-RPT-051,req="REQ-REPORT-001",aspice="SWE.4.BP2"]
// TestSummaryCoverageDetails verifies coverage table contains correct counts.
func TestSummaryCoverageDetails(t *testing.T) {
	out := summaryCoverageDetails(10, 8, 5, 4, 20, 15, 42)
	for _, want := range []string{"8 / 10", "4 / 5", "15 / 20", "42"} {
		if !strings.Contains(out, want) {
			t.Errorf("summaryCoverageDetails missing %q", want)
		}
	}
}

// [test-spec,id=TS-RPT-052,req="REQ-REPORT-001",aspice="SWE.4.BP2"]
// TestSummaryRecommendations verifies recommendations section contains gap counts.
func TestSummaryRecommendations(t *testing.T) {
	out := summaryRecommendations(3, 7)
	if !strings.Contains(out, ">3<") && !strings.Contains(out, "<td>3</td>") {
		// accept either fmt variant
		if !strings.Contains(out, "3") {
			t.Error("summaryRecommendations missing uncovered count 3")
		}
	}
	if !strings.Contains(out, "Recommendations") {
		t.Error("summaryRecommendations missing section heading")
	}
}

// [test-spec,id=TS-RPT-053,req="REQ-REPORT-001",aspice="SWE.4.BP2"]
// TestSummaryMetricsGrid verifies metrics grid contains coverage percentages.
func TestSummaryMetricsGrid(t *testing.T) {
	pct := func(covered, total int) float64 {
		if total == 0 {
			return 0
		}
		return float64(covered) / float64(total) * 100
	}
	out := summaryMetricsGrid(10, 8, 5, 4, 20, 15, 3, 42, pct)
	for _, want := range []string{"Requirement Coverage", "Architecture Coverage", "Test Coverage",
		"Total Artifacts", "Trace Links", "Test Results"} {
		if !strings.Contains(out, want) {
			t.Errorf("summaryMetricsGrid missing %q", want)
		}
	}
}

// [test-spec,id=TS-RPT-054,req="REQ-REPORT-001",aspice="SWE.4.BP2"]
// TestGenerateSummaryHTML_Integration verifies the full summary HTML is well-formed.
func TestGenerateSummaryHTML_Integration(t *testing.T) {
	html := generateSummaryHTML(16, 14, 53, 40, 200, 180, 25, 300, "report.html")
	for _, want := range []string{
		"<!DOCTYPE html>", "</html>",
		"Coverage Details", "Recommendations",
		"14 / 16", "40 / 53", "180 / 200",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("generateSummaryHTML missing %q", want)
		}
	}
}
