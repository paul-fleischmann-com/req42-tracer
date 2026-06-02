package validation

import (
	"fmt"
	"os"
	"strings"

	"github.com/paulefl/req42-tracer/src/internal/graph"
	"github.com/paulefl/req42-tracer/src/internal/model"
)

// Severity represents the severity of a rule violation.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityOff     Severity = "off"
)

// Violation represents a single rule violation.
type Violation struct {
	Rule     string
	Severity Severity
	Message  string
	Location string // optional: file:line or artifact ID
}

// RuleResult holds the outcome of running a single rule.
type RuleResult struct {
	RuleID     string
	Severity   Severity
	Violations []Violation
}

// RuleEngine evaluates custom validation rules against a traceability graph.
type RuleEngine struct {
	config   *model.Config
	analyzer *graph.Analyzer
}

// NewRuleEngine creates a new rule engine.
func NewRuleEngine(config *model.Config, analyzer *graph.Analyzer) *RuleEngine {
	return &RuleEngine{config: config, analyzer: analyzer}
}

// Run evaluates all configured rules and returns violations grouped by rule.
func (e *RuleEngine) Run() []*RuleResult {
	ruleRunners := map[string]func(Severity) *RuleResult{
		"missing-review":              e.ruleMissingReview,
		"missing-test-spec":           e.ruleMissingTestSpec,
		"missing-impl":                e.ruleMissingImpl,
		"impl-file-not-found":         e.ruleImplFileNotFound,
		"orphan-architecture":         e.ruleOrphanArchitecture,
		"orphan-tests":                e.ruleOrphanTests,
		"stale-traces":                e.ruleStaleTraces,
		"all-reqs-must-have-aspice":   e.ruleAllReqsMustHaveASPICE,
		"all-reqs-must-have-priority": e.ruleAllReqsMustHavePriority,
		"max-orphan-percentage":       e.ruleMaxOrphanPercentage,
		// undocumented-bausteinsicht-elements is intentionally not registered:
		// it requires the bausteinsicht binary which is not available at validate-time.
		// Users who set this rule get an "unknown rule" warning explaining this.
	}

	var results []*RuleResult
	for ruleID, sev := range e.config.Rules {
		severity := parseSeverity(sev)
		if severity == SeverityOff {
			continue
		}
		runner, ok := ruleRunners[ruleID]
		if !ok {
			results = append(results, &RuleResult{
				RuleID:   ruleID,
				Severity: SeverityWarning,
				Violations: []Violation{{
					Rule:    ruleID,
					Severity: SeverityWarning,
					Message: fmt.Sprintf("unknown rule %q — check .req42.yaml", ruleID),
				}},
			})
			continue
		}
		result := runner(severity)
		if result != nil {
			results = append(results, result)
		}
	}
	return results
}

// TotalViolations returns the total number of violations across all results.
func TotalViolations(results []*RuleResult) (errors, warnings int) {
	for _, r := range results {
		for _, v := range r.Violations {
			switch v.Severity {
			case SeverityError:
				errors++
			case SeverityWarning:
				warnings++
			}
		}
	}
	return
}

// FormatResults formats rule results as a human-readable string.
func FormatResults(results []*RuleResult) string {
	var sb strings.Builder
	for _, r := range results {
		if len(r.Violations) == 0 {
			continue
		}
		for _, v := range r.Violations {
			prefix := "⚠️  [WARNING]"
			if v.Severity == SeverityError {
				prefix = "❌ [ERROR]"
			}
			loc := ""
			if v.Location != "" {
				loc = " (" + v.Location + ")"
			}
			sb.WriteString(fmt.Sprintf("%s %s%s\n", prefix, v.Message, loc))
		}
	}
	return sb.String()
}

// parseSeverity parses a severity string, defaulting to warning for unknown values.
func parseSeverity(s string) Severity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "error":
		return SeverityError
	case "off", "disabled", "none":
		return SeverityOff
	default:
		return SeverityWarning
	}
}

