package testresult

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleJUnit = `<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="MyTests" package="com.example" tests="3" skipped="1" failures="1" errors="0" time="1.234" timestamp="2026-03-15T10:00:00Z">
  <testcase name="TestPass" classname="com.example.MyTests" time="0.5">
    <system-out>some output</system-out>
  </testcase>
  <testcase name="TestFail" classname="com.example.MyTests" time="0.3">
    <failure message="assertion failed">expected 1 got 2</failure>
  </testcase>
  <testcase name="TestSkip" classname="com.example.MyTests" time="0.0">
    <skipped message="not implemented"/>
  </testcase>
</testsuite>`

// [test-spec,id=TS-TR-001,req="REQ-TESTING-001",dsn=comp.testresult.junit,aspice=SWE.4-BP4]
// TestParseJUnit verifies that JUnit XML is correctly parsed into TestResult objects.
func TestParseJUnit(t *testing.T) {
	f := writeTempXML(t, sampleJUnit)
	results, err := ParseJUnit(f, "proj", "linux")
	if err != nil {
		t.Fatalf("ParseJUnit error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

// [test-spec,id=TS-TR-002,req="REQ-TESTING-001",dsn=comp.testresult.junit,aspice="SWE.5.BP3"]
// TestParseJUnit_Statuses verifies that pass/fail/skipped statuses are correctly assigned.
func TestParseJUnit_Statuses(t *testing.T) {
	f := writeTempXML(t, sampleJUnit)
	results, _ := ParseJUnit(f, "proj", "linux")

	statusMap := make(map[string]string)
	for _, r := range results {
		statusMap[r.TestName] = r.Status
	}
	if statusMap["TestPass"] != "passed" {
		t.Errorf("TestPass status = %q, want passed", statusMap["TestPass"])
	}
	if statusMap["TestFail"] != "failed" {
		t.Errorf("TestFail status = %q, want failed", statusMap["TestFail"])
	}
	if statusMap["TestSkip"] != "skipped" {
		t.Errorf("TestSkip status = %q, want skipped", statusMap["TestSkip"])
	}
}

// [test-spec,id=TS-TR-003,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseJUnit_ErrorMessage verifies that failure messages are captured.
func TestParseJUnit_ErrorMessage(t *testing.T) {
	f := writeTempXML(t, sampleJUnit)
	results, _ := ParseJUnit(f, "proj", "linux")
	for _, r := range results {
		if r.TestName == "TestFail" {
			if r.Error == "" {
				t.Error("expected error message for TestFail")
			}
			return
		}
	}
	t.Error("TestFail not found in results")
}

// [test-spec,id=TS-TR-004,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseJUnit_Metadata verifies ID, package, project and platform fields.
func TestParseJUnit_Metadata(t *testing.T) {
	f := writeTempXML(t, sampleJUnit)
	results, _ := ParseJUnit(f, "myproject", "windows")
	for _, r := range results {
		if r.Project != "myproject" {
			t.Errorf("Project = %q, want myproject", r.Project)
		}
		if r.Platform != "windows" {
			t.Errorf("Platform = %q, want windows", r.Platform)
		}
		if r.Package != "com.example.MyTests" {
			t.Errorf("Package = %q, want com.example.MyTests", r.Package)
		}
	}
}

// [test-spec,id=TS-TR-005,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseJUnit_Timestamp verifies that timestamp is parsed from XML.
func TestParseJUnit_Timestamp(t *testing.T) {
	f := writeTempXML(t, sampleJUnit)
	results, _ := ParseJUnit(f, "proj", "linux")
	if len(results) == 0 {
		t.Fatal("no results")
	}
	if results[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

// [test-spec,id=TS-TR-006,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseJUnit_FileNotFound verifies error for missing file.
func TestParseJUnit_FileNotFound(t *testing.T) {
	_, err := ParseJUnit("/nonexistent/file.xml", "proj", "linux")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// [test-spec,id=TS-TR-007,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseJUnit_InvalidXML verifies error for malformed XML.
func TestParseJUnit_InvalidXML(t *testing.T) {
	f := writeTempXML(t, "<not valid xml")
	_, err := ParseJUnit(f, "proj", "linux")
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

// [test-spec,id=TS-TR-008,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseJUnit_ErrorElement verifies that <error> elements set status to failed.
func TestParseJUnit_ErrorElement(t *testing.T) {
	xml := `<testsuite name="T" tests="1" time="0">
  <testcase name="TestErr" classname="pkg" time="0">
    <error message="panic">stack trace here</error>
  </testcase>
</testsuite>`
	f := writeTempXML(t, xml)
	results, _ := ParseJUnit(f, "proj", "linux")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "failed" {
		t.Errorf("status = %q, want failed", results[0].Status)
	}
}

// [test-spec,id=TS-TR-009,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseJUnit_SystemOut verifies that stdout is captured.
func TestParseJUnit_SystemOut(t *testing.T) {
	f := writeTempXML(t, sampleJUnit)
	results, _ := ParseJUnit(f, "proj", "linux")
	for _, r := range results {
		if r.TestName == "TestPass" && r.Stdout == "" {
			t.Error("expected stdout for TestPass")
		}
	}
}

// [test-spec,id=TS-TR-024,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestParseJUnit_ShortID verifies that IDs use short classname (last dot-segment, no platform prefix).
func TestParseJUnit_ShortID(t *testing.T) {
	f := writeTempXML(t, sampleJUnit)
	results, _ := ParseJUnit(f, "proj", "linux")
	for _, r := range results {
		want := "MyTests::" + r.TestName
		if r.ID != want {
			t.Errorf("ID = %q, want %q", r.ID, want)
		}
	}
}

func writeTempXML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "results.xml")
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}
