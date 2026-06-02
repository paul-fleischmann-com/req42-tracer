package parser

// [test-spec,id=TS-GOCODE-001,req=REQ-PARSE-002,dsn=comp.parser.gocode,aspice=SWE.4-BP3]
// TestParseGoTestFile_ExtractsTestCode verifies that [test-spec] comments above
// func Test* are parsed into TestCode entries with correct spec linkage.
// [end]

// [test-spec,id=TS-GOCODE-002,req=REQ-PARSE-002,aspice=SWE.4-BP3]
// TestParseGoTestFiles_Dir verifies that ParseGoTestFiles walks a directory and
// returns TestCode entries from all *_test.go files found.
// [end]

// [test-spec,id=TS-GOCODE-003,req=REQ-PARSE-002,aspice=SWE.4-BP3]
// TestParseGoTestFile_SkipsUnannotated verifies that test functions without a
// [test-spec] comment are not included in the output.
// [end]

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleGoTestFile = `package mypkg

// [test-spec,id=TS-FOO-001,req=REQ-001,aspice=SWE.4]
// TestFoo verifies foo behaviour.
func TestFoo(t *testing.T) {
	_ = t
}

func TestNoAnnotation(t *testing.T) {
	_ = t
}

// [test-spec,id=TS-BAR-001,req=REQ-002,aspice=SWE.4]
func TestBar(t *testing.T) {
	_ = t
}
`

// [test-spec,id=TS-GOCODE-010,req=REQ-PARSE-002,dsn=comp.parser.gocode,aspice=SWE.4-BP3]
// TestParseGoTestFile_ExtractsTestCode verifies [test-spec] comments above Test* funcs are parsed.
func TestParseGoTestFile_ExtractsTestCode(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "foo_test.go")
	if err := os.WriteFile(f, []byte(sampleGoTestFile), 0644); err != nil {
		t.Fatal(err)
	}

	_, codes, err := parseGoTestFile(f, "proj")
	if err != nil {
		t.Fatalf("parseGoTestFile: %v", err)
	}

	if len(codes) != 2 {
		t.Fatalf("expected 2 TestCode entries, got %d", len(codes))
	}

	byFunc := make(map[string]string) // function → specID
	for _, c := range codes {
		byFunc[c.Function] = c.TestSpec
	}

	if byFunc["TestFoo"] != "TS-FOO-001" {
		t.Errorf("TestFoo spec = %q, want TS-FOO-001", byFunc["TestFoo"])
	}
	if byFunc["TestBar"] != "TS-BAR-001" {
		t.Errorf("TestBar spec = %q, want TS-BAR-001", byFunc["TestBar"])
	}
	if _, ok := byFunc["TestNoAnnotation"]; ok {
		t.Error("TestNoAnnotation should not be included (no [test-spec] annotation)")
	}
}

// [test-spec,id=TS-GOCODE-011,req=REQ-PARSE-002,aspice=SWE.4-BP3]
// TestParseGoTestFiles_Dir verifies ParseGoTestFiles walks directory and returns TestCode entries.
func TestParseGoTestFiles_Dir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "pkg")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test file in subdir
	f := filepath.Join(subdir, "bar_test.go")
	content := "package pkg\n\n// [test-spec,id=TS-DIR-001,req=REQ-001]\nfunc TestDir(t *testing.T) {}\n"
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	// Non-test file — must be ignored
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	g, err := ParseGoTestFiles(dir, "proj")
	if err != nil {
		t.Fatalf("ParseGoTestFiles: %v", err)
	}
	if len(g.TestCodes) != 1 {
		t.Fatalf("expected 1 TestCode, got %d", len(g.TestCodes))
	}
	for _, code := range g.TestCodes {
		if code.TestSpec != "TS-DIR-001" {
			t.Errorf("TestSpec = %q, want TS-DIR-001", code.TestSpec)
		}
		if code.Function != "TestDir" {
			t.Errorf("Function = %q, want TestDir", code.Function)
		}
		if code.Language != "go" {
			t.Errorf("Language = %q, want go", code.Language)
		}
	}
}

// [test-spec,id=TS-GOCODE-012,req=REQ-PARSE-002,aspice=SWE.4-BP3]
// TestParseGoTestFile_SkipsUnannotated verifies unannotated test functions are excluded.
func TestParseGoTestFile_SkipsUnannotated(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "plain_test.go")
	content := "package pkg\n\nfunc TestPlain(t *testing.T) {}\nfunc TestAlsoPlain(t *testing.T) {}\n"
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, codes, err := parseGoTestFile(f, "proj")
	if err != nil {
		t.Fatalf("parseGoTestFile: %v", err)
	}
	if len(codes) != 0 {
		t.Errorf("expected 0 TestCode entries for unannotated file, got %d", len(codes))
	}
}
