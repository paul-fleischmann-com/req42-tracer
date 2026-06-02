package report

import (
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

// [test-spec,id=TS-RPT-058,req="REQ-REPORT-001",dsn=comp.report.graph,aspice="SWE.4.BP2"]
// TestExportGraphData verifies ExportGraphData returns nodes and edges from a minimal graph.
func TestExportGraphData(t *testing.T) {
	// Create a minimal graph
	graph := &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-001": {
				ID:       "REQ-001",
				Title:    "Test Requirement",
				Priority: "high",
				Status:   "approved",
				ASPICE:   "SWE.1",
			},
		},
		ArchElements: map[string]*model.ArchElement{
			"comp.api": {
				ID:    "comp.api",
				Title: "API Component",
				ASPICE: "SWE.2",
				Req:    []string{"REQ-001"},
			},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"spec.api": {
				ID:    "spec.api",
				Title: "API Test Spec",
				Req:   []string{"REQ-001"},
				Arch:  []string{"comp.api"},
			},
		},
		TestResults: map[string]*model.TestResult{
			"result-1": {
				ID:       "result-1",
				TestName: "TestAPI",
				Status:   "passed",
			},
		},
		Links: []*model.TraceLink{
			{
				FromID:   "REQ-001",
				FromType: "requirement",
				ToID:     "comp.api",
				ToType:   "arch",
				LinkType: "satisfies",
			},
		},
	}

	// Export graph data
	data := ExportGraphData(graph)

	// Validate
	if len(data.Nodes) != 4 {
		t.Errorf("Expected 4 nodes, got %d", len(data.Nodes))
	}

	if len(data.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(data.Edges))
	}

	// Check node types
	nodeTypes := make(map[string]int)
	for _, node := range data.Nodes {
		nodeTypes[node.Type]++
	}

	if nodeTypes["requirement"] != 1 {
		t.Errorf("Expected 1 requirement node, got %d", nodeTypes["requirement"])
	}

	if nodeTypes["arch"] != 1 {
		t.Errorf("Expected 1 arch node, got %d", nodeTypes["arch"])
	}

	if nodeTypes["test-spec"] != 1 {
		t.Errorf("Expected 1 test-spec node, got %d", nodeTypes["test-spec"])
	}

	if nodeTypes["test-result"] != 1 {
		t.Errorf("Expected 1 test-result node, got %d", nodeTypes["test-result"])
	}

	// Check edge
	edge := data.Edges[0]
	if edge.Label != "satisfies" {
		t.Errorf("Expected edge label 'satisfies', got %s", edge.Label)
	}

	t.Log("✅ All graph export tests passed")
}

// [test-spec,id=TS-RPT-030,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestExportGraphData_DesignElements verifies DesignElement nodes are exported with type "dsn".
func TestExportGraphData_DesignElements(t *testing.T) {
	graph := &model.TraceabilityGraph{
		Requirements:   make(map[string]*model.Requirement),
		ArchElements:   make(map[string]*model.ArchElement),
		TestSpecs:      make(map[string]*model.TestSpec),
		TestCodes:      make(map[string]*model.TestCode),
		TestResults:    make(map[string]*model.TestResult),
		DesignElements: map[string]*model.DesignElement{
			"DSN-001": {
				ID:    "DSN-001",
				Title: "Parser Detail Design",
				ASPICE: "SWE.3",
				Arch:  "comp.parser",
			},
		},
		Links: []*model.TraceLink{},
	}

	data := ExportGraphData(graph)

	if len(data.Nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(data.Nodes))
	}
	node := data.Nodes[0]
	if node.Type != "dsn" {
		t.Errorf("Type = %q, want \"dsn\"", node.Type)
	}
	if node.Group != 4 {
		t.Errorf("Group = %d, want 4", node.Group)
	}
	if node.ID != "DSN-001" {
		t.Errorf("ID = %q, want DSN-001", node.ID)
	}
}

// [test-spec,id=TS-RPT-031,req="REQ-REPORT-001",aspice="SWE.5.BP3"]
// TestFilterGraphByType_DSN verifies that dsn nodes can be filtered by type.
func TestFilterGraphByType_DSN(t *testing.T) {
	data := &GraphData{
		Nodes: []Node{
			{ID: "arch-001", Type: "arch"},
			{ID: "DSN-001", Type: "dsn"},
		},
		Edges: []Edge{
			{Source: "arch-001", Target: "DSN-001", Label: "derives"},
		},
	}

	filtered := FilterGraphByType(data, "dsn")
	if len(filtered.Nodes) != 1 {
		t.Errorf("Expected 1 dsn node, got %d", len(filtered.Nodes))
	}
	if len(filtered.Edges) != 0 {
		t.Errorf("Expected 0 edges (arch filtered out), got %d", len(filtered.Edges))
	}

	filtered = FilterGraphByType(data, "arch", "dsn")
	if len(filtered.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(filtered.Nodes))
	}
	if len(filtered.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(filtered.Edges))
	}
}

// [test-spec,id=TS-RPT-059,req="REQ-REPORT-001",aspice="SWE.4.BP2"]
// TestFilterGraphByType verifies FilterGraphByType filters nodes and edges by type.
func TestFilterGraphByType(t *testing.T) {
	// Create graph data
	data := &GraphData{
		Nodes: []Node{
			{ID: "REQ-001", Type: "requirement"},
			{ID: "arch-001", Type: "arch"},
			{ID: "spec-001", Type: "test-spec"},
		},
		Edges: []Edge{
			{Source: "REQ-001", Target: "arch-001", Label: "satisfies"},
			{Source: "arch-001", Target: "spec-001", Label: "implements"},
		},
	}

	// Filter by requirement only
	filtered := FilterGraphByType(data, "requirement")

	if len(filtered.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(filtered.Nodes))
	}

	if len(filtered.Edges) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(filtered.Edges))
	}

	// Filter by requirement and arch
	filtered = FilterGraphByType(data, "requirement", "arch")

	if len(filtered.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(filtered.Nodes))
	}

	if len(filtered.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(filtered.Edges))
	}

	t.Log("✅ All filter tests passed")
}
