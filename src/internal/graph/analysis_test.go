package graph

import (
	"sync"
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

func buildAnalysisGraph() *model.TraceabilityGraph {
	return &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{
			"REQ-001": {ID: "REQ-001", Project: "proj"},
			"REQ-002": {ID: "REQ-002", Project: "proj"},
		},
		ArchElements: map[string]*model.ArchElement{
			"comp.api":    {ID: "comp.api", Parent: "comp.system", Req: []string{"REQ-001"}},
			"comp.system": {ID: "comp.system", Parent: ""},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"TS-001": {ID: "TS-001", Req: []string{"REQ-001"}, Arch: []string{"comp.api"}},
		},
		TestCodes: make(map[string]*model.TestCode),
		TestResults: map[string]*model.TestResult{
			"linux::pkg::TestA": {ID: "linux::pkg::TestA", TestName: "TestA", Status: "passed", Project: "proj"},
			"linux::pkg::TestB": {ID: "linux::pkg::TestB", TestName: "TestB", Status: "failed", Project: "proj"},
			"linux::pkg::TestC": {ID: "linux::pkg::TestC", TestName: "TestC", Status: "skipped", Project: "proj"},
		},
		Links: []*model.TraceLink{
			{FromID: "REQ-001", ToID: "comp.api", FromType: "requirement", ToType: "arch", LinkType: "satisfied-by", Status: "active"},
			{FromID: "comp.api", ToID: "TS-001", FromType: "arch", ToType: "test-spec", LinkType: "verified-by", Status: "active"},
		},
	}
}

// [test-spec,id=TS-GRAPH-017,req="REQ-GRAPH-001",dsn=comp.graph.analyzer,aspice=SWE.4-BP4]
// TestNewAnalyzer verifies that NewAnalyzer stores the graph correctly.
func TestNewAnalyzer(t *testing.T) {
	g := buildAnalysisGraph()
	a := NewAnalyzer(g)
	if a.GetGraph() != g {
		t.Error("GetGraph should return the same graph passed to NewAnalyzer")
	}
}

// [test-spec,id=TS-GRAPH-018,req="REQ-GRAPH-001",dsn=comp.graph.analyzer,aspice="SWE.5.BP3"]
// TestAnalyzeGaps_OrphanRequirement verifies that uncovered requirements are detected.
func TestAnalyzeGaps_OrphanRequirement(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	gaps := a.AnalyzeGaps()
	var found bool
	for _, r := range gaps.OrphanRequirements {
		if r.ID == "REQ-002" {
			found = true
		}
	}
	if !found {
		t.Error("REQ-002 should be an orphan requirement")
	}
}

// [test-spec,id=TS-GRAPH-019,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestAnalyzeGaps_MissingImpl verifies that arch elements without impl are flagged.
func TestAnalyzeGaps_MissingImpl(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	gaps := a.AnalyzeGaps()
	var found bool
	for _, ae := range gaps.MissingImplementation {
		if ae.ID == "comp.api" {
			found = true
		}
	}
	if !found {
		t.Error("comp.api (no impl, has parent) should be in MissingImplementation")
	}
}

// [test-spec,id=TS-GRAPH-020,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestAnalyzeGaps_UntracedTestResults verifies unlinked test results are flagged.
func TestAnalyzeGaps_UntracedTestResults(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	gaps := a.AnalyzeGaps()
	if len(gaps.UntracedTestResults) != 3 {
		t.Errorf("UntracedTestResults = %d, want 3", len(gaps.UntracedTestResults))
	}
}

// [test-spec,id=TS-GRAPH-021,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestAnalyzeGaps_OrphanTestSpec verifies detection of unlinked test specs.
func TestAnalyzeGaps_OrphanTestSpec(t *testing.T) {
	g := buildAnalysisGraph()
	g.TestSpecs["TS-ORPHAN"] = &model.TestSpec{ID: "TS-ORPHAN"}
	a := NewAnalyzer(g)
	gaps := a.AnalyzeGaps()
	var found bool
	for _, s := range gaps.OrphanTestSpecs {
		if s.ID == "TS-ORPHAN" {
			found = true
		}
	}
	if !found {
		t.Error("TS-ORPHAN should be in OrphanTestSpecs")
	}
}

