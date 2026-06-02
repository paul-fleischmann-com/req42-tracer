package lsp

import (
	"strings"
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

func diagGraph() *model.TraceabilityGraph {
	return &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-LSP-001": {ID: "REQ-LSP-001", Title: "LSP integration"},
		},
		ArchElements: map[string]*model.ArchElement{
			"comp.lsp": {ID: "comp.lsp", Title: "LSP Component"},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"TS-LSP-001": {ID: "TS-LSP-001", Title: "Initialize handshake"},
		},
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
	}
}

// [test-spec,id=TS-LSP-013,req="REQ-LSP-001",dsn=comp.lsp.diag,aspice=SWE.4-BP4]
// TestComputeDiagnostics_UnknownReq verifies unknown req= IDs produce an error diagnostic.
func TestComputeDiagnostics_UnknownReq(t *testing.T) {
	g := diagGraph()
	lines := []string{"[arch,id=comp.api,req=REQ-UNKNOWN]"}
	diags := computeDiagnostics("file:///test.adoc", lines, g)

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	d := diags[0]
	if d.Severity != diagError {
		t.Errorf("severity = %d, want %d (error)", d.Severity, diagError)
	}
	if !strings.Contains(d.Message, "REQ-UNKNOWN") {
		t.Errorf("message %q does not mention ID", d.Message)
	}
	if d.Range.Start.Line != 0 {
		t.Errorf("line = %d, want 0", d.Range.Start.Line)
	}
	// Value starts at index 22 in "[arch,id=comp.api,req=REQ-UNKNOWN]"
	if d.Range.Start.Character != 22 {
		t.Errorf("start character = %d, want 22", d.Range.Start.Character)
	}
	if d.Source != "req42-tracer" {
		t.Errorf("source = %q, want req42-tracer", d.Source)
	}
}

// [test-spec,id=TS-LSP-014,req="REQ-LSP-001",dsn=comp.lsp.diag,aspice="SWE.5.BP3"]
// TestComputeDiagnostics_KnownIDs verifies known IDs produce no diagnostics.
func TestComputeDiagnostics_KnownIDs(t *testing.T) {
	g := diagGraph()
	lines := []string{
		"[arch,id=comp.api,req=REQ-LSP-001]",
		"[test-spec,id=TS-001,arch=comp.lsp,test-spec=TS-LSP-001]",
	}
	diags := computeDiagnostics("file:///test.adoc", lines, g)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for known IDs, got %d: %+v", len(diags), diags)
	}
}

// [test-spec,id=TS-LSP-015,req="REQ-LSP-001",aspice="SWE.5.BP3"]
// TestComputeDiagnostics_MultiLine verifies diagnostics across multiple lines with correct line numbers.
func TestComputeDiagnostics_MultiLine(t *testing.T) {
	g := diagGraph()
	lines := []string{
		"[req,id=SWE-001,req=REQ-LSP-001]",   // line 0 — known, no diag
		"[arch,id=comp.api,req=REQ-BAD]",      // line 1 — unknown req
		"== Normal heading",                    // line 2 — no attributes
		"[arch,id=comp.x,arch=comp.UNKNOWN]",  // line 3 — unknown arch
	}
	diags := computeDiagnostics("file:///test.adoc", lines, g)
	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d: %+v", len(diags), diags)
	}
	if diags[0].Range.Start.Line != 1 {
		t.Errorf("diag[0] line = %d, want 1", diags[0].Range.Start.Line)
	}
	if diags[1].Range.Start.Line != 3 {
		t.Errorf("diag[1] line = %d, want 3", diags[1].Range.Start.Line)
	}
}
