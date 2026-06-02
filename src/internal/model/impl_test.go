package model

import "testing"

// [test-spec,id=TS-IMPL-001,req="SWR-IMPL-REF-001",aspice=SWE.4-BP4]
// ParseImplRef — Go package path (no file extension → IsFile=false)
func TestParseImplRef_Package(t *testing.T) {
	r := ParseImplRef("src/internal/parser")
	if r.IsFile {
		t.Errorf("expected IsFile=false for package path")
	}
	if r.FilePath != "" {
		t.Errorf("expected empty FilePath for package path, got %q", r.FilePath)
	}
	if r.Raw != "src/internal/parser" {
		t.Errorf("unexpected Raw %q", r.Raw)
	}
}

// [test-spec,id=TS-IMPL-002,req="SWR-IMPL-REF-001",aspice=SWE.4-BP4]
// ParseImplRef — file-only format
func TestParseImplRef_FileOnly(t *testing.T) {
	r := ParseImplRef("src/hal/gpio.c")
	if !r.IsFile {
		t.Errorf("expected IsFile=true")
	}
	if r.FilePath != "src/hal/gpio.c" {
		t.Errorf("unexpected FilePath %q", r.FilePath)
	}
	if r.FuncName != "" || r.Line != 0 {
		t.Errorf("expected empty FuncName and Line=0")
	}
}

// [test-spec,id=TS-IMPL-003,req="SWR-IMPL-REF-001",aspice=SWE.4-BP4]
// ParseImplRef — file:func format
func TestParseImplRef_FileFunc(t *testing.T) {
	r := ParseImplRef("src/hal/gpio.c:gpio_init")
	if !r.IsFile || r.FilePath != "src/hal/gpio.c" || r.FuncName != "gpio_init" || r.Line != 0 {
		t.Errorf("unexpected %+v", r)
	}
}

// [test-spec,id=TS-IMPL-004,req="SWR-IMPL-REF-001",aspice=SWE.4-BP4]
// ParseImplRef — file:func:line format
func TestParseImplRef_FileFuncLine(t *testing.T) {
	r := ParseImplRef("src/hal/gpio.c:gpio_init:23")
	if !r.IsFile || r.FilePath != "src/hal/gpio.c" || r.FuncName != "gpio_init" || r.Line != 23 {
		t.Errorf("unexpected %+v", r)
	}
}

// [test-spec,id=TS-IMPL-005,req="SWR-IMPL-REF-001",aspice=SWE.4-BP4]
// ImplLink generates correct GitHub URL with line anchor
func TestImplRef_ImplLink(t *testing.T) {
	r := ParseImplRef("src/hal/gpio.c:gpio_init:23")
	got := r.ImplLink("https://github.com/org/repo/blob/master")
	want := "https://github.com/org/repo/blob/master/src/hal/gpio.c#L23"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// [test-spec,id=TS-IMPL-006,req="SWR-IMPL-REF-001",aspice=SWE.4-BP4]
// ImplLink returns empty string when no baseURL is set
func TestImplRef_ImplLink_NoBase(t *testing.T) {
	r := ParseImplRef("src/hal/gpio.c:gpio_init:23")
	if got := r.ImplLink(""); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// [test-spec,id=TS-IMPL-007,req="SWR-IMPL-REF-001",aspice=SWE.4-BP4]
// ImplRef.ShortLabel formats correctly for all variants
func TestImplRef_ShortLabel(t *testing.T) {
	cases := []struct{ in, want string }{
		{"src/internal/parser", "src/internal/parser"},
		{"src/hal/gpio.c", "src/hal/gpio.c"},
		{"src/hal/gpio.c:gpio_init", "src/hal/gpio.c:gpio_init"},
		{"src/hal/gpio.c:gpio_init:23", "src/hal/gpio.c:gpio_init:23"},
	}
	for _, c := range cases {
		if got := ParseImplRef(c.in).ShortLabel(); got != c.want {
			t.Errorf("ShortLabel(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
