package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/graph"
)

const sampleBausteinsicht = `{
	"$schema": "https://example.com/bausteinsicht.schema.json",
	"model": {
		"req42-tracer": {
			"description": "Requirements tracing tool",
			"elements": {
				"system": {
					"description": "Main system",
					"type": "system",
					"elements": {
						"parser": {
							"description": "AsciiDoc parser",
							"technology": "Go",
							"type": "component"
						}
					}
				}
			}
		}
	}
}`

const bausteinsichtWithComments = `{
	// This is a line comment
	"model": {
		"proj": {
			"description": "test", // inline comment
			"elements": {
				"comp": {
					"description": "a component"
				}
			}
		}
	}
}`

// [test-spec,id=TS-PARSE-021,req="REQ-PARSE-002",dsn=comp.parser.jsonc,aspice=SWE.4-BP4]
// TestBausteinsichtParser_Parse verifies that architecture.jsonc is parsed into arch elements.
func TestBausteinsichtParser_Parse(t *testing.T) {
	f := writeBausteinsicht(t, sampleBausteinsicht)
	p := NewBausteinsichtParser(f)
	g, err := p.Parse("proj")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(g.ArchElements) == 0 {
		t.Error("expected arch elements, got none")
	}
	// system and system.parser should be present
	if _, ok := g.ArchElements["system"]; !ok {
		t.Error("expected 'system' element")
	}
	if _, ok := g.ArchElements["system.parser"]; !ok {
		t.Error("expected 'system.parser' element")
	}
}

// [test-spec,id=TS-PARSE-022,req="REQ-PARSE-002",dsn=comp.parser.jsonc,aspice="SWE.5.BP3"]
// TestBausteinsichtParser_Attributes verifies that type and technology attributes are set.
func TestBausteinsichtParser_Attributes(t *testing.T) {
	f := writeBausteinsicht(t, sampleBausteinsicht)
	p := NewBausteinsichtParser(f)
	g, _ := p.Parse("proj")
	elem, ok := g.ArchElements["system.parser"]
	if !ok {
		t.Fatal("system.parser not found")
	}
	if elem.Attributes["type"] != "component" {
		t.Errorf("type = %q, want component", elem.Attributes["type"])
	}
	if elem.Attributes["technology"] != "Go" {
		t.Errorf("technology = %q, want Go", elem.Attributes["technology"])
	}
}

// [test-spec,id=TS-PARSE-023,req="REQ-PARSE-002",aspice="SWE.5.BP3"]
// TestBausteinsichtParser_FileNotFound verifies error for missing file.
func TestBausteinsichtParser_FileNotFound(t *testing.T) {
	p := NewBausteinsichtParser("/nonexistent/architecture.jsonc")
	_, err := p.Parse("proj")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// [test-spec,id=TS-PARSE-024,req="REQ-PARSE-002",aspice="SWE.5.BP3"]
// TestBausteinsichtParser_InvalidJSON verifies error for malformed JSON.
func TestBausteinsichtParser_InvalidJSON(t *testing.T) {
	f := writeBausteinsicht(t, `{"model": not valid json}`)
	p := NewBausteinsichtParser(f)
	_, err := p.Parse("proj")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// [test-spec,id=TS-PARSE-025,req="REQ-PARSE-002",aspice="SWE.5.BP3"]
// TestBausteinsichtParser_WithComments verifies that JSONC comments are stripped before parsing.
func TestBausteinsichtParser_WithComments(t *testing.T) {
	f := writeBausteinsicht(t, bausteinsichtWithComments)
	p := NewBausteinsichtParser(f)
	g, err := p.Parse("proj")
	if err != nil {
		t.Fatalf("Parse with comments error: %v", err)
	}
	if _, ok := g.ArchElements["comp"]; !ok {
		t.Error("expected 'comp' element after comment removal")
	}
}

// [test-spec,id=TS-PARSE-026,req="REQ-PARSE-002",aspice="SWE.5.BP3"]
// TestRemoveComments verifies that JSONC line and block comments are stripped
// and that URL-style // inside strings is preserved.
func TestRemoveComments(t *testing.T) {
	stripped := []struct {
		input   string
		noMatch string
	}{
		{`{"a": 1} // line comment`, "line comment"},
		{`{"a": /* block comment */ 1}`, "block comment"},
	}
	for _, tc := range stripped {
		got := removeComments(tc.input)
		if strings.Contains(got, tc.noMatch) {
			t.Errorf("removeComments(%q) still contains %q\ngot: %q", tc.input, tc.noMatch, got)
		}
	}

	// URLs inside strings must NOT be treated as comments
	urlInput := `{"url": "http://example.com"}`
	got := removeComments(urlInput)
	if !strings.Contains(got, "http://example.com") {
		t.Errorf("removeComments stripped URL inside string: %q", got)
	}
}

// [test-spec,id=TS-PARSE-027,req="REQ-PARSE-002",aspice="SWE.5.BP3"]
// TestBausteinsichtParser_Project verifies that parsed elements have the correct project.
func TestBausteinsichtParser_Project(t *testing.T) {
	f := writeBausteinsicht(t, sampleBausteinsicht)
	p := NewBausteinsichtParser(f)
	g, _ := p.Parse("myproject")
	for _, elem := range g.ArchElements {
		if elem.Project != "myproject" {
			t.Errorf("element %q has project %q, want myproject", elem.ID, elem.Project)
		}
	}
}

// [test-spec,id=TS-PARSE-028,req=REQ-PARSE-002,aspice=SWE.5-BP3]
// TestBausteinsichtParser_MergeIntoGraph verifies that parsed arch elements are
// successfully merged into a graph.Builder (command pipeline integration).
func TestBausteinsichtParser_MergeIntoGraph(t *testing.T) {
	f := writeBausteinsicht(t, sampleBausteinsicht)
	p := NewBausteinsichtParser(f)
	bGraph, err := p.Parse("software")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	builder := graph.NewBuilder()
	if err := builder.MergeGraph(bGraph); err != nil {
		t.Fatalf("MergeGraph error: %v", err)
	}

	g := builder.GetGraph()
	if len(g.ArchElements) == 0 {
		t.Error("expected arch elements in graph after merge, got none")
	}
	if _, ok := g.ArchElements["system"]; !ok {
		t.Error("expected 'system' element in merged graph")
	}
}

func writeBausteinsicht(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "architecture.jsonc")
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}
