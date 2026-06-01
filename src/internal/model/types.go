package model

import (
	"time"
)

// Requirement represents a REQ42 block.
type Requirement struct {
	ID           string            `json:"id"`           // Unique identifier (e.g., "SWE-001")
	Title        string            `json:"title"`        // Requirement title
	Text         string            `json:"text"`         // Full requirement text
	Version      int               `json:"version"`      // Version number
	Priority     string            `json:"priority"`     // high, medium, low
	ASPICE       string            `json:"aspice"`       // ASPICE process (e.g., "SWE.1")
	Status       string            `json:"status"`       // approved, draft, deprecated
	ReviewedBy   string            `json:"reviewed_by"`  // Reviewer ID
	ReviewedDate time.Time         `json:"reviewed_date"`
	Derives      []string          `json:"derives"`      // Derived from (other requirement IDs)
	Project      string            `json:"project"`      // Project key from config
	FilePath     string            `json:"file_path"`    // Source .adoc file
	LineNumber   int               `json:"line_number"`  // Line in source file
	Attributes   map[string]string `json:"attributes"`   // Additional attributes
}

// ArchElement represents an ARC42 [arch] block.
type ArchElement struct {
	ID          string            `json:"id"`           // Dot-separated ID (e.g., "comp.api.auth")
	Title       string            `json:"title"`        // Element title
	Text        string            `json:"text"`         // Description
	Parent      string            `json:"parent"`       // Parent element ID (empty for Level 1)
	ASPICE      string            `json:"aspice"`       // ASPICE process (auto-derived from hierarchy)
	Req         []string          `json:"req"`          // Referenced requirement IDs
	Impl        string            `json:"impl"`         // Implementation reference (file path)
	TestSpec    string            `json:"test_spec"`    // Linked test specification ID
	Project     string            `json:"project"`      // Project key
	FilePath    string            `json:"file_path"`    // Source .adoc file
	LineNumber  int               `json:"line_number"`  // Line in source file
	Attributes  map[string]string `json:"attributes"`   // Additional attributes
}

// DesignElement represents a [dsn] block (SWE.3 Software Detailed Design).
type DesignElement struct {
	ID         string            `json:"id"`          // Dot-separated ID (e.g., "comp.parser.tokenizer")
	Title      string            `json:"title"`       // Element title
	Text       string            `json:"text"`        // Description
	Arch       string            `json:"arch"`        // Parent architecture element ID (SWE.2)
	ASPICE     string            `json:"aspice"`      // ASPICE process (typically SWE.3)
	Impl       string            `json:"impl"`        // Implementation reference (file path)
	Project    string            `json:"project"`     // Project key
	FilePath   string            `json:"file_path"`   // Source .adoc file
	LineNumber int               `json:"line_number"` // Line in source file
	Attributes map[string]string `json:"attributes"`  // Additional attributes
}

// TestSpec represents a [test-spec] block.
type TestSpec struct {
	ID         string            `json:"id"`          // Unique identifier (e.g., "spec.api.auth")
	Title      string            `json:"title"`       // Specification title
	Text       string            `json:"text"`        // Full specification
	Req        []string          `json:"req"`         // Associated requirement IDs (SWE.6)
	Arch       []string          `json:"arch"`        // Associated architecture element IDs (SWE.5)
	Dsn        []string          `json:"dsn"`         // Associated design element IDs (SWE.4)
	Project    string            `json:"project"`     // Project key
	FilePath   string            `json:"file_path"`   // Source .adoc file
	LineNumber int               `json:"line_number"` // Line in source file
	Attributes map[string]string `json:"attributes"`  // Additional attributes
}

// TestCode represents a [test-code] block or code annotation.
type TestCode struct {
	ID          string            `json:"id"`           // Unique identifier
	TestSpec    string            `json:"test_spec"`    // Linked TestSpec ID
	File        string            `json:"file"`         // Test file path
	Function    string            `json:"function"`     // Test function/method name
	Language    string            `json:"language"`     // go, python, java, etc.
	Project     string            `json:"project"`      // Project key
	FilePath    string            `json:"file_path"`    // Source file (.adoc or code)
	LineNumber  int               `json:"line_number"`  // Line in source file
	Attributes  map[string]string `json:"attributes"`   // Additional attributes
}

// TestResult represents a parsed test execution result from JUnit XML or go-test JSON.
type TestResult struct {
	ID          string            `json:"id"`           // Unique ID in report
	Package     string            `json:"package"`      // Test package/class
	TestName    string            `json:"test_name"`    // Test function/method name
	FullName    string            `json:"full_name"`    // Full qualified name (package::test)
	Duration    float64           `json:"duration"`     // Execution time in seconds
	Status      string            `json:"status"`       // passed, failed, skipped
	Error       string            `json:"error"`        // Error message (if failed)
	Stdout      string            `json:"stdout"`       // Captured stdout
	Stderr      string            `json:"stderr"`       // Captured stderr
	Timestamp   time.Time         `json:"timestamp"`    // Execution time
	Project     string            `json:"project"`      // Project key
	Platform    string            `json:"platform"`     // linux, windows, macos
	LinkedSpec  string            `json:"linked_spec"`  // Linked TestSpec ID (from tracing)
	LinkedCode  string            `json:"linked_code"`  // Linked TestCode ID (from tracing)
	Attributes  map[string]string `json:"attributes"`   // Additional attributes
}