// ruleMissingReview flags requirements without a reviewed-by attribute.
func (e *RuleEngine) ruleMissingReview(sev Severity) *RuleResult {
	g := e.analyzer.GetGraph()
	result := &RuleResult{RuleID: "missing-review", Severity: sev}
	for _, req := range g.Requirements {
		if req.ReviewedBy == "" {
			result.Violations = append(result.Violations, Violation{
				Rule:     "missing-review",
				Severity: sev,
				Message:  fmt.Sprintf("requirement %s has no reviewed-by", req.ID),
				Location: fmt.Sprintf("%s:%d", req.FilePath, req.LineNumber),
			})
		}
	}
	return result
}

// ruleMissingTestSpec flags arch elements without a linked test spec.
func (e *RuleEngine) ruleMissingTestSpec(sev Severity) *RuleResult {
	g := e.analyzer.GetGraph()
	result := &RuleResult{RuleID: "missing-test-spec", Severity: sev}
	specArchIdx := buildTestSpecArchIndex(g) // O(|TestSpecs|) once
	for _, arch := range g.ArchElements {
		if arch.Parent == "" {
			continue // skip top-level elements
		}
		_, hasLink := specArchIdx[arch.ID]
		if arch.TestSpec == "" && !hasLink {
			result.Violations = append(result.Violations, Violation{
				Rule:     "missing-test-spec",
				Severity: sev,
				Message:  fmt.Sprintf("architecture element %s has no test specification", arch.ID),
				Location: fmt.Sprintf("%s:%d", arch.FilePath, arch.LineNumber),
			})
		}
	}
	return result
}

// ruleMissingImpl flags arch elements without an implementation reference.
func (e *RuleEngine) ruleMissingImpl(sev Severity) *RuleResult {
	g := e.analyzer.GetGraph()
	result := &RuleResult{RuleID: "missing-impl", Severity: sev}
	for _, arch := range g.ArchElements {
		if arch.Parent != "" && arch.Impl == "" {
			result.Violations = append(result.Violations, Violation{
				Rule:     "missing-impl",
				Severity: sev,
				Message:  fmt.Sprintf("architecture element %s has no implementation reference", arch.ID),
				Location: fmt.Sprintf("%s:%d", arch.FilePath, arch.LineNumber),
			})
		}
	}
	return result
}

// ruleOrphanArchitecture flags arch elements not linked to any requirement.
func (e *RuleEngine) ruleOrphanArchitecture(sev Severity) *RuleResult {
	gaps := e.analyzer.AnalyzeGaps()
	result := &RuleResult{RuleID: "orphan-architecture", Severity: sev}
	for _, arch := range gaps.OrphanArchElements {
		result.Violations = append(result.Violations, Violation{
			Rule:     "orphan-architecture",
			Severity: sev,
			Message:  fmt.Sprintf("architecture element %s not linked to any requirement", arch.ID),
			Location: fmt.Sprintf("%s:%d", arch.FilePath, arch.LineNumber),
		})
	}
	return result
}

// ruleOrphanTests flags test specs not linked to requirements or arch.
func (e *RuleEngine) ruleOrphanTests(sev Severity) *RuleResult {
	gaps := e.analyzer.AnalyzeGaps()
	result := &RuleResult{RuleID: "orphan-tests", Severity: sev}
	for _, spec := range gaps.OrphanTestSpecs {
		result.Violations = append(result.Violations, Violation{
			Rule:     "orphan-tests",
			Severity: sev,
			Message:  fmt.Sprintf("test spec %s not linked to any requirement or architecture", spec.ID),
			Location: fmt.Sprintf("%s:%d", spec.FilePath, spec.LineNumber),
		})
	}
	return result
}

// ruleStaleTraces flags trace links with stale status.
func (e *RuleEngine) ruleStaleTraces(sev Severity) *RuleResult {
	gaps := e.analyzer.AnalyzeGaps()
	result := &RuleResult{RuleID: "stale-traces", Severity: sev}
	for _, link := range gaps.StaleTraces {
		result.Violations = append(result.Violations, Violation{
			Rule:     "stale-traces",
			Severity: sev,
			Message:  fmt.Sprintf("stale trace link: %s → %s (%s)", link.FromID, link.ToID, link.LinkType),
		})
	}
	return result
}

