package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// [test-spec,id=TS-PARSE-001,req="REQ-PARSE-001",dsn=comp.parser.adoc,aspice=SWE.4-BP4]
// TestParseRequirements verifies that [req] blocks are correctly parsed from AsciiDoc.
func TestParseRequirements(t *testing.T) {
	content := `= Test Doc

[req,id=REQ-001,version=1,priority=high,aspice=SWE.1,status=approved,reviewed-by=alice,reviewed-date=2026-01-15,derives=SYS-001]
== My Requirement

Description text.
`
	f := writeTemp(t, content)
	p := NewADocParser(f)
	reqs, err := p.ParseRequirements("testproject")
	if err != nil {
		t.Fatalf("ParseRequirements error: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	req := reqs[0]
	if req.ID != "REQ-001" {
		t.Errorf("ID = %q, want REQ-001", req.ID)
	}
	if req.Version != 1 {
		t.Errorf("Version = %d, want 1", req.Version)
	}
	if req.Priority != "high" {
		t.Errorf("Priority = %q, want high", req.Priority)
	}
	if req.ASPICE != "SWE.1" {
		t.Errorf("ASPICE = %q, want SWE.1", req.ASPICE)
	}
	if req.Status != "approved" {
		t.Errorf("Status = %q, want approved", req.Status)
	}
	if req.ReviewedBy != "alice" {
		t.Errorf("ReviewedBy = %q, want alice", req.ReviewedBy)
	}
	if len(req.Derives) != 1 || req.Derives[0] != "SYS-001" {
		t.Errorf("Derives = %v, want [SYS-001]", req.Derives)
	}
	if req.Title != "My Requirement" {
		t.Errorf("Title = %q, want 'My Requirement'", req.Title)
	}
	if req.Project != "testproject" {
		t.Errorf("Project = %q, want testproject", req.Project)
	}
}

// [test-spec,id=TS-PARSE-002,req="REQ-PARSE-001",dsn=comp.parser.adoc,aspice="SWE.5.BP3"]
// TestParseRequirements_MissingID verifies that blocks without id are skipped.
func TestParseRequirements_MissingID(t *testing.T) {
	content := `[req,version=1]
== No ID req
`
	f := writeTemp(t, content)
	p := NewADocParser(f)
	reqs, err := p.ParseRequirements("proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reqs) != 0 {
		t.Errorf("expected 0 requirements, got %d", len(reqs))
	}
}

// [test-spec,id=TS-PARSE-003,req="REQ-PARSE-001",aspice="SWE.5.BP3"]
// TestParseRequirements_Defaults verifies default values for priority and status.
func TestParseRequirements_Defaults(t *testing.T) {
	content := `[req,id=REQ-002]
== Minimal Req
`
	f := writeTemp(t, content)
	p := NewADocParser(f)
	reqs, _ := p.ParseRequirements("proj")
	if len(reqs) != 1 {
		t.Fatalf("expected 1 req, got %d", len(reqs))
	}
	if reqs[0].Priority != "medium" {
		t.Errorf("default Priority = %q, want medium", reqs[0].Priority)
	}
	if reqs[0].Status != "draft" {
		t.Errorf("default Status = %q, want draft", reqs[0].Status)
	}
}

// [test-spec,id=TS-PARSE-004,req="REQ-PARSE-002",aspice="SWE.5.BP3"]
// TestParseArchElements verifies that [arch] blocks are correctly parsed.
func TestParseArchElements(t *testing.T) {
	content := `[arch,id=comp.parser,parent=comp.system,aspice=SWE.2,req=REQ-001,impl=internal/parser/adoc.go]
== Parser Component

Description.
`
	f := writeTemp(t, content)
	p := NewADocParser(f)
	archs, err := p.ParseArchElements("proj")
	if err != nil {
		t.Fatalf("ParseArchElements error: %v", err)
	}
	if len(archs) != 1 {
		t.Fatalf("expected 1 arch, got %d", len(archs))
	}
	a := archs[0]
	if a.ID != "comp.parser" {
		t.Errorf("ID = %q", a.ID)
	}
	if a.Parent != "comp.system" {
		t.Errorf("Parent = %q", a.Parent)
	}
	if a.ASPICE != "SWE.2" {
		t.Errorf("ASPICE = %q", a.ASPICE)
	}
	if len(a.Req) != 1 || a.Req[0] != "REQ-001" {
		t.Errorf("Req = %v, want [REQ-001]", a.Req)
	}
	if a.Impl != "internal/parser/adoc.go" {
		t.Errorf("Impl = %q", a.Impl)
	}
	if a.Title != "Parser Component" {
		t.Errorf("Title = %q", a.Title)
	}
}

// [test-spec,id=TS-PARSE-005,req="REQ-PARSE-002",aspice="SWE.5.BP3"]
// TestParseArchElements_MissingID verifies that arch blocks without id are skipped.
func TestParseArchElements_MissingID(t *testing.T) {
	content := `[arch,parent=comp.system]
== No ID
`
	f := writeTemp(t, content)
	p := NewADocParser(f)
	archs, _ := p.ParseArchElements("proj")
	if len(archs) != 0 {
		t.Errorf("expected 0 archs, got %d", len(archs))
	}
}

// [test-spec,id=TS-PARSE-006,req="REQ-PARSE-002",aspice="SWE.5.BP3"]
// TestParseArchElements_WithTestSpec verifies test-spec attribute on arch blocks.
func TestParseArchElements_WithTestSpec(t *testing.T) {
	content := `[arch,id=comp.x,test-spec=TS-001]
=== X Component
`
	f := writeTemp(t, content)
	p := NewADocParser(f)
	archs, _ := p.ParseArchElements("proj")
	if len(archs) != 1 {
		t.Fatalf("expected 1 arch, got %d", len(archs))
	}
	if archs[0].TestSpec != "TS-001" {
		t.Errorf("TestSpec = %q, want TS-001", archs[0].TestSpec)
	}
}

// [test-spec,id=TS-PARSE-007,req="REQ-PARSE-003",aspice="SWE.5.BP3"]
// TestParseTestSpecs verifies that [test-spec] blocks are correctly parsed.
func TestParseTestSpecs(t *testing.T) {
	content := `[test-spec,id=TS-001,req=REQ-001,arch=comp.parser]
== My Test Spec
`
	f := writeTemp(t, content)
	p := NewADocParser(f)
	specs, err := p.ParseTestSpecs("proj")
	if err != nil {
		t.Fatalf("ParseTestSpecs error: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}
	s := specs[0]
	if s.ID != "TS-001" {
		t.Errorf("ID = %q", s.ID)
	}
	if len(s.Req) != 1 || s.Req[0] != "REQ-001" {
		t.Errorf("Req = %v", s.Req)
	}
	if len(s.Arch) != 1 || s.Arch[0] != "comp.parser" {
		t.Errorf("Arch = %v", s.Arch)
	}
	if s.Title != "My Test Spec" {
		t.Errorf("Title = %q", s.Title)
	}
}

// [test-spec,id=TS-PARSE-008,req="REQ-PARSE-003",aspice="SWE.5.BP3"]
// TestParseTestSpecs_MissingID verifies that test-spec blocks without id are skipped.
func TestParseTestSpecs_MissingID(t *testing.T) {
	content := `[test-spec,req=REQ-001]
== No ID
`
	f := writeTemp(t, content)
	p := NewADocParser(f)
	specs, _ := p.ParseTestSpecs("proj")
	if len(specs) != 0 {
		t.Errorf("expected 0 specs, got %d", len(specs))
	}
}

// [test-spec,id=TS-PARSE-009,req="REQ-PARSE-001",aspice="SWE.5.BP3"]
// TestParseRequirements_FileNotFound verifies that missing file returns error.
func TestParseRequirements_FileNotFound(t *testing.T) {
	p := NewADocParser("/nonexistent/file.adoc")
	_, err := p.ParseRequirements("proj")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// [test-spec,id=TS-PARSE-010,req="REQ-PARSE-002",aspice="SWE.5.BP3"]
// TestParseArchElements_FileNotFound verifies error on missing file.
func TestParseArchElements_FileNotFound(t *testing.T) {
	p := NewADocParser("/nonexistent/file.adoc")
	_, err := p.ParseArchElements("proj")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// [test-spec,id=TS-PARSE-011,req="REQ-PARSE-003",aspice="SWE.5.BP3"]
// TestParseTestSpecs_FileNotFound verifies error on missing file.
func TestParseTestSpecs_FileNotFound(t *testing.T) {
	p := NewADocParser("/nonexistent/file.adoc")
	_, err := p.ParseTestSpecs("proj")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// [test-spec,id=TS-PARSE-012,req="REQ-PARSE-001",aspice="SWE.5.BP3"]
// TestExtractAttributes verifies correct attribute extraction from block headers.
func TestExtractAttributes(t *testing.T) {
	cases := []struct {
		input    string
		wantType string
		wantID   string
	}{
		{"[req,id=REQ-001,priority=high]", "req", "REQ-001"},
		{"[arch,id=comp.api,parent=comp.system]", "arch", "comp.api"},
		{"[test-spec,id=TS-001]", "test-spec", "TS-001"},
		{"no brackets", "", ""},
		{"[req]", "req", ""},
	}
	for _, tc := range cases {
		attrs := extractAttributes(tc.input)
		if attrs["type"] != tc.wantType {
			t.Errorf("extractAttributes(%q) type = %q, want %q", tc.input, attrs["type"], tc.wantType)
		}
		if attrs["id"] != tc.wantID {
			t.Errorf("extractAttributes(%q) id = %q, want %q", tc.input, attrs["id"], tc.wantID)
		}
	}
}

// [test-spec,id=TS-PARSE-013,req="REQ-PARSE-001",aspice="SWE.5.BP3"]
// TestExtractAttributes_Quoted verifies that quoted values are unquoted correctly.
func TestExtractAttributes_Quoted(t *testing.T) {
	attrs := extractAttributes(`[req,id="REQ-001",title='My Title']`)
	if attrs["id"] != "REQ-001" {
		t.Errorf("quoted id = %q, want REQ-001", attrs["id"])
	}
}

// [test-spec,id=TS-PARSE-014,req="REQ-PARSE-001",aspice="SWE.5.BP3"]
// TestParseIDList verifies that comma-separated IDs are correctly split.
func TestParseIDList(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"REQ-001,REQ-002", []string{"REQ-001", "REQ-002"}},
		{"REQ-001", []string{"REQ-001"}},
		{" REQ-001 , REQ-002 ", []string{"REQ-001", "REQ-002"}},
		{"", nil},
	}
	for _, tc := range cases {
		got := parseIDList(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("parseIDList(%q) = %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("parseIDList(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

// [test-spec,id=TS-PARSE-015,req="REQ-PARSE-001",aspice="SWE.5.BP3"]
// TestParseAllFromDir verifies that all .adoc files in a directory are parsed.
func TestParseAllFromDir(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "req.adoc"), `[req,id=REQ-D-001]
== Dir Requirement
`)
	writeFile(t, filepath.Join(dir, "arch.adoc"), `[arch,id=comp.dir]
== Dir Component
`)

	g, err := ParseAllFromDir(dir, "proj")
	if err != nil {
		t.Fatalf("ParseAllFromDir error: %v", err)
	}
	if len(g.Requirements) != 1 {
		t.Errorf("Requirements = %d, want 1", len(g.Requirements))
	}
	if len(g.ArchElements) != 1 {
		t.Errorf("ArchElements = %d, want 1", len(g.ArchElements))
	}
}

// [test-spec,id=TS-PARSE-016,req="REQ-PARSE-001",aspice="SWE.5.BP3"]
// TestParseAllFromDir_EmptyDir verifies that empty directory returns empty graph.
func TestParseAllFromDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	g, err := ParseAllFromDir(dir, "proj")
	if err != nil {
		t.Fatalf("ParseAllFromDir error: %v", err)
	}
	if len(g.Requirements) != 0 || len(g.ArchElements) != 0 {
		t.Error("expected empty graph for empty dir")
	}
}

// [test-spec,id=TS-PARSE-017,req="REQ-PARSE-001",aspice="SWE.5.BP3"]
// TestParseAllFromDir_NonExistentDir verifies error for missing directory.
func TestParseAllFromDir_NonExistentDir(t *testing.T) {
	_, err := ParseAllFromDir("/nonexistent/path", "proj")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// [test-spec,id=TS-PARSE-018,req="REQ-PARSE-001",aspice="SWE.5.BP3"]
// TestParseMultipleRequirements verifies parsing multiple req blocks from one file.
func TestParseMultipleRequirements(t *testing.T) {
	content := `[req,id=REQ-A]
== Req A

[req,id=REQ-B,priority=low]
== Req B
`
	f := writeTemp(t, content)
	p := NewADocParser(f)
	reqs, _ := p.ParseRequirements("proj")
	if len(reqs) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(reqs))
	}
}

// [test-spec,id=TS-PARSE-019,req="REQ-PARSE-001",aspice="SWE.5.BP3"]
// TestParseRequirements_MultiLineBlock verifies multi-line block attribute parsing.
func TestParseRequirements_MultiLineBlock(t *testing.T) {
	content := `[req,id=REQ-ML-001,version=2,
 priority=high,aspice=SWE.1]
== Multi-line Req
`
	f := writeTemp(t, content)
	p := NewADocParser(f)
	reqs, _ := p.ParseRequirements("proj")
	if len(reqs) != 1 {
		t.Fatalf("expected 1 req from multi-line block, got %d", len(reqs))
	}
	if reqs[0].ID != "REQ-ML-001" {
		t.Errorf("ID = %q, want REQ-ML-001", reqs[0].ID)
	}
}

// [test-spec,id=TS-PARSE-020,req="REQ-PARSE-001",aspice="SWE.5.BP3"]
// TestParseRequirements_ReviewedDate verifies reviewed-date parsing.
func TestParseRequirements_ReviewedDate(t *testing.T) {
	content := `[req,id=REQ-DATE,reviewed-date=2026-03-15]
== Date Req
`
	f := writeTemp(t, content)
	p := NewADocParser(f)
	reqs, _ := p.ParseRequirements("proj")
	if len(reqs) != 1 {
		t.Fatalf("expected 1 req, got %d", len(reqs))
	}
	if reqs[0].ReviewedDate.IsZero() {
		t.Error("ReviewedDate should be parsed")
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.adoc")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// [test-spec,id=TS-PARSE-030,req=REQ-PARSE-002,aspice=SWE.4-BP3]
// TestParseDsnBlock verifies that [dsn] blocks are parsed into DesignElements.
func TestParseDsnBlock(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "design.adoc")
	content := `[dsn,id=comp.parser.tokenizer,arch=comp.parser,aspice=SWE.3,impl=src/parser/tokenizer.go]
== Tokenizer Unit
`
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewADocParser(f)
	elems, err := p.ParseDesignElements("proj")
	if err != nil {
		t.Fatalf("ParseDesignElements: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 design element, got %d", len(elems))
	}
	e := elems[0]
	if e.ID != "comp.parser.tokenizer" {
		t.Errorf("ID = %q", e.ID)
	}
	if e.Arch != "comp.parser" {
		t.Errorf("Arch = %q", e.Arch)
	}
	if e.ASPICE != "SWE.3" {
		t.Errorf("ASPICE = %q", e.ASPICE)
	}
	if e.Impl != "src/parser/tokenizer.go" {
		t.Errorf("Impl = %q", e.Impl)
	}
	if e.Title != "Tokenizer Unit" {
		t.Errorf("Title = %q", e.Title)
	}
}

// [test-spec,id=TS-PARSE-031,req=REQ-PARSE-002,aspice=SWE.4-BP3]
// TestParseTestSpec_DsnAttr verifies that dsn= on [test-spec] is parsed into TestSpec.Dsn.
func TestParseTestSpec_DsnAttr(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "tests.adoc")
	content := `[test-spec,id=TS-UNIT-001,dsn=comp.parser.tokenizer,aspice=SWE.4]
== Unit Test: Tokenizer
`
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewADocParser(f)
	specs, err := p.ParseTestSpecs("proj")
	if err != nil {
		t.Fatalf("ParseTestSpecs: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 test-spec, got %d", len(specs))
	}
	if len(specs[0].Dsn) != 1 || specs[0].Dsn[0] != "comp.parser.tokenizer" {
		t.Errorf("Dsn = %v", specs[0].Dsn)
	}
}
