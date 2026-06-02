package testresult

import (
	"path/filepath"
	"testing"
)

// [test-spec,id=TS-UNITY-001,req="SWR-UNITY-PARSER-001",dsn=comp.testresult.unity,aspice=SWE.4-BP4]
// ParseUnity correctly parses passed and failed test cases from Unity XML
func TestParseUnity(t *testing.T) {
	results, err := ParseUnity(filepath.Join("testdata", "unity_results.xml"), "test", "linux")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	passed := results[0]
	if passed.TestName != "test_gpio_init" {
		t.Errorf("unexpected TestName %q", passed.TestName)
	}
	if passed.Status != "passed" {
		t.Errorf("expected passed, got %q", passed.Status)
	}

	failed := results[1]
	if failed.TestName != "test_gpio_read_fails" {
		t.Errorf("unexpected TestName %q", failed.TestName)
	}
	if failed.Status != "failed" {
		t.Errorf("expected failed, got %q", failed.Status)
	}
	if failed.Error == "" {
		t.Error("expected non-empty Error for failed test")
	}
}

// [test-spec,id=TS-UNITY-002,req="SWR-UNITY-PARSER-001",aspice=SWE.4-BP4]
// ParseUnity sets project and platform on all results
func TestParseUnity_Metadata(t *testing.T) {
	results, err := ParseUnity(filepath.Join("testdata", "unity_results.xml"), "myproject", "embedded")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.Project != "myproject" {
			t.Errorf("expected project=myproject, got %q", r.Project)
		}
		if r.Platform != "embedded" {
			t.Errorf("expected platform=embedded, got %q", r.Platform)
		}
	}
}
