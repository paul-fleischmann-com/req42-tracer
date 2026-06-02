package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

// writeTempC creates a temp .c file with the given content and returns its path.
func writeTempC(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.c")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// [test-spec,id=TS-CPARSER-001,req="SWR-C-PARSER-001",aspice=SWE.4-BP4]
// C parser extracts @req and @arch from JavaDoc-style block comments
func TestParseCFile_JavaDocStyle(t *testing.T) {
	src := `/**
 * @req SWR-GPIO-001
 * @arch comp.hal.gpio
 * @aspice SWE.3
 */
void gpio_init(uint8_t pin) {
}
`
	anns, err := parseCFile(writeTempC(t, src))
	if err != nil {
		t.Fatal(err)
	}
	if len(anns) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(anns))
	}
	a := anns[0]
	if len(a.Reqs) != 1 || a.Reqs[0] != "SWR-GPIO-001" {
		t.Errorf("unexpected Reqs: %v", a.Reqs)
	}
	if len(a.Archs) != 1 || a.Archs[0] != "comp.hal.gpio" {
		t.Errorf("unexpected Archs: %v", a.Archs)
	}
	if a.ASPICE != "SWE.3" {
		t.Errorf("unexpected ASPICE: %q", a.ASPICE)
	}
	if a.FuncName != "gpio_init" {
		t.Errorf("unexpected FuncName: %q", a.FuncName)
	}
}

// [test-spec,id=TS-CPARSER-002,req="SWR-C-PARSER-001",aspice=SWE.4-BP4]
// C parser extracts inline-style "req: / arch:" markers
func TestParseCFile_InlineStyle(t *testing.T) {
	src := `/* req: SWR-GPIO-001, arch: comp.hal.gpio */
int gpio_read(uint8_t pin) { return 0; }
`
	anns, err := parseCFile(writeTempC(t, src))
	if err != nil {
		t.Fatal(err)
	}
	if len(anns) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(anns))
	}
	a := anns[0]
	if len(a.Reqs) == 0 || a.Reqs[0] != "SWR-GPIO-001" {
		t.Errorf("unexpected Reqs: %v", a.Reqs)
	}
	if len(a.Archs) == 0 || a.Archs[0] != "comp.hal.gpio" {
		t.Errorf("unexpected Archs: %v", a.Archs)
	}
	if a.FuncName != "gpio_read" {
		t.Errorf("unexpected FuncName: %q", a.FuncName)
	}
}

// [test-spec,id=TS-CPARSER-003,req="SWR-C-PARSER-001",aspice=SWE.4-BP4]
// C parser extracts annotations from single-line // comments
func TestParseCFile_LineComments(t *testing.T) {
	src := `// @req SWR-GPIO-002
// @arch comp.hal.gpio
void gpio_deinit(uint8_t pin) {}
`
	anns, err := parseCFile(writeTempC(t, src))
	if err != nil {
		t.Fatal(err)
	}
	if len(anns) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(anns))
	}
	if anns[0].Reqs[0] != "SWR-GPIO-002" || anns[0].FuncName != "gpio_deinit" {
		t.Errorf("unexpected annotation: %+v", anns[0])
	}
}

// [test-spec,id=TS-CPARSER-004,req="SWR-C-PARSER-001",aspice=SWE.4-BP4]
// C parser returns no annotations when no traceability markers are present
func TestParseCFile_NoAnnotations(t *testing.T) {
	src := `/* This is a plain comment without traceability. */
void helper(void) {}
`
	anns, err := parseCFile(writeTempC(t, src))
	if err != nil {
		t.Fatal(err)
	}
	if len(anns) != 0 {
		t.Errorf("expected 0 annotations, got %d: %+v", len(anns), anns)
	}
}

// [test-spec,id=TS-CPARSER-005,req="SWR-C-PARSER-001",aspice=SWE.4-BP4]
// ParseCSourceDirs returns impl-ref map from a directory of .c files
func TestParseCSourceDirs(t *testing.T) {
	dir := t.TempDir()
	content := `/**
 * @req SWR-GPIO-001
 * @arch comp.hal.gpio
 */
void gpio_init(uint8_t pin) {}
`
	if err := os.WriteFile(filepath.Join(dir, "gpio.c"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	sources := []model.CSourceConfig{{Paths: []string{dir}}}
	refs, err := ParseCSourceDirs(sources, "test")
	if err != nil {
		t.Fatal(err)
	}
	ref, ok := refs["comp.hal.gpio"]
	if !ok {
		t.Fatalf("expected impl ref for comp.hal.gpio, got %v", refs)
	}
	if ref == "" {
		t.Error("impl ref is empty")
	}
}
