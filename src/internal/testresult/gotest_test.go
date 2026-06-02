package testresult

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleGoTest = `{"Time":"2026-03-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg","Test":"TestFoo"}
{"Time":"2026-03-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Test":"TestFoo","Output":"--- PASS: TestFoo\n"}
{"Time":"2026-03-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg","Test":"TestFoo","Elapsed":0.5}
{"Time":"2026-03-15T10:00:02Z","Action":"run","Package":"github.com/example/pkg","Test":"TestBar"}
{"Time":"2026-03-15T10:00:03Z","Action":"fail","Package":"github.com/example/pkg","Test":"TestBar","Elapsed":0.3}
`

// [test-spec,id=TS-TR-010,req="REQ-TESTING-001",dsn=comp.testresult.gotest,aspice=SWE.4-BP4]
// TestParseGoTest verifies that go test -json output is parsed correctly.
func TestParseGoTest(t *testing.T) {
	f := writeTempJSON(t, sampleGoTest)
	results, err := ParseGoTest(f, "proj", "linux")
	if err != nil {
		t.Fatalf("ParseGoTest error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

// [test-spec,id=TS-TR-011,req="REQ-TESTING-001",dsn=comp.testresult.gotest,aspice="SWE.5.BP3"]
// TestParseGoTest_Statuses verifies that pass/fail statuses are correctly assigned.
func TestParseGoTest_Statuses(t *testing.T) {
	f := writeTempJSON(t, sampleGoTest)
	results, _ := ParseGoTest(f, "proj", "linux")
	statusMap := make(map[string]string)
	for _, r := range results {
		statusMap[r.TestName] = r.Status
	}
	if statusMap["TestFoo"] != "passed" {
		t.Errorf("TestFoo status = %q, want passed", statusMap["TestFoo"])
	}
	if statusMap["TestBar"] != "failed" {
		t.Errorf("TestBar status = %q, want failed", statusMap["TestBar"])
	}
}

// [test-spec,id=TS-TR-012,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseGoTest_Elapsed verifies that elapsed time is captured from pass/fail events.
func TestParseGoTest_Elapsed(t *testing.T) {
	f := writeTempJSON(t, sampleGoTest)
	results, _ := ParseGoTest(f, "proj", "linux")
	for _, r := range results {
		if r.TestName == "TestFoo" && r.Duration == 0 {
			t.Error("expected non-zero duration for TestFoo")
		}
	}
}

// [test-spec,id=TS-TR-013,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseGoTest_Metadata verifies project, platform and package fields.
func TestParseGoTest_Metadata(t *testing.T) {
	f := writeTempJSON(t, sampleGoTest)
	results, _ := ParseGoTest(f, "testproj", "macos")
	for _, r := range results {
		if r.Project != "testproj" {
			t.Errorf("Project = %q, want testproj", r.Project)
		}
		if r.Platform != "macos" {
			t.Errorf("Platform = %q, want macos", r.Platform)
		}
	}
}

// [test-spec,id=TS-TR-014,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseGoTest_FileNotFound verifies error for missing file.
func TestParseGoTest_FileNotFound(t *testing.T) {
	_, err := ParseGoTest("/nonexistent/file.json", "proj", "linux")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// [test-spec,id=TS-TR-015,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseGoTest_MalformedLines verifies that malformed JSON lines are skipped.
func TestParseGoTest_MalformedLines(t *testing.T) {
	content := `not valid json
{"Action":"pass","Package":"pkg","Test":"TestOK","Elapsed":0.1}
also not json
`
	f := writeTempJSON(t, content)
	results, err := ParseGoTest(f, "proj", "linux")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// TestOK should still be parsed
	if len(results) != 1 {
		t.Errorf("expected 1 result (malformed lines skipped), got %d", len(results))
	}
}

// [test-spec,id=TS-TR-016,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseGoTest_OutputCaptured verifies that output lines are concatenated.
func TestParseGoTest_OutputCaptured(t *testing.T) {
	f := writeTempJSON(t, sampleGoTest)
	results, _ := ParseGoTest(f, "proj", "linux")
	for _, r := range results {
		if r.TestName == "TestFoo" && r.Stdout == "" {
			t.Error("expected stdout captured for TestFoo")
		}
	}
}

// [test-spec,id=TS-TR-017,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseGoTest_EmptyLines verifies that empty lines are skipped without error.
func TestParseGoTest_EmptyLines(t *testing.T) {
	content := `
{"Action":"pass","Package":"pkg","Test":"TestA","Elapsed":0.1}

`
	f := writeTempJSON(t, content)
	results, err := ParseGoTest(f, "proj", "linux")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// [test-spec,id=TS-TR-023,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseGoTest_ShortID verifies that IDs use short package name (last segment only, no platform prefix).
func TestParseGoTest_ShortID(t *testing.T) {
	f := writeTempJSON(t, sampleGoTest)
	results, _ := ParseGoTest(f, "proj", "linux")
	for _, r := range results {
		want := "pkg::" + r.TestName
		if r.ID != want {
			t.Errorf("ID = %q, want %q", r.ID, want)
		}
	}
}

func writeTempJSON(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "results.json")
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}