// [test-spec,id=TS-GRAPH-022,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestAnalyzeGaps_StaleTraces verifies that stale links are flagged.
func TestAnalyzeGaps_StaleTraces(t *testing.T) {
	g := buildAnalysisGraph()
	g.Links = append(g.Links, &model.TraceLink{FromID: "X", ToID: "Y", Status: "stale"})
	a := NewAnalyzer(g)
	gaps := a.AnalyzeGaps()
	if len(gaps.StaleTraces) != 1 {
		t.Errorf("StaleTraces = %d, want 1", len(gaps.StaleTraces))
	}
}

// [test-spec,id=TS-GRAPH-023,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestCalculateCoverage verifies coverage metrics calculation.
func TestCalculateCoverage(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	cov := a.CalculateCoverage()
	if cov.TotalRequirements != 2 {
		t.Errorf("TotalRequirements = %d, want 2", cov.TotalRequirements)
	}
	if cov.CoveredByArch != 1 {
		t.Errorf("CoveredByArch = %d, want 1", cov.CoveredByArch)
	}
	if cov.RequirementCoverage != 50.0 {
		t.Errorf("RequirementCoverage = %.1f, want 50.0", cov.RequirementCoverage)
	}
	if cov.TotalTestResults != 3 {
		t.Errorf("TotalTestResults = %d, want 3", cov.TotalTestResults)
	}
	if cov.PassedTests != 1 {
		t.Errorf("PassedTests = %d, want 1", cov.PassedTests)
	}
	if cov.FailedTests != 1 {
		t.Errorf("FailedTests = %d, want 1", cov.FailedTests)
	}
	if cov.SkippedTests != 1 {
		t.Errorf("SkippedTests = %d, want 1", cov.SkippedTests)
	}
}

// [test-spec,id=TS-GRAPH-024,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestCalculateCoverage_EmptyGraph verifies zero coverage for empty graph.
func TestCalculateCoverage_EmptyGraph(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    make(map[string]*model.TestSpec),
		TestCodes:    make(map[string]*model.TestCode),
		TestResults:  make(map[string]*model.TestResult),
		Links:        []*model.TraceLink{},
	}
	a := NewAnalyzer(g)
	cov := a.CalculateCoverage()
	if cov.RequirementCoverage != 0 {
		t.Errorf("expected 0%% coverage for empty graph, got %.1f", cov.RequirementCoverage)
	}
	if cov.PassRate != 0 {
		t.Errorf("expected 0%% pass rate for empty graph, got %.1f", cov.PassRate)
	}
}

// [test-spec,id=TS-GRAPH-025,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestCalculateCoverage_ArchCoverage verifies arch coverage metric.
func TestCalculateCoverage_ArchCoverage(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	cov := a.CalculateCoverage()
	// 2 arch elements total: comp.api (has req=REQ-001) + comp.system (no req) → 50%
	const want = 50.0
	if cov.ArchCoverage != want {
		t.Errorf("ArchCoverage = %.1f, want %.1f", cov.ArchCoverage, want)
	}
}

// [test-spec,id=TS-GRAPH-026,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestGetOrphanRequirementsByProject verifies project-scoped orphan filtering.
func TestGetOrphanRequirementsByProject(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	orphans := a.GetOrphanRequirementsByProject("proj")
	if len(orphans) != 1 || orphans[0].ID != "REQ-002" {
		t.Errorf("expected [REQ-002], got %v", orphans)
	}
}

// [test-spec,id=TS-GRAPH-027,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestGetOrphanRequirementsByProject_WrongProject verifies no results for unknown project.
func TestGetOrphanRequirementsByProject_WrongProject(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	orphans := a.GetOrphanRequirementsByProject("other-project")
	if len(orphans) != 0 {
		t.Errorf("expected no orphans for unknown project, got %d", len(orphans))
	}
}

// [test-spec,id=TS-GRAPH-028,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestGetCoverageByProject verifies project-filtered coverage report.
func TestGetCoverageByProject(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	cov := a.GetCoverageByProject("proj")
	if cov.TotalRequirements != 2 {
		t.Errorf("TotalRequirements = %d, want 2", cov.TotalRequirements)
	}
	if cov.CoveredByArch != 1 {
		t.Errorf("CoveredByArch = %d, want 1", cov.CoveredByArch)
	}
	if cov.TotalTestResults != 3 {
		t.Errorf("TotalTestResults = %d, want 3", cov.TotalTestResults)
	}
}

