package report

import (
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

// [test-spec,id=TS-RPT-043,req="REQ-REPORT-001",dsn=comp.report.matrix,aspice=SWE.4-BP4]
// TestBuildMatrixData verifies requirement rows and arch/spec columns are built correctly.
func TestBuildMatrixData(t *testing.T) {
	// Create a test graph
	graph := &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-001": {ID: "REQ-001", Title: "Test Req 1", Priority: "high", Status: "approved"},
			"REQ-002": {ID: "REQ-002", Title: "Test Req 2", Priority: "low", Status: "draft"},
		},
		ArchElements: map[string]*model.ArchElement{
			"arch-001": {ID: "arch-001", Title: "Arch 1"},
			"arch-002": {ID: "arch-002", Title: "Arch 2"},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"spec-001": {ID: "spec-001", Title: "Spec 1"},
		},
		TestResults:  make(map[string]*model.TestResult),
		TestCodes:    make(map[string]*model.TestCode),
		Links: []*model.TraceLink{
			{FromID: "REQ-001", FromType: "requirement", ToID: "arch-001", ToType: "arch", LinkType: "satisfies", Status: "active"},
			{FromID: "REQ-001", FromType: "requirement", ToID: "spec-001", ToType: "test-spec", LinkType: "verifies", Status: "active"},
		},
	}

	matrix := BuildMatrixData(graph)

	if len(matrix.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(matrix.Rows))
	}

	if len(matrix.Columns) != 3 {
		t.Errorf("Expected 3 columns (2 arch + 1 spec), got %d", len(matrix.Columns))
	}

	if matrix.Statistics.TotalRequirements != 2 {
		t.Errorf("Expected 2 total requirements, got %d", matrix.Statistics.TotalRequirements)
	}

	if matrix.Statistics.CoveredRequirements != 1 {
		t.Errorf("Expected 1 covered requirement, got %d", matrix.Statistics.CoveredRequirements)
	}

	t.Log("✅ Matrix data structure test passed")
}

// [test-spec,id=TS-RPT-044,req="REQ-REPORT-001",dsn=comp.report.matrix,aspice="SWE.5.BP3"]
// TestFilterMatrix verifies matrix rows are filtered by priority and status.
func TestFilterMatrix(t *testing.T) {
	graph := &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-001": {ID: "REQ-001", Title: "High Req", Priority: "high", Status: "approved"},
			"REQ-002": {ID: "REQ-002", Title: "Low Req", Priority: "low", Status: "draft"},
		},
		ArchElements: map[string]*model.ArchElement{
			"arch-001": {ID: "arch-001", Title: "Arch 1"},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:   make(map[string]*model.TestSpec),
		TestResults: make(map[string]*model.TestResult),
		TestCodes:   make(map[string]*model.TestCode),
		Links:       []*model.TraceLink{},
	}

	matrix := BuildMatrixData(graph)

	// Filter by high priority
	filtered := FilterMatrix(matrix, []string{"high"}, []string{"approved", "draft"})

	if len(filtered.Rows) != 1 {
		t.Errorf("Expected 1 filtered row, got %d", len(filtered.Rows))
	}

	if filtered.Rows[0].RequirementID != "REQ-001" {
		t.Errorf("Expected REQ-001, got %s", filtered.Rows[0].RequirementID)
	}

	t.Log("✅ Matrix filtering test passed")
}

// [test-spec,id=TS-RPT-045,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestExportMatrixCSV verifies matrix is exported as valid CSV with header and data rows.
func TestExportMatrixCSV(t *testing.T) {
	graph := &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-001": {ID: "REQ-001", Title: "Test", Priority: "high", Status: "approved"},
		},
		ArchElements: map[string]*model.ArchElement{
			"arch-001": {ID: "arch-001", Title: "Arch"},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:   make(map[string]*model.TestSpec),
		TestResults: make(map[string]*model.TestResult),
		TestCodes:   make(map[string]*model.TestCode),
		Links: []*model.TraceLink{
			{FromID: "REQ-001", FromType: "requirement", ToID: "arch-001", ToType: "arch", LinkType: "satisfies"},
		},
	}

	matrix := BuildMatrixData(graph)
	csv := ExportMatrixToCSV(matrix)

	if !Contains(csv, "Requirement,Priority,Status") {
		t.Error("CSV header not found")
	}

	if !Contains(csv, "REQ-001,high,approved") {
		t.Error("CSV requirement row not found")
	}

	t.Log("✅ CSV export test passed")
}

// [test-spec,id=TS-RPT-034,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestBuildMatrixData_DSNColumns verifies that DesignElement columns appear and get covered via arch chain.
func TestBuildMatrixData_DSNColumns(t *testing.T) {
	graph := &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-001": {ID: "REQ-001", Title: "Req", Priority: "high", Status: "approved"},
		},
		ArchElements: map[string]*model.ArchElement{
			"arch-001": {ID: "arch-001", Title: "Arch"},
		},
		DesignElements: map[string]*model.DesignElement{
			"DSN-001": {ID: "DSN-001", Title: "Design", Arch: "arch-001"},
		},
		TestSpecs:   make(map[string]*model.TestSpec),
		TestResults: make(map[string]*model.TestResult),
		TestCodes:   make(map[string]*model.TestCode),
		Links: []*model.TraceLink{
			{FromID: "REQ-001", FromType: "requirement", ToID: "arch-001", ToType: "arch", LinkType: "satisfies", Status: "active"},
		},
	}

	data := BuildMatrixData(graph)

	// Should have arch + dsn columns
	colTypes := make(map[string]bool)
	for _, col := range data.Columns {
		colTypes[col.Type] = true
	}
	if !colTypes["dsn"] {
		t.Error("Expected dsn column type, not found")
	}

	// DSN-001 cell for REQ-001 should be covered (via arch-001)
	row := data.Rows[0]
	cell, ok := row.Cells["DSN-001"]
	if !ok {
		t.Fatal("DSN-001 cell not found in row")
	}
	if cell.Status != "covered" {
		t.Errorf("DSN-001 cell status = %q, want covered", cell.Status)
	}
}

// [test-spec,id=TS-RPT-035,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestExportGraphData_TestResultDetails verifies error and stdout fields in TestResult metadata.
func TestExportGraphData_TestResultDetails(t *testing.T) {
	graph := &model.TraceabilityGraph{
		Requirements:   make(map[string]*model.Requirement),
		ArchElements:   make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:      make(map[string]*model.TestSpec),
		TestCodes:      make(map[string]*model.TestCode),
		TestResults: map[string]*model.TestResult{
			"res-001": {
				ID:       "res-001",
				TestName: "TestFoo",
				Status:   "failed",
				Error:    "assertion failed: want 1 got 2",
				Stdout:   "=== RUN TestFoo\n--- FAIL: TestFoo",
				Duration: 0.42,
				Platform: "linux",
			},
		},
		Links: []*model.TraceLink{},
	}

	data := ExportGraphData(graph)
	if len(data.Nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(data.Nodes))
	}
	meta := data.Nodes[0].Metadata
	if meta["error"] != "assertion failed: want 1 got 2" {
		t.Errorf("error metadata = %v, want assertion message", meta["error"])
	}
	if meta["stdout"] == "" {
		t.Error("stdout metadata should not be empty")
	}
}

func Contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
