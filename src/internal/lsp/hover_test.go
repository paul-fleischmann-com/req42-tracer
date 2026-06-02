package lsp

import (
	"strings"
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

func hoverGraph() *model.TraceabilityGraph {
	return &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-LSP-001": {
				ID:       "REQ-LSP-001",
				Title:    "LSP server integration",
				Text:     "The system shall provide req42-tracer lsp",
				Priority: "medium",
				Status:   "draft",
				ASPICE:   "SWE.3",
			},
		},
		ArchElements: map[string]*model.ArchElement{
			"comp.lsp": {
				ID:    "comp.lsp",
				Title: "LSP Server Component",
				Text:  "JSON-RPC 2.0 over stdio",
				Impl:  "internal/lsp",
			},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"TS-LSP-001": {
				ID:    "TS-LSP-001",
				Title: "Initialize handshake",
				Text:  "Verifies LSP initialize response",
			},
		},
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
	}
}

// [test-spec,id=TS-LSP-010,req="REQ-LSP-001",dsn=comp.lsp.hover,aspice=SWE.4-BP4]
// TestDetectHoverValue verifies cursor-position detection of attribute values.
func TestDetectHoverValue(t *testing.T) {
	cases := []struct {
		line      string
		col       int
		wantAttr  string
		wantValue string
		wantOk    bool
	}{
		// cursor inside req value
		{"[req,id=SWE-001,req=REQ-LSP-001]", 24, "req", "REQ-LSP-001", true},
		// cursor at start of req value
		{"[req,id=SWE-001,req=REQ-LSP-001]", 20, "req", "REQ-LSP-001", true},
		// cursor at last char of req value (inclusive)
		{"[req,id=SWE-001,req=REQ-LSP-001]", 30, "req", "REQ-LSP-001", true},
		// cursor one past end of value (exclusive — on ']')
		{"[req,id=SWE-001,req=REQ-LSP-001]", 31, "", "", false},
		// cursor on arch value
		{"[arch,id=comp.api,arch=comp.lsp]", 24, "arch", "comp.lsp", true},
		// cursor on test-spec value
		{"[test-spec,id=TS-001,test-spec=TS-LSP-001]", 33, "test-spec", "TS-LSP-001", true},
		// cursor outside any value (on bracket)
		{"[req,id=SWE-001,req=REQ-LSP-001]", 0, "", "", false},
		// cursor on key name, not value
		{"[req,id=SWE-001,req=REQ-LSP-001]", 17, "", "", false},
	}
	for _, c := range cases {
		attr, val, ok := detectHoverValue(c.line, c.col)
		if ok != c.wantOk {
			t.Errorf("detectHoverValue(%q, %d): ok=%v, want %v", c.line, c.col, ok, c.wantOk)
			continue
		}
		if ok {
			if attr != c.wantAttr {
				t.Errorf("detectHoverValue(%q, %d): attr=%q, want %q", c.line, c.col, attr, c.wantAttr)
			}
			if val != c.wantValue {
				t.Errorf("detectHoverValue(%q, %d): value=%q, want %q", c.line, c.col, val, c.wantValue)
			}
		}
	}
}

// [test-spec,id=TS-LSP-011,req="REQ-LSP-001",dsn=comp.lsp.hover,aspice="SWE.5.BP3"]
// TestBuildHoverContent_Req verifies hover content for a req= attribute value.
func TestBuildHoverContent_Req(t *testing.T) {
	g := hoverGraph()
	result := buildHoverContent("req", "REQ-LSP-001", g)
	if result == nil {
		t.Fatal("expected non-nil hover result for known requirement")
	}
	if result.Contents.Kind != "markdown" {
		t.Errorf("kind = %q, want markdown", result.Contents.Kind)
	}
	if result.Contents.Value == "" {
		t.Error("expected non-empty markdown content")
	}
	// Must contain ID and title
	for _, want := range []string{"REQ-LSP-001", "LSP server integration"} {
		if !strings.Contains(result.Contents.Value, want) {
			t.Errorf("hover content missing %q\ncontent: %s", want, result.Contents.Value)
		}
	}
}

// [test-spec,id=TS-LSP-012,req="REQ-LSP-001",aspice="SWE.5.BP3"]
// TestBuildHoverContent_Arch verifies hover content for an arch= attribute value.
func TestBuildHoverContent_Arch(t *testing.T) {
	g := hoverGraph()
	result := buildHoverContent("arch", "comp.lsp", g)
	if result == nil {
		t.Fatal("expected non-nil hover result for known arch element")
	}
	for _, want := range []string{"comp.lsp", "LSP Server Component"} {
		if !strings.Contains(result.Contents.Value, want) {
			t.Errorf("hover content missing %q\ncontent: %s", want, result.Contents.Value)
		}
	}
}

// [test-spec,id=TS-LSP-013,req="REQ-LSP-001",aspice="SWE.4.BP2"]
// TestBuildHoverContent_Unknown verifies nil is returned for unknown element IDs.
func TestBuildHoverContent_Unknown(t *testing.T) {
	g := hoverGraph()
	result := buildHoverContent("req", "REQ-UNKNOWN", g)
	if result != nil {
		t.Error("expected nil for unknown requirement")
	}
}

// [test-spec,id=TS-LSP-014,req="REQ-LSP-001",aspice="SWE.4.BP2"]
// TestBuildHoverContent_TestSpec verifies hover content is built for test-spec elements.
func TestBuildHoverContent_TestSpec(t *testing.T) {
	g := hoverGraph()
	result := buildHoverContent("test-spec", "TS-LSP-001", g)
	if result == nil {
		t.Fatal("expected non-nil hover result for known test spec")
	}
	for _, want := range []string{"TS-LSP-001", "Initialize handshake"} {
		if !strings.Contains(result.Contents.Value, want) {
			t.Errorf("hover content missing %q", want)
		}
	}
}
