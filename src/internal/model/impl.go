package model

import (
	"strconv"
	"strings"
)

// ImplRef represents a parsed impl= attribute value.
// Supports four formats:
//
//	"src/internal/parser"          → Go package path
//	"src/hal/gpio.c"               → source file
//	"src/hal/gpio.c:gpio_init"     → file + function
//	"src/hal/gpio.c:gpio_init:23"  → file + function + line
type ImplRef struct {
	Raw      string // original unparsed value
	FilePath string // file path component (empty for pure Go package refs)
	FuncName string // function/symbol name (may be empty)
	Line     int    // line number (0 = not set)
	IsFile   bool   // true when FilePath is a concrete file (has extension)
}

// ParseImplRef parses an impl= value into its components.
func ParseImplRef(s string) ImplRef {
	ref := ImplRef{Raw: s}
	if s == "" {
		return ref
	}
	parts := strings.SplitN(s, ":", 3)

	// Detect whether the first component is a file (has an extension with a dot after the last slash).
	first := parts[0]
	base := first
	if idx := strings.LastIndex(first, "/"); idx >= 0 {
		base = first[idx+1:]
	}
	ref.IsFile = strings.Contains(base, ".")

	if !ref.IsFile {
		// Pure Go package path — no further parsing.
		return ref
	}

	ref.FilePath = first
	if len(parts) >= 2 {
		ref.FuncName = parts[1]
	}
	if len(parts) == 3 {
		if n, err := strconv.Atoi(parts[2]); err == nil {
			ref.Line = n
		}
	}
	return ref
}

// ImplLink returns a GitHub-style URL for this impl ref given a base URL.
// baseURL should be like "https://github.com/org/repo/blob/master".
// Returns "" when the ref is not a file ref or baseURL is empty.
func (r ImplRef) ImplLink(baseURL string) string {
	if baseURL == "" || !r.IsFile || r.FilePath == "" {
		return ""
	}
	url := strings.TrimRight(baseURL, "/") + "/" + r.FilePath
	if r.Line > 0 {
		url += "#L" + strconv.Itoa(r.Line)
	}
	return url
}

// ShortLabel returns a display string: "file:func:line", "file:func", or "file".
func (r ImplRef) ShortLabel() string {
	if !r.IsFile {
		return r.Raw
	}
	if r.FuncName == "" {
		return r.FilePath
	}
	if r.Line > 0 {
		return r.FilePath + ":" + r.FuncName + ":" + strconv.Itoa(r.Line)
	}
	return r.FilePath + ":" + r.FuncName
}
