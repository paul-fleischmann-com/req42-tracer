package graph

import (
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

func newTestGraph() *model.TraceabilityGraph {
	return &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-001": {ID: "REQ-001"},
		},
		ArchElements: map[string]*model.ArchElement{
			"comp.api": {ID: "comp.api", Parent: "comp.system", Req: []string{"REQ-001"}},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"TS-001": {ID: "TS-001", Req: []string{"REQ-001"}, Arch: []string{"comp.api"}},
		},
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
		Links:       []*model.TraceLink{},
	}
}

// [test-spec,id=TS-GRAPH-001,req="REQ-GRAPH-001",dsn=comp.graph.builder,aspice=SWE.4-BP4]
// TestNewBuilder verifies that NewBuilder initializes an empty graph.
func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	g := b.GetGraph()
	if len(g.Requirements) != 0 || len(g.ArchElements) != 0 || len(g.Links) != 0 {
		t.Error("expected empty graph from NewBuilder")
	}
}

// [test-spec,id=TS-GRAPH-002,req="REQ-GRAPH-001",dsn=comp.graph.builder,aspice="SWE.5.BP3"]
// TestMergeGraph verifies that graphs are correctly merged.
func TestMergeGraph(t *testing.T) {
	b := NewBuilder()
	if err := b.MergeGraph(newTestGraph()); err != nil {
		t.Fatalf("MergeGraph error: %v", err)
	}
	g := b.GetGraph()
	if len(g.Requirements) != 1 {
		t.Errorf("Requirements = %d, want 1", len(g.Requirements))
	}
	if len(g.ArchElements) != 1 {
		t.Errorf("ArchElements = %d, want 1", len(g.ArchElements))
	}
	if len(g.TestSpecs) != 1 {
		t.Errorf("TestSpecs = %d, want 1", len(g.TestSpecs))
	}
}

// [test-spec,id=TS-GRAPH-003,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestMergeGraph_NilGraph verifies that merging nil graph is a no-op.
func TestMergeGraph_NilGraph(t *testing.T) {
	b := NewBuilder()
	if err := b.MergeGraph(nil); err != nil {
		t.Errorf("expected nil error for nil graph, got %v", err)
	}
}

// [test-spec,id=TS-GRAPH-004,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestMergeGraph_DuplicateReq verifies that duplicate req IDs return error.
func TestMergeGraph_DuplicateReq(t *testing.T) {
	b := NewBuilder()
	b.MergeGraph(newTestGraph())
	err := b.MergeGraph(newTestGraph())
	if err == nil {
		t.Error("expected duplicate error for requirement, got nil")
	}
}

// [test-spec,id=TS-GRAPH-005,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestMergeGraph_DuplicateArch verifies that duplicate arch IDs return error.
func TestMergeGraph_DuplicateArch(t *testing.T) {
	b := NewBuilder()
	g1 := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: map[string]*model.ArchElement{"comp.x": {ID: "comp.x"}},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    make(map[string]*model.TestSpec),
		TestCodes:    make(map[string]*model.TestCode),
		TestResults:  make(map[string]*model.TestResult),
		Links:        []*model.TraceLink{},
	}
	g2 := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: map[string]*model.ArchElement{"comp.x": {ID: "comp.x"}},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    make(map[string]*model.TestSpec),
		TestCodes:    make(map[string]*model.TestCode),
		TestResults:  make(map[string]*model.TestResult),
		Links:        []*model.TraceLink{},
	}
	b.MergeGraph(g1)
	if err := b.MergeGraph(g2); err == nil {
		t.Error("expected duplicate arch error, got nil")
	}
}

