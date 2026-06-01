package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

const defaultTimeout = 30 * time.Second

// Request is sent to every plugin via stdin.
type Request struct {
	File    string `json:"file,omitempty"`
	Project string `json:"project"`
}

// EnrichmentPatch is returned by enricher plugins.
type EnrichmentPatch struct {
	ElementID string `json:"element_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

// ValidationIssue is returned by validator plugins.
type ValidationIssue struct {
	Rule      string `json:"rule"`
	Severity  string `json:"severity"`
	ElementID string `json:"element_id"`
	Message   string `json:"message"`
}

// PluginTraceLink is returned by linker plugins.
type PluginTraceLink struct {
	FromID   string `json:"from_id"`
	FromType string `json:"from_type"`
	ToID     string `json:"to_id"`
	ToType   string `json:"to_type"`
	LinkType string `json:"link_type"`
	Reason   string `json:"reason"`
}

// PluginTestResult is returned by test-result-parser plugins.
type PluginTestResult struct {
	ID       string  `json:"id"`
	TestName string  `json:"test_name"`
	Package  string  `json:"package"`
	Status   string  `json:"status"`
	Duration float64 `json:"duration"`
}

// PluginRequirement is returned by requirement-parser plugins.
type PluginRequirement struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Priority string `json:"priority"`
	Status   string `json:"status"`
	ASPICE   string `json:"aspice"`
}

// PluginArchElement is returned by arch-parser plugins.
type PluginArchElement struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// PluginTestSpec is returned by test-spec-parser plugins.
type PluginTestSpec struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Req   []string `json:"req"`
	ASPICE string  `json:"aspice"`
}

// PluginDesignElement is returned by design-parser plugins.
type PluginDesignElement struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Arch  string `json:"arch"`
}

// ReporterRequest is sent to reporter/diagram/heatmap plugins.
type ReporterRequest struct {
	Project   string                  `json:"project"`
	OutputDir string                  `json:"output_dir"`
	Graph     *model.TraceabilityGraph `json:"graph"`
}

// ReporterResponse is returned by reporter plugins.
type ReporterResponse struct {
	Files []string `json:"files"`
}

// run executes a plugin binary, sends reqJSON on stdin, returns stdout bytes.
func run(ctx context.Context, pluginPath string, reqJSON []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, pluginPath)
	cmd.Stdin = bytes.NewReader(reqJSON)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := stderr.String()
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("plugin %s: %s", pluginPath, msg)
	}
	return out.Bytes(), nil
}

func runWithTimeout(pluginPath string, reqJSON []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return run(ctx, pluginPath, reqJSON)
}

// RunTestResultParser calls a test-result-parser plugin and returns TestResults.
func RunTestResultParser(pluginPath, filePath, project string) ([]*model.TestResult, error) {
	req := Request{File: filePath, Project: project}
	reqJSON, _ := json.Marshal(req)

	out, err := runWithTimeout(pluginPath, reqJSON)
	if err != nil {
		return nil, err
	}

	var raw []PluginTestResult
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("plugin %s: invalid response: %w", pluginPath, err)
	}

	results := make([]*model.TestResult, 0, len(raw))
	for _, r := range raw {
		id := r.ID
		if id == "" {
			id = r.Package + "::" + r.TestName
		}
		results = append(results, &model.TestResult{
			ID:       id,
			TestName: r.TestName,
			Package:  r.Package,
			Status:   r.Status,
			Duration: r.Duration,
			Project:  project,
		})
	}
	return results, nil
}

// RunRequirementParser calls a requirement-parser plugin and returns Requirements.
func RunRequirementParser(pluginPath, filePath, project string) ([]*model.Requirement, error) {
	req := Request{File: filePath, Project: project}
	reqJSON, _ := json.Marshal(req)

	out, err := runWithTimeout(pluginPath, reqJSON)
	if err != nil {
		return nil, err
	}

	var raw []PluginRequirement
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("plugin %s: invalid response: %w", pluginPath, err)
	}

	reqs := make([]*model.Requirement, 0, len(raw))
	for _, r := range raw {
		reqs = append(reqs, &model.Requirement{
			ID:       r.ID,
			Title:    r.Title,
			Priority: r.Priority,
			Status:   r.Status,
			ASPICE:   r.ASPICE,
			Project:  project,
		})
	}
	return reqs, nil
}

// RunArchParser calls an arch-parser plugin and returns ArchElements.
func RunArchParser(pluginPath, filePath, project string) ([]*model.ArchElement, error) {
	req := Request{File: filePath, Project: project}
	reqJSON, _ := json.Marshal(req)

	out, err := runWithTimeout(pluginPath, reqJSON)
	if err != nil {
		return nil, err
	}

	var raw []PluginArchElement
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("plugin %s: invalid response: %w", pluginPath, err)
	}

	elems := make([]*model.ArchElement, 0, len(raw))
	for _, r := range raw {
		elems = append(elems, &model.ArchElement{
			ID:      r.ID,
			Title:   r.Title,
			Project: project,
		})
	}
	return elems, nil
}

// RunTestSpecParser calls a test-spec-parser plugin and returns TestSpecs.
func RunTestSpecParser(pluginPath, filePath, project string) ([]*model.TestSpec, error) {
	req := Request{File: filePath, Project: project}
	reqJSON, _ := json.Marshal(req)

	out, err := runWithTimeout(pluginPath, reqJSON)
	if err != nil {
		return nil, err
	}

	var raw []PluginTestSpec
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("plugin %s: invalid response: %w", pluginPath, err)
	}

	specs := make([]*model.TestSpec, 0, len(raw))
	for _, r := range raw {
		specs = append(specs, &model.TestSpec{
			ID:      r.ID,
			Title:   r.Title,
			Req:     r.Req,
			Project: project,
		})
	}
	return specs, nil
}

// RunDesignParser calls a design-parser plugin and returns DesignElements.
func RunDesignParser(pluginPath, filePath, project string) ([]*model.DesignElement, error) {
	req := Request{File: filePath, Project: project}
	reqJSON, _ := json.Marshal(req)

	out, err := runWithTimeout(pluginPath, reqJSON)
	if err != nil {
		return nil, err
	}

	var raw []PluginDesignElement
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("plugin %s: invalid response: %w", pluginPath, err)
	}

	elems := make([]*model.DesignElement, 0, len(raw))
	for _, r := range raw {
		elems = append(elems, &model.DesignElement{
			ID:      r.ID,
			Title:   r.Title,
			Project: project,
		})
	}
	return elems, nil
}

// RunEnricher calls an enricher plugin and returns enrichment patches.
func RunEnricher(pluginPath string, g *model.TraceabilityGraph) ([]EnrichmentPatch, error) {
	reqJSON, _ := json.Marshal(g)

	out, err := runWithTimeout(pluginPath, reqJSON)
	if err != nil {
		return nil, err
	}

	var patches []EnrichmentPatch
	if err := json.Unmarshal(out, &patches); err != nil {
		return nil, fmt.Errorf("plugin %s: invalid response: %w", pluginPath, err)
	}
	return patches, nil
}

// ApplyEnrichments applies enrichment patches to the graph metadata.
func ApplyEnrichments(g *model.TraceabilityGraph, patches []EnrichmentPatch) {
	for _, p := range patches {
		if req, ok := g.Requirements[p.ElementID]; ok {
			if req.Attributes == nil {
				req.Attributes = make(map[string]string)
			}
			req.Attributes[p.Key] = p.Value
			continue
		}
		if arch, ok := g.ArchElements[p.ElementID]; ok {
			if arch.Attributes == nil {
				arch.Attributes = make(map[string]string)
			}
			arch.Attributes[p.Key] = p.Value
			continue
		}
		if spec, ok := g.TestSpecs[p.ElementID]; ok {
			if spec.Attributes == nil {
				spec.Attributes = make(map[string]string)
			}
			spec.Attributes[p.Key] = p.Value
		}
	}
}

// RunValidator calls a validator plugin and returns validation issues.
func RunValidator(pluginPath string, g *model.TraceabilityGraph) ([]ValidationIssue, error) {
	reqJSON, _ := json.Marshal(g)

	out, err := runWithTimeout(pluginPath, reqJSON)
	if err != nil {
		return nil, err
	}

	var issues []ValidationIssue
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, fmt.Errorf("plugin %s: invalid response: %w", pluginPath, err)
	}
	return issues, nil
}

// RunLinker calls a linker plugin and returns trace links.
func RunLinker(pluginPath string, g *model.TraceabilityGraph) ([]*model.TraceLink, error) {
	reqJSON, _ := json.Marshal(g)

	out, err := runWithTimeout(pluginPath, reqJSON)
	if err != nil {
		return nil, err
	}

	var raw []PluginTraceLink
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("plugin %s: invalid response: %w", pluginPath, err)
	}

	links := make([]*model.TraceLink, 0, len(raw))
	for _, r := range raw {
		links = append(links, &model.TraceLink{
			FromID:   r.FromID,
			FromType: r.FromType,
			ToID:     r.ToID,
			ToType:   r.ToType,
			LinkType: r.LinkType,
			Status:   "active",
			Reason:   r.Reason,
		})
	}
	return links, nil
}

// RunReporter calls a reporter plugin to generate output files.
func RunReporter(pluginPath, outputDir, project string, g *model.TraceabilityGraph) ([]string, error) {
	req := ReporterRequest{Project: project, OutputDir: outputDir, Graph: g}
	reqJSON, _ := json.Marshal(req)

	out, err := runWithTimeout(pluginPath, reqJSON)
	if err != nil {
		return nil, err
	}

	var resp ReporterResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("plugin %s: invalid response: %w", pluginPath, err)
	}
	return resp.Files, nil
}

// RunNotifier calls a notifier plugin with the gaps data.
func RunNotifier(pluginPath string, payload any) error {
	reqJSON, _ := json.Marshal(payload)
	_, err := runWithTimeout(pluginPath, reqJSON)
	return err
}
