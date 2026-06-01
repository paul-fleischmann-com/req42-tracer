package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

// writeMockPlugin writes an executable shell script that prints fixedOutput to stdout.
func writeMockPlugin(t *testing.T, fixedOutput string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "mock-plugin")
	script := "#!/bin/sh\ncat <<'EOF'\n" + fixedOutput + "\nEOF\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

// writeMockPluginFailing writes a plugin that exits with code 1.
func writeMockPluginFailing(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "fail-plugin")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho 'error' >&2\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

// [test-spec,id=TS-PLUG-001,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunTestResultParser verifies a test-result-parser plugin is called and results parsed.
func TestRunTestResultParser(t *testing.T) {
	output := `[{"id":"pkg::TestFoo","test_name":"TestFoo","package":"pkg","status":"passed","duration":0.042}]`
	plugin := writeMockPlugin(t, output)

	results, err := RunTestResultParser(plugin, "dummy.xml", "proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].TestName != "TestFoo" {
		t.Errorf("test_name = %q, want TestFoo", results[0].TestName)
	}
	if results[0].Status != "passed" {
		t.Errorf("status = %q, want passed", results[0].Status)
	}
	if results[0].Project != "proj" {
		t.Errorf("project = %q, want proj", results[0].Project)
	}
}

// [test-spec,id=TS-PLUG-002,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunTestResultParser_PluginFails verifies error is returned when plugin exits non-zero.
func TestRunTestResultParser_PluginFails(t *testing.T) {
	plugin := writeMockPluginFailing(t)
	_, err := RunTestResultParser(plugin, "dummy.xml", "proj")
	if err == nil {
		t.Error("expected error for failing plugin")
	}
}

// [test-spec,id=TS-PLUG-003,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunRequirementParser verifies a requirement-parser plugin is called and results parsed.
func TestRunRequirementParser(t *testing.T) {
	output := `[{"id":"REQ-001","title":"System shall work","priority":"high","status":"approved","aspice":"SWE.1"}]`
	plugin := writeMockPlugin(t, output)

	reqs, err := RunRequirementParser(plugin, "reqs.md", "proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	if reqs[0].ID != "REQ-001" {
		t.Errorf("id = %q, want REQ-001", reqs[0].ID)
	}
	if reqs[0].Priority != "high" {
		t.Errorf("priority = %q, want high", reqs[0].Priority)
	}
}

// [test-spec,id=TS-PLUG-004,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunArchParser verifies an arch-parser plugin is called and results parsed.
func TestRunArchParser(t *testing.T) {
	output := `[{"id":"arch.parser","title":"Parser Component","type":"component"}]`
	plugin := writeMockPlugin(t, output)

	elems, err := RunArchParser(plugin, "arch.puml", "proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].ID != "arch.parser" {
		t.Errorf("id = %q, want arch.parser", elems[0].ID)
	}
}