// [test-spec,id=TS-GRAPH-006,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestMergeGraph_DuplicateTestSpec verifies that duplicate test-spec IDs return error.
func TestMergeGraph_DuplicateTestSpec(t *testing.T) {
	b := NewBuilder()
	g1 := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    map[string]*model.TestSpec{"TS-X": {ID: "TS-X"}},
		TestCodes:    make(map[string]*model.TestCode),
		TestResults:  make(map[string]*model.TestResult),
		Links:        []*model.TraceLink{},
	}
	g2 := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    map[string]*model.TestSpec{"TS-X": {ID: "TS-X"}},
		TestCodes:    make(map[string]*model.TestCode),
		TestResults:  make(map[string]*model.TestResult),
		Links:        []*model.TraceLink{},
	}
	b.MergeGraph(g1)
	if err := b.MergeGraph(g2); err == nil {
		t.Error("expected duplicate test-spec error, got nil")
	}
}

// [test-spec,id=TS-GRAPH-007,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestBuildLinks_ReqToArch verifies satisfied-by link from requirement to arch.
func TestBuildLinks_ReqToArch(t *testing.T) {
	b := NewBuilder()
	b.MergeGraph(newTestGraph())
	if err := b.BuildLinks(); err != nil {
		t.Fatalf("BuildLinks error: %v", err)
	}
	g := b.GetGraph()
	var found bool
	for _, link := range g.Links {
		if link.LinkType == "satisfied-by" && link.FromID == "REQ-001" && link.ToID == "comp.api" {
			found = true
		}
	}
	if !found {
		t.Error("expected satisfied-by link from REQ-001 to comp.api")
	}
}

// [test-spec,id=TS-GRAPH-008,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestBuildLinks_ArchToTestSpec verifies verified-by link from arch to test-spec.
func TestBuildLinks_ArchToTestSpec(t *testing.T) {
	b := NewBuilder()
	b.MergeGraph(newTestGraph())
	b.BuildLinks()
	var found bool
	for _, link := range b.GetGraph().Links {
		if link.LinkType == "verified-by" && link.FromID == "comp.api" && link.ToID == "TS-001" {
			found = true
		}
	}
	if !found {
		t.Error("expected verified-by link from comp.api to TS-001")
	}
}

// [test-spec,id=TS-GRAPH-009,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestBuildLinks_ReqToTestSpec verifies verified-by link from requirement to test-spec.
func TestBuildLinks_ReqToTestSpec(t *testing.T) {
	b := NewBuilder()
	b.MergeGraph(newTestGraph())
	b.BuildLinks()
	var found bool
	for _, link := range b.GetGraph().Links {
		if link.LinkType == "verified-by" && link.FromID == "REQ-001" && link.ToID == "TS-001" {
			found = true
		}
	}
	if !found {
		t.Error("expected verified-by link from REQ-001 to TS-001")
	}
}

// [test-spec,id=TS-GRAPH-010,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestBuildLinks_Dedup verifies that duplicate links are not added.
func TestBuildLinks_Dedup(t *testing.T) {
	b := NewBuilder()
	b.MergeGraph(newTestGraph())
	b.BuildLinks()
	before := len(b.GetGraph().Links)
	b.BuildLinks()
	after := len(b.GetGraph().Links)
	if after != before {
		t.Errorf("BuildLinks added duplicates: before=%d, after=%d", before, after)
	}
}

// [test-spec,id=TS-GRAPH-011,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestBuildLinks_TestCodeToSpec verifies implements link from test-code to test-spec.
func TestBuildLinks_TestCodeToSpec(t *testing.T) {
	b := NewBuilder()
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    map[string]*model.TestSpec{"TS-001": {ID: "TS-001"}},
		TestCodes:    map[string]*model.TestCode{"TC-001": {ID: "TC-001", TestSpec: "TS-001"}},
		TestResults:  make(map[string]*model.TestResult),
		Links:        []*model.TraceLink{},
	}
	b.MergeGraph(g)
	b.BuildLinks()
	var found bool
	for _, link := range b.GetGraph().Links {
		if link.LinkType == "implements" && link.FromID == "TS-001" && link.ToID == "TC-001" {
			found = true
		}
	}
	if !found {
		t.Error("expected implements link from TS-001 to TC-001")
	}
}

