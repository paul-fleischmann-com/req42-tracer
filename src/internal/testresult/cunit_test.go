package testresult

import (
	"path/filepath"
	"testing"
)

// [test-spec,id=TS-CUNIT-001,req="SWR-CUNIT-PARSER-001",dsn=comp.testresult.cunit,aspice=SWE.4-BP4]
// ParseCUnit correctly parses passed and failed test cases from CUnit XML
func TestParseCUnit(t *testing.T) {
	results, err := ParseCUnit(filepath.Join("testdata", "cunit_results.xml"), "test", "linux")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	byName := make(map[string]string) // name → status
	for _, r := range results {
		byName[r.TestName] = r.Status
	}

	if byName["test_gpio_init"] != "passed" {
		t.Errorf("test_gpio_init: expected passed, got %q", byName["test_gpio_init"])
	}
	if byName["test_gpio_read"] != "failed" {
		t.Errorf("test_gpio_read: expected failed, got %q", byName["test_gpio_read"])
	}
}

// [test-spec,id=TS-CUNIT-002,req="SWR-CUNIT-PARSER-001",aspice=SWE.4-BP4]
// ParseCUnit sets suite name as package and populates Error on failures
func TestParseCUnit_Metadata(t *testing.T) {
	results, err := ParseCUnit(filepath.Join("testdata", "cunit_results.xml"), "myproject", "linux")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.Package != "GPIO Tests" {
			t.Errorf("expected Package='GPIO Tests', got %q", r.Package)
		}
		if r.Project != "myproject" {
			t.Errorf("expected project=myproject, got %q", r.Project)
		}
		if r.Status == "failed" && r.Error == "" {
			t.Errorf("expected non-empty Error for failed test %q", r.TestName)
		}
	}
}
