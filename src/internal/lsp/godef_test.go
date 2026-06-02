package lsp

import (
	"strings"
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

func godefGraph() *model.TraceabilityGraph {
	return &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-LSP-001": {
				ID:         "REQ-LSP-001",
				FilePath:   "docs/requirements/req42.adoc",
				LineNumber: 42,
			},
		},
		ArchElements: map[string]*model.ArchElement{
			"comp.lsp": {
				ID:         "comp.lsp",
				FilePath:   "docs/arc42/arc42.adoc",
				LineNumber: 110,
			},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"TS-LSP-001": {
				ID:         "TS-LSP-001",
				FilePath:   "internal/lsp/server_test.go",
				LineNumber: 15,
			},
		},
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
	}
}

// [test-spec,id=TS-LSP-016,req="REQ-LSP-001",dsn=comp.lsp.godef,aspice=SWE.4-BP4]
// TestFindDefinition_Req verifies go-to-definition for a known requirement ID.
func TestFindDefinition_Req(t *testing.T) {
	g := godefGraph()
	loc := findDefinition("req", "REQ-LSP-001", g)
	if loc == nil {
		t.Fatal("expected non-nil location for known requirement")
	}
	if !strings.HasSuffix(loc.URI, "req42.adoc") {
		t.Errorf("URI %q does not end with req42.adoc", loc.URI)
	}
	if !strings.HasPrefix(loc.URI, "file://") {
		t.Errorf("URI %q does not start with file://", loc.URI)
	}
	// LineNumber 42 → 0-based line 41
	if loc.Range.Start.Line != 41 {
		t.Errorf("line = %d, want 41 (0-based from LineNumber=42)", loc.Range.Start.Line)
	}
}

// [test-spec,id=TS-LSP-017,req="REQ-LSP-001",dsn=comp.lsp.godef,aspice="SWE.5.BP3"]
// TestFindDefinition_Arch verifies go-to-definition for a known arch element.
func TestFindDefinition_Arch(t *testing.T) {
	g := godefGraph()
	loc := findDefinition("arch", "comp.lsp", g)
	if loc == nil {
		t.Fatal("expected non-nil location for known arch element")
	}
	if !strings.HasSuffix(loc.URI, "arc42.adoc") {
		t.Errorf("URI %q does not end with arc42.adoc", loc.URI)
	}
	if loc.Range.Start.Line != 109 { // 110 → 0-based = 109
		t.Errorf("line = %d, want 109", loc.Range.Start.Line)
	}
}

// [test-spec,id=TS-LSP-018,req="REQ-LSP-001",aspice="SWE.5.BP3"]
// TestFindDefinition_Unknown verifies nil is returned for unknown IDs.
func TestFindDefinition_Unknown(t *testing.T) {
	g := godefGraph()
	if loc := findDefinition("req", "REQ-UNKNOWN", g); loc != nil {
		t.Errorf("expected nil for unknown ID, got %+v", loc)
	}
	if loc := findDefinition("arch", "comp.unknown", g); loc != nil {
		t.Errorf("expected nil for unknown arch, got %+v", loc)
	}
	if loc := findDefinition("test-spec", "TS-UNKNOWN", g); loc != nil {
		t.Errorf("expected nil for unknown test-spec, got %+v", loc)
	}
}

// [test-spec,id=TS-LSP-019,req="REQ-LSP-001",aspice="SWE.5.BP3"]
// TestFindDefinition_TestSpec verifies go-to-definition for a known test-spec ID.
func TestFindDefinition_TestSpec(t *testing.T) {
	g := godefGraph()
	loc := findDefinition("test-spec", "TS-LSP-001", g)
	if loc == nil {
		t.Fatal("expected location for known test-spec")
	}
	if !strings.HasSuffix(loc.URI, "server_test.go") {
		t.Errorf("URI %q does not point to server_test.go", loc.URI)
	}
	if loc.Range.Start.Line != 14 { // 15 → 0-based = 14
		t.Errorf("line = %d, want 14", loc.Range.Start.Line)
	}
}