// [test-spec,id=TS-GRAPH-012,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestBuildLinks_TestResultNameMatch verifies name-based test result to test-code linking.
func TestBuildLinks_TestResultNameMatch(t *testing.T) {
	b := NewBuilder()
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    make(map[string]*model.TestSpec),
		TestCodes:    map[string]*model.TestCode{"TC-001": {ID: "TC-001", Function: "TestFoo"}},
		TestResults:  map[string]*model.TestResult{"TR-001": {ID: "TR-001", TestName: "TestFoo"}},
		Links:        []*model.TraceLink{},
	}
	b.MergeGraph(g)
	b.BuildLinks()
	var found bool
	for _, link := range b.GetGraph().Links {
		if link.LinkType == "produces" && link.FromID == "TC-001" && link.ToID == "TR-001" {
			found = true
		}
	}
	if !found {
		t.Error("expected produces link from TC-001 to TR-001 via name match")
	}
}

// [test-spec,id=TS-GRAPH-013,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestDeriveASPICELevels verifies ASPICE level derivation for arch hierarchy.
func TestDeriveASPICELevels(t *testing.T) {
	b := NewBuilder()
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: map[string]*model.ArchElement{
			"comp.system": {ID: "comp.system", Parent: ""},
			"comp.api":    {ID: "comp.api", Parent: "comp.system"},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:   make(map[string]*model.TestSpec),
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
		Links:       []*model.TraceLink{},
	}
	b.MergeGraph(g)
	b.DeriveASPICELevels()
	result := b.GetGraph()
	if result.ArchElements["comp.system"].ASPICE != "SWE.2" {
		t.Errorf("root ASPICE = %q, want SWE.2", result.ArchElements["comp.system"].ASPICE)
	}
	if result.ArchElements["comp.api"].ASPICE != "SWE.3" {
		t.Errorf("child ASPICE = %q, want SWE.3", result.ArchElements["comp.api"].ASPICE)
	}
}

// [test-spec,id=TS-GRAPH-014,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestDeriveASPICELevels_ExplicitNotOverridden verifies that explicit ASPICE is preserved.
func TestDeriveASPICELevels_ExplicitNotOverridden(t *testing.T) {
	b := NewBuilder()
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: map[string]*model.ArchElement{
			"comp.api": {ID: "comp.api", Parent: "comp.system", ASPICE: "SYS.2"},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:   make(map[string]*model.TestSpec),
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
		Links:       []*model.TraceLink{},
	}
	b.MergeGraph(g)
	b.DeriveASPICELevels()
	if b.GetGraph().ArchElements["comp.api"].ASPICE != "SYS.2" {
		t.Error("explicit ASPICE should not be overridden")
	}
}

// [test-spec,id=TS-GRAPH-015,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestNameMatches verifies test name matching between results and code functions.
func TestNameMatches(t *testing.T) {
	cases := []struct {
		result string
		fn     string
		want   bool
	}{
		{"TestFoo", "TestFoo", true},
		{"TestFoo", "testfoo", true},
		{"TestAPIAuth", "TestAPIAuth", true},
		{"TestFoo", "TestBar", false},
		{"TestFoo", "", false},
		{"TestFooBar", "testfoobar", true},
	}
	for _, tc := range cases {
		got := nameMatches(tc.result, tc.fn)
		if got != tc.want {
			t.Errorf("nameMatches(%q, %q) = %v, want %v", tc.result, tc.fn, got, tc.want)
		}
	}
}

// [test-spec,id=TS-GRAPH-016,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestMergeGraph_MergesLinks verifies that links are merged from other graph.
func TestMergeGraph_MergesLinks(t *testing.T) {
	b := NewBuilder()
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    make(map[string]*model.TestSpec),
		TestCodes:    make(map[string]*model.TestCode),
		TestResults:  make(map[string]*model.TestResult),
		Links: []*model.TraceLink{
			{FromID: "A", ToID: "B", LinkType: "satisfied-by"},
		},
	}
	b.MergeGraph(g)
	if len(b.GetGraph().Links) != 1 {
		t.Errorf("expected 1 link after merge, got %d", len(b.GetGraph().Links))
	}
}
