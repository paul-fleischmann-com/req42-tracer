package graph

import (
	"fmt"
	"strings"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

// Builder constructs the traceability graph from parsed artifacts.
type Builder struct {
	graph    *model.TraceabilityGraph
	linkSeen map[string]struct{} // dedup key → prevents O(n²) scan in addLink
}

// NewBuilder creates a new graph builder.
func NewBuilder() *Builder {
	return &Builder{
		graph: &model.TraceabilityGraph{
			Requirements:   make(map[string]*model.Requirement),
			ArchElements:   make(map[string]*model.ArchElement),
			DesignElements: make(map[string]*model.DesignElement),
			TestSpecs:      make(map[string]*model.TestSpec),
			TestCodes:      make(map[string]*model.TestCode),
			TestResults:    make(map[string]*model.TestResult),
			Links:          []*model.TraceLink{},
		},
		linkSeen: make(map[string]struct{}),
	}
}

// MergeGraph merges another graph into this one.
func (b *Builder) MergeGraph(other *model.TraceabilityGraph) error {
	if other == nil {
		return nil
	}

	// Merge all artifact types
	for id, req := range other.Requirements {
		if _, exists := b.graph.Requirements[id]; exists {
			return fmt.Errorf("duplicate requirement ID: %s", id)
		}
		b.graph.Requirements[id] = req
	}

	for id, arch := range other.ArchElements {
		if _, exists := b.graph.ArchElements[id]; exists {
			return fmt.Errorf("duplicate architecture element ID: %s", id)
		}
		b.graph.ArchElements[id] = arch
	}

	for id, dsn := range other.DesignElements {
		if _, exists := b.graph.DesignElements[id]; exists {
			return fmt.Errorf("duplicate design element ID: %s", id)
		}
		b.graph.DesignElements[id] = dsn
	}

	// TestSpecs and TestCodes from Go inline annotations may share IDs with adoc-declared
	// specs (same ID, different source). Skip duplicates so one conflict can't block all
	// TestCode linking; accumulate warnings to return to the caller.
	var dupWarnings []string
	for id, spec := range other.TestSpecs {
		if _, exists := b.graph.TestSpecs[id]; exists {
			dupWarnings = append(dupWarnings, "duplicate test-spec ID: "+id)
			continue
		}
		b.graph.TestSpecs[id] = spec
	}

	for id, code := range other.TestCodes {
		if _, exists := b.graph.TestCodes[id]; exists {
			dupWarnings = append(dupWarnings, "duplicate test-code ID: "+id)
			continue
		}
		b.graph.TestCodes[id] = code
	}

	for id, result := range other.TestResults {
		if _, exists := b.graph.TestResults[id]; exists {
			return fmt.Errorf("duplicate test-result ID: %s", id)
		}
		b.graph.TestResults[id] = result
	}

	// Register merged links in linkSeen so BuildLinks() treats them as existing.
	for _, link := range other.Links {
		key := link.FromType + ":" + link.FromID + "->" + link.ToType + ":" + link.ToID
		b.linkSeen[key] = struct{}{}
	}
	b.graph.Links = append(b.graph.Links, other.Links...)

	if len(dupWarnings) > 0 {
		return fmt.Errorf("%s", strings.Join(dupWarnings, "; "))
	}
	return nil
}

// BuildLinks constructs trace links between artifacts based on explicit attributes and references.
func (b *Builder) BuildLinks() error {
	// Link requirements to architecture
	for archID, arch := range b.graph.ArchElements {
		for _, reqID := range arch.Req {
			if _, exists := b.graph.Requirements[reqID]; !exists {
				continue // Skip non-existent requirements (will be flagged as error elsewhere)
			}

			b.addLink(&model.TraceLink{
				FromID:   reqID,
				FromType: "requirement",
				ToID:     archID,
				ToType:   "arch",
				LinkType: "satisfied-by",
				Status:   "active",
				Reason:   "Explicit ref= attribute",
			})
		}
	}

	// Link arch to design elements (SWE.2 → SWE.3)
	for dsnID, dsn := range b.graph.DesignElements {
		if dsn.Arch == "" {
			continue
		}
		if _, exists := b.graph.ArchElements[dsn.Arch]; !exists {
			continue
		}
		b.addLink(&model.TraceLink{
			FromID:   dsn.Arch,
			FromType: "arch",
			ToID:     dsnID,
			ToType:   "design",
			LinkType: "refined-by",
			Status:   "active",
			Reason:   "Explicit arch= attribute on dsn",
		})
	}

	// Link architecture to test specs (SWE.5)
	for specID, spec := range b.graph.TestSpecs {
		for _, archID := range spec.Arch {
			if _, exists := b.graph.ArchElements[archID]; !exists {
				continue
			}

			b.addLink(&model.TraceLink{
				FromID:   archID,
				FromType: "arch",
				ToID:     specID,
				ToType:   "test-spec",
				LinkType: "verified-by",
				Status:   "active",
				Reason:   "Explicit arch= attribute",
			})
		}
	}

	// Link design elements to test specs (SWE.4)
	for specID, spec := range b.graph.TestSpecs {
		for _, dsnID := range spec.Dsn {
			if _, exists := b.graph.DesignElements[dsnID]; !exists {
				continue
			}
			b.addLink(&model.TraceLink{
				FromID:   dsnID,
				FromType: "design",
				ToID:     specID,
				ToType:   "test-spec",
				LinkType: "verified-by",
				Status:   "active",
				Reason:   "Explicit dsn= attribute",
			})
		}
	}

	// Link test specs to requirements
	for specID, spec := range b.graph.TestSpecs {
		for _, reqID := range spec.Req {
			if _, exists := b.graph.Requirements[reqID]; !exists {
				continue
			}

			b.addLink(&model.TraceLink{
				FromID:   reqID,
				FromType: "requirement",
				ToID:     specID,
				ToType:   "test-spec",
				LinkType: "verified-by",
				Status:   "active",
				Reason:   "Explicit req= attribute",
			})
		}
	}

	// Link test codes to test specs
	for codeID, code := range b.graph.TestCodes {
		if code.TestSpec != "" {
			if _, exists := b.graph.TestSpecs[code.TestSpec]; !exists {
				continue
			}

			b.addLink(&model.TraceLink{
				FromID:   code.TestSpec,
				FromType: "test-spec",
				ToID:     codeID,
				ToType:   "test-code",
				LinkType: "implements",
				Status:   "active",
				Reason:   "Explicit test-spec= attribute",
			})
		}
	}

	// Link test results to test codes (via name matching as fallback)
	for resultID, result := range b.graph.TestResults {
		if result.LinkedCode == "" && result.LinkedSpec == "" {
			// Try name-based matching: test name -> TestCode function name
			for codeID, code := range b.graph.TestCodes {
				if nameMatches(result.TestName, code.Function) {
					result.LinkedCode = codeID
					b.addLink(&model.TraceLink{
						FromID:   codeID,
						FromType: "test-code",
						ToID:     resultID,
						ToType:   "test-result",
						LinkType: "produces",
						Status:   "active",
						Reason:   "Name-based matching",
					})
					break
				}
			}
		}

		// If still unlinked, try spec matching
		if result.LinkedSpec == "" {
			for specID := range b.graph.TestSpecs {
				// Check if test name matches spec ID pattern
				if strings.Contains(strings.ToLower(result.TestName), strings.ToLower(specID)) {
					result.LinkedSpec = specID
					b.addLink(&model.TraceLink{
						FromID:   specID,
						FromType: "test-spec",
						ToID:     resultID,
						ToType:   "test-result",
						LinkType: "produces",
						Status:   "active",
						Reason:   "Name-based spec matching",
					})
					break
				}
			}
		}
	}

	return nil
}

// addLink adds a trace link, avoiding duplicates via O(1) map lookup.
func (b *Builder) addLink(link *model.TraceLink) {
	key := link.FromType + ":" + link.FromID + "->" + link.ToType + ":" + link.ToID
	if _, exists := b.linkSeen[key]; exists {
		return
	}
	b.linkSeen[key] = struct{}{}
	b.graph.Links = append(b.graph.Links, link)
}

// DeriveASPICELevels automatically derives ASPICE levels from ARC42 hierarchy.
func (b *Builder) DeriveASPICELevels() {
	for _, arch := range b.graph.ArchElements {
		if arch.ASPICE != "" {
			continue // Already explicitly set
		}

		if arch.Parent == "" {
			// Level 1 (no parent) -> SWE.2
			arch.ASPICE = "SWE.2"
		} else {
			// Level 2+ (has parent) -> SWE.3
			arch.ASPICE = "SWE.3"
		}
	}
}

// ApplyImplRefs sets the Impl field on existing arch elements from a map of
// archID → implRef without creating new elements or triggering duplicate errors.
// Only updates elements whose Impl is currently empty (adoc-declared impl wins).
func (b *Builder) ApplyImplRefs(refs map[string]string) {
	for archID, implRef := range refs {
		if arch, exists := b.graph.ArchElements[archID]; exists {
			if arch.Impl == "" {
				arch.Impl = implRef
			}
		}
	}
}

// GetGraph returns the built traceability graph.
func (b *Builder) GetGraph() *model.TraceabilityGraph {
	return b.graph
}

// nameMatches checks if a test result name matches a test code function name.
// Example: "TestAPIAuth" matches function "TestAPIAuth" or "test_api_auth"
func nameMatches(resultName, funcName string) bool {
	// Normalize both to lowercase for comparison
	rLower := strings.ToLower(resultName)
	fLower := strings.ToLower(funcName)

	// Exact match
	if rLower == fLower {
		return true
	}

	// Remove common prefixes/suffixes
	rNorm := strings.TrimPrefix(rLower, "test")
	rNorm = strings.TrimPrefix(rNorm, "test_")
	fNorm := strings.TrimPrefix(fLower, "test")
	fNorm = strings.TrimPrefix(fNorm, "test_")

	// Convert underscores to empty for comparison
	rNorm = strings.ReplaceAll(rNorm, "_", "")
	fNorm = strings.ReplaceAll(fNorm, "_", "")

	return rNorm != "" && rNorm == fNorm
}