// ruleAllReqsMustHaveASPICE flags requirements without aspice= attribute.
func (e *RuleEngine) ruleAllReqsMustHaveASPICE(sev Severity) *RuleResult {
	g := e.analyzer.GetGraph()
	result := &RuleResult{RuleID: "all-reqs-must-have-aspice", Severity: sev}
	for _, req := range g.Requirements {
		if req.ASPICE == "" {
			result.Violations = append(result.Violations, Violation{
				Rule:     "all-reqs-must-have-aspice",
				Severity: sev,
				Message:  fmt.Sprintf("requirement %s has no aspice= attribute", req.ID),
				Location: fmt.Sprintf("%s:%d", req.FilePath, req.LineNumber),
			})
		}
	}
	return result
}

// ruleAllReqsMustHavePriority flags requirements without priority= attribute.
func (e *RuleEngine) ruleAllReqsMustHavePriority(sev Severity) *RuleResult {
	g := e.analyzer.GetGraph()
	result := &RuleResult{RuleID: "all-reqs-must-have-priority", Severity: sev}
	for _, req := range g.Requirements {
		if req.Priority == "" {
			result.Violations = append(result.Violations, Violation{
				Rule:     "all-reqs-must-have-priority",
				Severity: sev,
				Message:  fmt.Sprintf("requirement %s has no priority= attribute", req.ID),
				Location: fmt.Sprintf("%s:%d", req.FilePath, req.LineNumber),
			})
		}
	}
	return result
}

// ruleMaxOrphanPercentage flags if orphan requirements exceed a threshold.
// Threshold is read from config.RuleParams["max-orphan-percentage"] (default: 20).
func (e *RuleEngine) ruleMaxOrphanPercentage(sev Severity) *RuleResult {
	result := &RuleResult{RuleID: "max-orphan-percentage", Severity: sev}
	g := e.analyzer.GetGraph()
	if len(g.Requirements) == 0 {
		return result
	}

	threshold := 20 // default 20%
	if e.config.RuleParams != nil {
		if t, ok := e.config.RuleParams["max-orphan-percentage"]; ok && t > 0 {
			threshold = t
		}
	}

	gaps := e.analyzer.AnalyzeGaps()
	orphanPct := float64(len(gaps.OrphanRequirements)) / float64(len(g.Requirements)) * 100

	if orphanPct > float64(threshold) {
		result.Violations = append(result.Violations, Violation{
			Rule:     "max-orphan-percentage",
			Severity: sev,
			Message: fmt.Sprintf(
				"orphan requirements: %.1f%% exceeds threshold of %d%%  (%d/%d untraced)",
				orphanPct, threshold, len(gaps.OrphanRequirements), len(g.Requirements),
			),
		})
	}
	return result
}

// ruleImplFileNotFound flags arch/design elements whose impl= value references a file that does not exist on disk.
// Only file-based impl refs are checked (those containing a '.' in the file name component).
// Pure Go package paths (no extension) are skipped.
func (e *RuleEngine) ruleImplFileNotFound(sev Severity) *RuleResult {
	g := e.analyzer.GetGraph()
	result := &RuleResult{RuleID: "impl-file-not-found", Severity: sev}
	check := func(implVal, id, filePath string, line int) {
		if implVal == "" {
			return
		}
		ref := model.ParseImplRef(implVal)
		if !ref.IsFile {
			return
		}
		if _, err := os.Stat(ref.FilePath); os.IsNotExist(err) {
			result.Violations = append(result.Violations, Violation{
				Rule:     "impl-file-not-found",
				Severity: sev,
				Message:  fmt.Sprintf("%s: impl= references non-existent file %q", id, ref.FilePath),
				Location: fmt.Sprintf("%s:%d", filePath, line),
			})
		}
	}
	for _, arch := range g.ArchElements {
		check(arch.Impl, arch.ID, arch.FilePath, arch.LineNumber)
	}
	for _, dsn := range g.DesignElements {
		check(dsn.Impl, dsn.ID, dsn.FilePath, dsn.LineNumber)
	}
	return result
}

// buildTestSpecArchIndex returns a set of arch IDs that have at least one test spec referencing them.
// O(|TestSpecs| × avg|Arch|) — called once per ruleMissingTestSpec invocation.
func buildTestSpecArchIndex(g *model.TraceabilityGraph) map[string]struct{} {
	idx := make(map[string]struct{})
	for _, spec := range g.TestSpecs {
		for _, aid := range spec.Arch {
			idx[aid] = struct{}{}
		}
	}
	return idx
}