// TraceLink represents a bidirectional link in the traceability graph.
type TraceLink struct {
	FromID   string `json:"from_id"`   // Source artifact ID
	FromType string `json:"from_type"` // requirement, arch, test-spec, test-code, test-result
	ToID     string `json:"to_id"`     // Target artifact ID
	ToType   string `json:"to_type"`   // requirement, arch, test-spec, test-code, test-result
	LinkType string `json:"link_type"` // satisfies, implements, verifies, derives, covers
	Status   string `json:"status"`    // active, stale, broken
	Reason   string `json:"reason"`    // Why this link exists
}

// TraceabilityGraph represents the complete dependency graph.
type TraceabilityGraph struct {
	Requirements   map[string]*Requirement   `json:"requirements"`
	ArchElements   map[string]*ArchElement   `json:"arch_elements"`
	DesignElements map[string]*DesignElement `json:"design_elements"`
	TestSpecs      map[string]*TestSpec      `json:"test_specs"`
	TestCodes      map[string]*TestCode      `json:"test_codes"`
	TestResults    map[string]*TestResult    `json:"test_results"`
	Links          []*TraceLink              `json:"links"`
}

// EmptyGraph returns a TraceabilityGraph with all maps initialised (no nil panics).
func EmptyGraph() *TraceabilityGraph {
	return &TraceabilityGraph{
		Requirements:   make(map[string]*Requirement),
		ArchElements:   make(map[string]*ArchElement),
		DesignElements: make(map[string]*DesignElement),
		TestSpecs:      make(map[string]*TestSpec),
		TestCodes:      make(map[string]*TestCode),
		TestResults:    make(map[string]*TestResult),
		Links:          []*TraceLink{},
	}
}

// GapAnalysisResult represents findings from gap analysis.
type GapAnalysisResult struct {
	OrphanRequirements      []*Requirement   // Requirements without architecture or test coverage
	OrphanArchElements      []*ArchElement   // Architecture elements without requirements
	OrphanTestSpecs         []*TestSpec      // Test specs without linked requirements
	UntracedTestResults     []*TestResult    // Test results without linked specifications
	MissingImplementation   []*ArchElement   // Architecture without impl references
	StaleTraces             []*TraceLink     // Links to outdated versions
	UntestedArchElements    []*ArchElement   // SWE.2 arch elements without an integration test (arch= on test-spec)
	OrphanDesignElements    []*DesignElement // SWE.3 design elements without an arch= parent
	UntestedDesignElements  []*DesignElement // SWE.3 design elements without a unit test (dsn= on test-spec)
}

// CoverageReport represents traceability coverage metrics.
type CoverageReport struct {
	TotalRequirements       int     `json:"total_requirements"`
	CoveredByArch           int     `json:"covered_by_arch"`
	CoveredByTests          int     `json:"covered_by_tests"`
	RequirementCoverage     float64 `json:"requirement_coverage"`      // Percentage
	ArchCoverage            float64 `json:"arch_coverage"`             // Percentage
	TestCoverage            float64 `json:"test_coverage"`             // Percentage
	TotalTestResults        int     `json:"total_test_results"`
	PassedTests             int     `json:"passed_tests"`
	FailedTests             int     `json:"failed_tests"`
	SkippedTests            int     `json:"skipped_tests"`
	PassRate                float64 `json:"pass_rate"`                 // Percentage
}

// ASPICEProcessLevel represents a process definition in PAM 4.0.
type ASPICEProcessLevel struct {
	ID          string `json:"id"`          // e.g., "SWE.1"
	Name        string `json:"name"`        // Process name
	Description string `json:"description"` // Process description
}

// ASPICEBestPractice represents a single best practice within a process.
type ASPICEBestPractice struct {
	ID          string `json:"id"`          // e.g., "SWE.1.BP1"
	ProcessID   string `json:"process_id"`  // Parent process
	Title       string `json:"title"`       // BP title
	Description string `json:"description"` // BP description
}

// ASPICECheckResult represents the result of checking one best practice.
type ASPICECheckResult struct {
	BP       *ASPICEBestPractice `json:"bp"`
	Status   string              `json:"status"` // satisfied, not-satisfied, partial
	Coverage float64             `json:"coverage"` // Percentage (0-100)
	Evidence []*TraceLink        `json:"evidence"` // Supporting trace links
	Gaps     []string            `json:"gaps"`    // Gap descriptions
}

// ASPICEReport represents the result of ASPICE validation.
type ASPICEReport struct {
	Processes map[string][]*ASPICECheckResult `json:"processes"` // Process ID -> BP results
	Overall   float64                         `json:"overall"`   // Overall coverage percentage
}