// [test-spec,id=TS-PLUG-005,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunTestSpecParser verifies a test-spec-parser plugin is called and results parsed.
func TestRunTestSpecParser(t *testing.T) {
	output := `[{"id":"TS-EXT-001","title":"External Test","req":["REQ-001"],"aspice":"SWE.4-BP2"}]`
	plugin := writeMockPlugin(t, output)

	specs, err := RunTestSpecParser(plugin, "testrail.json", "proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}
	if specs[0].ID != "TS-EXT-001" {
		t.Errorf("id = %q, want TS-EXT-001", specs[0].ID)
	}
	if len(specs[0].Req) != 1 || specs[0].Req[0] != "REQ-001" {
		t.Errorf("req = %v, want [REQ-001]", specs[0].Req)
	}
}

// [test-spec,id=TS-PLUG-006,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunDesignParser verifies a design-parser plugin is called and results parsed.
func TestRunDesignParser(t *testing.T) {
	output := `[{"id":"dsn.auth","title":"Auth Module","arch":"arch.core"}]`
	plugin := writeMockPlugin(t, output)

	elems, err := RunDesignParser(plugin, "design.xml", "proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].ID != "dsn.auth" {
		t.Errorf("id = %q, want dsn.auth", elems[0].ID)
	}
}

// [test-spec,id=TS-PLUG-007,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunEnricher verifies an enricher plugin applies patches to graph elements.
func TestRunEnricher(t *testing.T) {
	output := `[{"element_id":"REQ-001","key":"jira_status","value":"In Progress"}]`
	plugin := writeMockPlugin(t, output)

	g := model.EmptyGraph()
	g.Requirements["REQ-001"] = &model.Requirement{ID: "REQ-001", Attributes: make(map[string]string)}

	patches, err := RunEnricher(plugin, g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	ApplyEnrichments(g, patches)
	if g.Requirements["REQ-001"].Attributes["jira_status"] != "In Progress" {
		t.Errorf("attribute not applied, got %q", g.Requirements["REQ-001"].Attributes["jira_status"])
	}
}

// [test-spec,id=TS-PLUG-008,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunValidator verifies a validator plugin returns validation issues.
func TestRunValidator(t *testing.T) {
	output := `[{"rule":"naming","severity":"error","element_id":"REQ-001","message":"ID must start with SYS-"}]`
	plugin := writeMockPlugin(t, output)

	g := model.EmptyGraph()
	issues, err := RunValidator(plugin, g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != "error" {
		t.Errorf("severity = %q, want error", issues[0].Severity)
	}
}

// [test-spec,id=TS-PLUG-009,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunLinker verifies a linker plugin returns trace links.
func TestRunLinker(t *testing.T) {
	output := `[{"from_id":"TestSafety","from_type":"test-result","to_id":"REQ-SAFETY","to_type":"requirement","link_type":"verifies","reason":"keyword"}]`
	plugin := writeMockPlugin(t, output)

	g := model.EmptyGraph()
	links, err := RunLinker(plugin, g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].FromID != "TestSafety" {
		t.Errorf("from_id = %q, want TestSafety", links[0].FromID)
	}
	if links[0].Status != "active" {
		t.Errorf("status = %q, want active", links[0].Status)
	}
}

// [test-spec,id=TS-PLUG-010,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunReporter verifies a reporter plugin returns generated file paths.
func TestRunReporter(t *testing.T) {
	output := `{"files":["reports/matrix.docx"]}`
	plugin := writeMockPlugin(t, output)

	g := model.EmptyGraph()
	files, err := RunReporter(plugin, "reports/", "proj", g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0] != "reports/matrix.docx" {
		t.Errorf("files = %v, want [reports/matrix.docx]", files)
	}
}

// [test-spec,id=TS-PLUG-011,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunNotifier verifies a notifier plugin is called without error.
func TestRunNotifier(t *testing.T) {
	plugin := writeMockPlugin(t, "")
	payload := map[string]string{"event": "gaps-found"}
	if err := RunNotifier(plugin, payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// [test-spec,id=TS-PLUG-012,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunTestResultParser_InvalidJSON verifies error for malformed plugin output.
func TestRunTestResultParser_InvalidJSON(t *testing.T) {
	plugin := writeMockPlugin(t, "not valid json")
	_, err := RunTestResultParser(plugin, "dummy.xml", "proj")
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

// [test-spec,id=TS-PLUG-013,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunTestResultParser_AutoID verifies ID is auto-generated from package+test_name when empty.
func TestRunTestResultParser_AutoID(t *testing.T) {
	output := `[{"test_name":"TestBar","package":"mypkg","status":"failed"}]`
	plugin := writeMockPlugin(t, output)

	results, err := RunTestResultParser(plugin, "dummy.xml", "proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].ID != "mypkg::TestBar" {
		t.Errorf("auto-id = %q, want mypkg::TestBar", results[0].ID)
	}
}

// [test-spec,id=TS-PLUG-014,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestApplyEnrichments_ArchElement verifies patches are applied to arch elements.
func TestApplyEnrichments_ArchElement(t *testing.T) {
	g := model.EmptyGraph()
	g.ArchElements["arch.core"] = &model.ArchElement{ID: "arch.core", Attributes: make(map[string]string)}

	ApplyEnrichments(g, []EnrichmentPatch{
		{ElementID: "arch.core", Key: "ci_coverage", Value: "82.3%"},
	})

	if g.ArchElements["arch.core"].Attributes["ci_coverage"] != "82.3%" {
		t.Error("arch element attribute not applied")
	}
}

// [test-spec,id=TS-PLUG-015,req="REQ-PLUGIN-001",aspice="SWE.4.BP2"]
// TestRunTestResultParser_PluginNotFound verifies error when plugin binary does not exist.
func TestRunTestResultParser_PluginNotFound(t *testing.T) {
	_, err := RunTestResultParser("/nonexistent/plugin", "dummy.xml", "proj")
	if err == nil {
		t.Error("expected error for missing plugin binary")
	}
}

// helper: ensure Request is correctly serialised (used by all runners)
func TestRequestSerialization(t *testing.T) {
	req := Request{File: "test.xml", Project: "proj"}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var out Request
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.File != req.File || out.Project != req.Project {
		t.Errorf("roundtrip mismatch: %+v", out)
	}
}
