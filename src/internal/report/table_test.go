package report

import (
	"strings"
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/graph"
	"github.com/paulefl/req42-tracer/src/internal/model"
)

func buildReportGraph() *graph.Analyzer {
	g := &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-001": {ID: "REQ-001", Title: "Requirement One", Priority: "high", Status: "approved"},
			"REQ-002": {ID: "REQ-002", Title: "Requirement Two", Priority: "low", Status: "draft"},
		},
		ArchElements: map[string]*model.ArchElement{
			"comp.api":    {ID: "comp.api", Title: "API", Parent: "comp.system", Req: []string{"REQ-001"}, Impl: "internal/api.go"},
			"comp.system": {ID: "comp.system", Title: "System"},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"TS-001": {ID: "TS-001", Title: "API Test", Req: []string{"REQ-001"}, Arch: []string{"comp.api"}},
		},
		TestCodes: make(map[string]*model.TestCode),
		TestResults: map[string]*model.TestResult{
			"linux::pkg::TestA": {ID: "linux::pkg::TestA", TestName: "TestA", Status: "passed"},
			"linux::pkg::TestB": {ID: "linux::pkg::TestB", TestName: "TestB", Status: "failed"},
		},
		Links: []*model.TraceLink{
			{FromID: "REQ-001", ToID: "comp.api", FromType: "requirement", ToType: "arch", LinkType: "satisfied-by", Status: "active"},
			{FromID: "comp.api", ToID: "TS-001", FromType: "arch", ToType: "test-spec", LinkType: "verified-by", Status: "active"},
		},
	}
	return graph.NewAnalyzer(g)
}

// [test-spec,id=TS-RPT-001,req="REQ-REPORT-001",dsn=comp.report.table,aspice=SWE.4-BP4]
// TestTableReporter_TraceabilityMatrix_Text verifies text format traceability matrix output.
func TestTableReporter_TraceabilityMatrix_Text(t *testing.T) {
	tr := NewTableReporter(buildReportGraph(), "text")
	out := tr.TraceabilityMatrix()
	if !strings.Contains(out, "TRACEABILITY MATRIX") {
		t.Error("expected TRACEABILITY MATRIX header")
	}
	if !strings.Contains(out, "REQ-001") {
		t.Error("expected REQ-001 in output")
	}
}

// [test-spec,id=TS-RPT-002,req="REQ-REPORT-001",dsn=comp.report.table,aspice="SWE.5.BP3"]
// TestTableReporter_TraceabilityMatrix_Markdown verifies markdown format output.
func TestTableReporter_TraceabilityMatrix_Markdown(t *testing.T) {
	tr := NewTableReporter(buildReportGraph(), "markdown")
	out := tr.TraceabilityMatrix()
	if !strings.Contains(out, "REQ-001") {
		t.Error("expected REQ-001 in markdown output")
	}
	if !strings.Contains(out, "#") {
		t.Error("expected markdown headers (#) in output")
	}
}

// [test-spec,id=TS-RPT-003,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestTableReporter_TraceabilityMatrix_JSON verifies JSON format output.
func TestTableReporter_TraceabilityMatrix_JSON(t *testing.T) {
	tr := NewTableReporter(buildReportGraph(), "json")
	out := tr.TraceabilityMatrix()
	if !strings.Contains(out, "{") {
		t.Error("expected JSON output with braces")
	}
	if !strings.Contains(out, "REQ-001") {
		t.Error("expected REQ-001 in JSON output")
	}
}

// [test-spec,id=TS-RPT-004,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestTableReporter_GapReport_Text verifies text format gap report.
func TestTableReporter_GapReport_Text(t *testing.T) {
	tr := NewTableReporter(buildReportGraph(), "text")
	out := tr.GapReport()
	if !strings.Contains(out, "GAP") {
		t.Error("expected GAP in gap report header")
	}
}

// [test-spec,id=TS-RPT-005,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestTableReporter_GapReport_Markdown verifies markdown format gap report.
func TestTableReporter_GapReport_Markdown(t *testing.T) {
	tr := NewTableReporter(buildReportGraph(), "markdown")
	out := tr.GapReport()
	if out == "" {
		t.Error("expected non-empty gap report")
	}
}

// [test-spec,id=TS-RPT-006,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestTableReporter_GapReport_JSON verifies JSON format gap report.
func TestTableReporter_GapReport_JSON(t *testing.T) {
	tr := NewTableReporter(buildReportGraph(), "json")
	out := tr.GapReport()
	if !strings.Contains(out, "{") {
		t.Error("expected JSON in gap report")
	}
}

// [test-spec,id=TS-RPT-007,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestTableReporter_CoverageReport_Text verifies text format coverage report.
func TestTableReporter_CoverageReport_Text(t *testing.T) {
	tr := NewTableReporter(buildReportGraph(), "text")
	out := tr.CoverageReport()
	if !strings.Contains(out, "COVERAGE") {
		t.Error("expected COVERAGE in coverage report header")
	}
}

// [test-spec,id=TS-RPT-008,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestTableReporter_CoverageReport_Markdown verifies markdown format coverage report.
func TestTableReporter_CoverageReport_Markdown(t *testing.T) {
	tr := NewTableReporter(buildReportGraph(), "markdown")
	out := tr.CoverageReport()
	if out == "" {
		t.Error("expected non-empty coverage report")
	}
}

// [test-spec,id=TS-RPT-009,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestTableReporter_CoverageReport_JSON verifies JSON format coverage report.
func TestTableReporter_CoverageReport_JSON(t *testing.T) {
	tr := NewTableReporter(buildReportGraph(), "json")
	out := tr.CoverageReport()
	if !strings.Contains(out, "{") {
		t.Error("expected JSON in coverage report")
	}
}

// [test-spec,id=TS-RPT-010,req="REQ-REPORT-002",aspice="SWE.5.BP3"]
// TestTableReporter_EmptyGraph verifies reports work with empty graph.
func TestTableReporter_EmptyGraph(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    make(map[string]*model.TestSpec),
		TestCodes:    make(map[string]*model.TestCode),
		TestResults:  make(map[string]*model.TestResult),
		Links:        []*model.TraceLink{},
	}
	tr := NewTableReporter(graph.NewAnalyzer(g), "text")
	// Should not panic
	tr.TraceabilityMatrix()
	tr.GapReport()
	tr.CoverageReport()
}

// [test-spec,id=TS-RPT-011,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestTableReporter_DefaultFormat verifies default format (text) is used for unknown format.
func TestTableReporter_DefaultFormat(t *testing.T) {
	tr := NewTableReporter(buildReportGraph(), "unknown")
	out := tr.TraceabilityMatrix()
	if !strings.Contains(out, "TRACEABILITY MATRIX") {
		t.Error("expected default text format output")
	}
}

// [test-spec,id=TS-RPT-012,req="REQ-REPORT-002",aspice="SWE.5.BP3"]
// TestTableReporter_GapReport_HasOrphans verifies that orphan reqs appear in gap report.
func TestTableReporter_GapReport_HasOrphans(t *testing.T) {
	tr := NewTableReporter(buildReportGraph(), "text")
	out := tr.GapReport()
	// REQ-002 has no arch coverage
	if !strings.Contains(out, "REQ-002") {
		t.Error("expected REQ-002 (orphan) in gap report")
	}
}
