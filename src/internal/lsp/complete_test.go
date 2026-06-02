package lsp

import (
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

func testGraph() *model.TraceabilityGraph {
	return &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-PARSE-001": {ID: "REQ-PARSE-001", Title: "Parse AsciiDoc blocks"},
			"REQ-GRAPH-001": {ID: "REQ-GRAPH-001", Title: "Build traceability graph"},
			"REQ-LSP-001":   {ID: "REQ-LSP-001", Title: "LSP server integration"},
		},
		ArchElements: map[string]*model.ArchElement{
			"comp.parser": {ID: "comp.parser", Title: "Parser Component"},
			"comp.graph":  {ID: "comp.graph", Title: "Graph Component"},
			"comp.lsp":    {ID: "comp.lsp", Title: "LSP Component"},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"TS-LSP-001":   {ID: "TS-LSP-001", Title: "Initialize handshake"},
			"TS-PARSE-001": {ID: "TS-PARSE-001", Title: "Parse req block"},
		},
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
	}
}

// [test-spec,id=TS-LSP-004,req="REQ-LSP-001",dsn=comp.lsp.complete,aspice=SWE.4-BP4]
// TestDetectContext verifies attribute-context detection from line prefixes.
func TestDetectContext(t *testing.T) {
	cases := []struct {
		line    string
		wantCtx completionContext
		wantPfx string
	}{
		{"[req,id=SWE-001,req=", ctxReq, ""},
		{"[req,id=SWE-001,req=REQ-P", ctxReq, "REQ-P"},
		{"[arch,id=comp.api,arch=", ctxArch, ""},
		{"[arch,id=comp.api,arch=comp.", ctxArch, "comp."},
		{"[test-spec,id=TS-001,test-spec=", ctxTestSpec, ""},
		{"[test-spec,id=TS-001,test-spec=TS-L", ctxTestSpec, "TS-L"},
		{"== Normal heading", ctxNone, ""},
		{"[req,id=SWE-001]", ctxNone, ""},
	}
	for _, c := range cases {
		ctx, pfx := detectContext(c.line)
		if ctx != c.wantCtx {
			t.Errorf("detectContext(%q): ctx=%d, want %d", c.line, ctx, c.wantCtx)
		}
		if pfx != c.wantPfx {
			t.Errorf("detectContext(%q): prefix=%q, want %q", c.line, pfx, c.wantPfx)
		}
	}
}

// [test-spec,id=TS-LSP-005,req="REQ-LSP-001",dsn=comp.lsp.complete,aspice="SWE.5.BP3"]
// TestBuildCompletions_Req verifies req= completions return all requirement IDs.
func TestBuildCompletions_Req(t *testing.T) {
	g := testGraph()
	list := buildCompletions("[req,id=SWE-001,req=", g)
	if list.IsIncomplete {
		t.Error("expected IsIncomplete=false")
	}
	if len(list.Items) != 3 {
		t.Errorf("got %d items, want 3", len(list.Items))
	}
	// Items must be sorted
	if list.Items[0].Label != "REQ-GRAPH-001" {
		t.Errorf("first item = %q, want REQ-GRAPH-001", list.Items[0].Label)
	}
}

// [test-spec,id=TS-LSP-006,req="REQ-LSP-001",aspice="SWE.5.BP3"]
// TestBuildCompletions_Prefix verifies that prefix filtering works correctly.
func TestBuildCompletions_Prefix(t *testing.T) {
	g := testGraph()
	list := buildCompletions("[arch,id=comp.api,req=REQ-P", g)
	if len(list.Items) != 1 {
		t.Errorf("got %d items, want 1 (REQ-PARSE-001)", len(list.Items))
	}
	if len(list.Items) > 0 && list.Items[0].Label != "REQ-PARSE-001" {
		t.Errorf("item = %q, want REQ-PARSE-001", list.Items[0].Label)
	}
}

// [test-spec,id=TS-LSP-007,req="REQ-LSP-001",aspice="SWE.5.BP3"]
// TestBuildCompletions_Arch verifies arch= completions return architecture IDs.
func TestBuildCompletions_Arch(t *testing.T) {
	g := testGraph()
	list := buildCompletions("[test-spec,id=TS-001,arch=comp.", g)
	if len(list.Items) != 3 {
		t.Errorf("got %d items, want 3", len(list.Items))
	}
}

// [test-spec,id=TS-LSP-008,req="REQ-LSP-001",aspice="SWE.5.BP3"]
// TestBuildCompletions_TestSpec verifies test-spec= completions return spec IDs.
func TestBuildCompletions_TestSpec(t *testing.T) {
	g := testGraph()
	list := buildCompletions("[arch,id=comp.api,test-spec=TS-L", g)
	if len(list.Items) != 1 {
		t.Errorf("got %d items, want 1 (TS-LSP-001)", len(list.Items))
	}
}

// [test-spec,id=TS-LSP-009,req="REQ-LSP-001",aspice="SWE.5.BP3"]
// TestBuildCompletions_NoContext verifies that non-attribute lines return empty list.
func TestBuildCompletions_NoContext(t *testing.T) {
	g := testGraph()
	list := buildCompletions("== Normal AsciiDoc heading", g)
	if len(list.Items) != 0 {
		t.Errorf("expected no completions outside attribute context, got %d", len(list.Items))
	}
}