// [test-spec,id=TS-GRAPH-029,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestGetLinksFor verifies that links from/to an artifact are returned.
func TestGetLinksFor(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	links := a.GetLinksFor("REQ-001")
	if len(links) == 0 {
		t.Error("expected links for REQ-001, got none")
	}
	links2 := a.GetLinksFor("comp.api")
	if len(links2) == 0 {
		t.Error("expected links for comp.api, got none")
	}
}

// [test-spec,id=TS-GRAPH-030,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestGetLinksFor_NoLinks verifies empty result for artifact without links.
func TestGetLinksFor_NoLinks(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	links := a.GetLinksFor("UNKNOWN")
	if len(links) != 0 {
		t.Errorf("expected no links for unknown artifact, got %d", len(links))
	}
}

// [test-spec,id=TS-GRAPH-031,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestValidateReferences_BrokenReq verifies detection of broken requirement references.
func TestValidateReferences_BrokenReq(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: map[string]*model.ArchElement{
			"comp.api": {ID: "comp.api", Req: []string{"REQ-NONEXISTENT"}},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:   make(map[string]*model.TestSpec),
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
		Links:       []*model.TraceLink{},
	}
	a := NewAnalyzer(g)
	errors := a.ValidateReferences()
	if len(errors) == 0 {
		t.Error("expected validation errors for broken req reference")
	}
}

// [test-spec,id=TS-GRAPH-032,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestValidateReferences_BrokenTestSpecArch verifies broken arch references in test-specs.
func TestValidateReferences_BrokenTestSpecArch(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"TS-001": {ID: "TS-001", Arch: []string{"comp.nonexistent"}},
		},
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
		Links:       []*model.TraceLink{},
	}
	a := NewAnalyzer(g)
	errs := a.ValidateReferences()
	if len(errs) == 0 {
		t.Error("expected error for broken arch reference in test-spec")
	}
}

// [test-spec,id=TS-GRAPH-033,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestValidateReferences_BrokenTestSpecReq verifies broken req references in test-specs.
func TestValidateReferences_BrokenTestSpecReq(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"TS-001": {ID: "TS-001", Req: []string{"REQ-NONEXISTENT"}},
		},
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
		Links:       []*model.TraceLink{},
	}
	a := NewAnalyzer(g)
	errs := a.ValidateReferences()
	if len(errs) == 0 {
		t.Error("expected error for broken req reference in test-spec")
	}
}

// [test-spec,id=TS-GRAPH-034,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestValidateReferences_BrokenTestCodeSpec verifies broken test-spec references in test-codes.
func TestValidateReferences_BrokenTestCodeSpec(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    make(map[string]*model.TestSpec),
		TestCodes: map[string]*model.TestCode{
			"TC-001": {ID: "TC-001", TestSpec: "TS-NONEXISTENT"},
		},
		TestResults: make(map[string]*model.TestResult),
		Links:       []*model.TraceLink{},
	}
	a := NewAnalyzer(g)
	errs := a.ValidateReferences()
	if len(errs) == 0 {
		t.Error("expected error for broken test-spec reference in test-code")
	}
}

// [test-spec,id=TS-GRAPH-035,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestValidateReferences_Clean verifies no errors with fully valid references.
func TestValidateReferences_Clean(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	errors := a.ValidateReferences()
	if len(errors) != 0 {
		t.Errorf("expected no errors for valid graph, got: %v", errors)
	}
}

// [test-spec,id=TS-GRAPH-036,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestCalculateCoverage_PassRate verifies pass rate calculation.
func TestCalculateCoverage_PassRate(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    make(map[string]*model.TestSpec),
		TestCodes:    make(map[string]*model.TestCode),
		TestResults: map[string]*model.TestResult{
			"T1": {ID: "T1", Status: "passed"},
			"T2": {ID: "T2", Status: "passed"},
			"T3": {ID: "T3", Status: "failed"},
		},
		Links: []*model.TraceLink{},
	}
	a := NewAnalyzer(g)
	cov := a.CalculateCoverage()
	// PassRate = 2/3 * 100 ≈ 66.67
	if cov.PassRate < 66 || cov.PassRate > 67 {
		t.Errorf("PassRate = %.2f, want ~66.67", cov.PassRate)
	}
}

