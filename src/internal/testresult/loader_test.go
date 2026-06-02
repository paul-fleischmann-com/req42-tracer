package testresult

import (
	"testing"
)

// [test-spec,id=TS-TR-018,req="REQ-TESTING-001",dsn=comp.testresult.loader,aspice=SWE.4-BP4]
// TestLoader_Load_JUnit verifies that Loader dispatches to ParseJUnit for junit format.
func TestLoader_Load_JUnit(t *testing.T) {
	f := writeTempXML(t, sampleJUnit)
	loader := NewLoader("proj", "linux")
	results, err := loader.Load(f, "junit")
	if err != nil {
		t.Fatalf("Load junit error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

// [test-spec,id=TS-TR-019,req="REQ-TESTING-001",dsn=comp.testresult.loader,aspice="SWE.5.BP3"]
// TestLoader_Load_GoTest verifies that Loader dispatches to ParseGoTest for go-test format.
func TestLoader_Load_GoTest(t *testing.T) {
	f := writeTempJSON(t, sampleGoTest)
	loader := NewLoader("proj", "linux")
	results, err := loader.Load(f, "go-test")
	if err != nil {
		t.Fatalf("Load go-test error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

// [test-spec,id=TS-TR-020,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestLoader_Load_GoTestJSON verifies that go-test-json format alias works.
func TestLoader_Load_GoTestJSON(t *testing.T) {
	f := writeTempJSON(t, sampleGoTest)
	loader := NewLoader("proj", "linux")
	results, err := loader.Load(f, "go-test-json")
	if err != nil {
		t.Fatalf("Load go-test-json error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

// [test-spec,id=TS-TR-021,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestLoader_Load_UnknownFormat verifies error for unsupported format.
func TestLoader_Load_UnknownFormat(t *testing.T) {
	loader := NewLoader("proj", "linux")
	_, err := loader.Load("any.file", "unknown-format")
	if err == nil {
		t.Error("expected error for unknown format")
	}
}

// [test-spec,id=TS-TR-022,req="REQ-TESTING-001",aspice="SWE.5.BP3"]
// TestLoader_Load_CaseInsensitive verifies format matching is case-insensitive.
func TestLoader_Load_CaseInsensitive(t *testing.T) {
	f := writeTempXML(t, sampleJUnit)
	loader := NewLoader("proj", "linux")
	results, err := loader.Load(f, "JUNIT")
	if err != nil {
		t.Fatalf("Load JUNIT (uppercase) error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for uppercase format")
	}
}