// [test-spec,id=TS-GRAPH-037,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestAnalyzer_ConcurrentBuildIndex verifies thread-safe lazy index construction.
func TestAnalyzer_ConcurrentBuildIndex(t *testing.T) {
	a := NewAnalyzer(buildAnalysisGraph())
	var wg sync.WaitGroup
	// 20 goroutines all trigger buildIndex() simultaneously
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = a.AnalyzeGaps()
			_ = a.CalculateCoverage()
		}()
	}
	wg.Wait()
	// If there's a data race, -race will catch it; if maps are corrupt this panics
}

// [test-spec,id=TS-GRAPH-038,req="REQ-GRAPH-001",aspice="SWE.5.BP3"]
// TestMergeGraph_LinkSeenPopulated verifies MergeGraph registers existing links in linkSeen.
func TestMergeGraph_LinkSeenPopulated(t *testing.T) {
	existingLink := &model.TraceLink{
		FromID: "REQ-001", FromType: "requirement",
		ToID: "comp.api", ToType: "arch",
		LinkType: "satisfied-by", Status: "active",
	}
	b := NewBuilder()
	g := &model.TraceabilityGraph{
		Requirements: map[string]*model.Requirement{"REQ-001": {ID: "REQ-001"}},
		ArchElements: map[string]*model.ArchElement{"comp.api": {ID: "comp.api", Req: []string{"REQ-001"}}},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:    make(map[string]*model.TestSpec),
		TestCodes:    make(map[string]*model.TestCode),
		TestResults:  make(map[string]*model.TestResult),
		Links:        []*model.TraceLink{existingLink},
	}
	b.MergeGraph(g)
	linkCountBefore := len(b.GetGraph().Links)

	// BuildLinks tries to add the same req→arch link — should be deduped
	b.BuildLinks()
	linkCountAfter := len(b.GetGraph().Links)
	if linkCountAfter != linkCountBefore {
		t.Errorf("MergeGraph+BuildLinks added duplicate link: before=%d, after=%d", linkCountBefore, linkCountAfter)
	}
}

// [test-spec,id=TS-GRAPH-039,req=REQ-GRAPH-001,aspice=SWE.5-BP4]
// TestAnalyzeGaps_UntestedArchElement verifies that top-level arch elements without
// an integration test-spec (arch= linkage) are flagged as SWE.5 gaps.
func TestAnalyzeGaps_UntestedArchElement(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: map[string]*model.ArchElement{
			// impl= gesetzt → wird als SWE.5-Gap gemeldet (braucht Integrationstest)
			"comp.system": {ID: "comp.system", Parent: "", Title: "System Component", Impl: "internal/system"},
			// kein impl= → konzeptionelles Element, kein Gap
			"intro.goals": {ID: "intro.goals", Parent: "", Title: "Quality Goals"},
			"comp.sub":    {ID: "comp.sub", Parent: "comp.system", Title: "Sub Component"},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:   make(map[string]*model.TestSpec),
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
		Links:       []*model.TraceLink{},
	}
	a := NewAnalyzer(g)
	gaps := a.AnalyzeGaps()

	if len(gaps.UntestedArchElements) != 1 {
		t.Fatalf("expected 1 untested arch element (impl only), got %d", len(gaps.UntestedArchElements))
	}
	if gaps.UntestedArchElements[0].ID != "comp.system" {
		t.Errorf("expected comp.system flagged, got %s", gaps.UntestedArchElements[0].ID)
	}
}

// [test-spec,id=TS-GRAPH-041,req=REQ-GRAPH-001,aspice=SWE.5-BP4]
// TestAnalyzeGaps_ConceptualArchNotFlagged verifies that arch elements without impl=
// (documentation/conceptual) are NOT reported as SWE.5 gaps.
func TestAnalyzeGaps_ConceptualArchNotFlagged(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: map[string]*model.ArchElement{
			"intro.goals":          {ID: "intro.goals", Parent: "", Title: "Quality Goals"},
			"constraints.technical": {ID: "constraints.technical", Parent: "", Title: "Technical Constraints"},
			"quality.performance":  {ID: "quality.performance", Parent: "", Title: "Performance"},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:      make(map[string]*model.TestSpec),
		TestCodes:      make(map[string]*model.TestCode),
		TestResults:    make(map[string]*model.TestResult),
		Links:          []*model.TraceLink{},
	}
	a := NewAnalyzer(g)
	gaps := a.AnalyzeGaps()

	if len(gaps.UntestedArchElements) != 0 {
		t.Errorf("expected 0 untested arch elements for conceptual-only graph, got %d: %v",
			len(gaps.UntestedArchElements), gaps.UntestedArchElements)
	}
}

// [test-spec,id=TS-GRAPH-040,req=REQ-GRAPH-001,aspice=SWE.5-BP4]
// TestAnalyzeGaps_TestedArchElement verifies that arch elements WITH an integration
// test-spec (verified-by link) are NOT flagged as SWE.5 gaps.
func TestAnalyzeGaps_TestedArchElement(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements: make(map[string]*model.Requirement),
		ArchElements: map[string]*model.ArchElement{
			"comp.system": {ID: "comp.system", Parent: "", Title: "System Component"},
		},
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs: map[string]*model.TestSpec{
			"TS-INT-001": {ID: "TS-INT-001", Arch: []string{"comp.system"}},
		},
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
		Links: []*model.TraceLink{
			{FromID: "comp.system", ToID: "TS-INT-001", FromType: "arch", ToType: "test-spec", LinkType: "verified-by", Status: "active"},
		},
	}
	a := NewAnalyzer(g)
	gaps := a.AnalyzeGaps()

	if len(gaps.UntestedArchElements) != 0 {
		t.Errorf("expected no untested arch elements, got %d: %v", len(gaps.UntestedArchElements), gaps.UntestedArchElements)
	}
}

// [test-spec,id=TS-GRAPH-041,req=REQ-GRAPH-001,aspice=SWE.4-BP4]
// TestAnalyzeGaps_OrphanDesignElement verifies that dsn elements without arch= are flagged.
func TestAnalyzeGaps_OrphanDesignElement(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements:   make(map[string]*model.Requirement),
		ArchElements:   make(map[string]*model.ArchElement),
		DesignElements: map[string]*model.DesignElement{
			"comp.tokenizer": {ID: "comp.tokenizer", Arch: ""},
		},
		TestSpecs:   make(map[string]*model.TestSpec),
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
		Links:       []*model.TraceLink{},
	}
	gaps := NewAnalyzer(g).AnalyzeGaps()
	if len(gaps.OrphanDesignElements) != 1 {
		t.Errorf("expected 1 orphan dsn, got %d", len(gaps.OrphanDesignElements))
	}
}

// [test-spec,id=TS-GRAPH-042,req=REQ-GRAPH-001,aspice=SWE.4-BP4]
// TestAnalyzeGaps_UntestedDesignElement verifies that dsn elements without a unit test are flagged.
func TestAnalyzeGaps_UntestedDesignElement(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements:   make(map[string]*model.Requirement),
		ArchElements:   make(map[string]*model.ArchElement),
		DesignElements: map[string]*model.DesignElement{
			"comp.tokenizer": {ID: "comp.tokenizer", Arch: "comp.parser"},
		},
		TestSpecs:   make(map[string]*model.TestSpec),
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
		Links:       []*model.TraceLink{},
	}
	gaps := NewAnalyzer(g).AnalyzeGaps()
	if len(gaps.UntestedDesignElements) != 1 {
		t.Errorf("expected 1 untested dsn, got %d", len(gaps.UntestedDesignElements))
	}
}

// [test-spec,id=TS-GRAPH-043,req=REQ-GRAPH-001,aspice=SWE.4-BP4]
// TestAnalyzeGaps_TestedDesignElement verifies that covered dsn elements are not flagged.
func TestAnalyzeGaps_TestedDesignElement(t *testing.T) {
	g := &model.TraceabilityGraph{
		Requirements:   make(map[string]*model.Requirement),
		ArchElements:   make(map[string]*model.ArchElement),
		DesignElements: map[string]*model.DesignElement{
			"comp.tokenizer": {ID: "comp.tokenizer", Arch: "comp.parser"},
		},
		TestSpecs: map[string]*model.TestSpec{
			"TS-UNIT-001": {ID: "TS-UNIT-001", Dsn: []string{"comp.tokenizer"}},
		},
		TestCodes:   make(map[string]*model.TestCode),
		TestResults: make(map[string]*model.TestResult),
		Links: []*model.TraceLink{
			{FromID: "comp.tokenizer", ToID: "TS-UNIT-001", FromType: "design", ToType: "test-spec", LinkType: "verified-by", Status: "active"},
		},
	}
	gaps := NewAnalyzer(g).AnalyzeGaps()
	if len(gaps.UntestedDesignElements) != 0 {
		t.Errorf("expected no untested dsn elements, got %d", len(gaps.UntestedDesignElements))
	}
}
