package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

// ParseGoTestFiles walks dirPath recursively, reads all *_test.go files and
// extracts [test-spec] comment annotations placed directly before func Test*
// declarations. Each annotation produces:
//   - a TestSpec entry (inline declaration — the annotation IS the spec)
//   - a TestCode entry linking the test function to the spec ID
func ParseGoTestFiles(dirPath, project string) (*model.TraceabilityGraph, error) {
	graph := &model.TraceabilityGraph{
		Requirements:   make(map[string]*model.Requirement),
		ArchElements:   make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:      make(map[string]*model.TestSpec),
		TestCodes:      make(map[string]*model.TestCode),
		TestResults:    make(map[string]*model.TestResult),
		Links:          []*model.TraceLink{},
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, "_test.go") {
			return err
		}
		specs, codes, parseErr := parseGoTestFile(path, project)
		if parseErr == nil {
			for _, spec := range specs {
				if _, exists := graph.TestSpecs[spec.ID]; !exists {
					graph.TestSpecs[spec.ID] = spec
				}
			}
			for _, code := range codes {
				graph.TestCodes[code.ID] = code
			}
		}
		return nil
	})

	return graph, err
}

// parseGoTestFile extracts TestSpec and TestCode entries from a single Go test file.
// A [test-spec,id=...,req=...,aspice=...] annotation above a func Test* both
// declares the TestSpec (inline) and links it to the implementing test function.
func parseGoTestFile(filePath, project string) ([]*model.TestSpec, []*model.TestCode, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	var specs []*model.TestSpec
	var codes []*model.TestCode
	var pendingSpec *model.TestSpec // spec built from last [test-spec] comment

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Detect [test-spec] comment: // [test-spec,id=...,req=...,aspice=...]
		if strings.HasPrefix(line, "// [test-spec,") {
			inner := line[3:] // strip "// "
			attrs := extractAttributes(inner)
			id, ok := attrs["id"]
			if !ok || id == "" {
				continue
			}
			// Parse req= (comma-separated, may be quoted)
			var reqs []string
			if r := strings.Trim(attrs["req"], `"`); r != "" {
				for _, v := range strings.Split(r, ",") {
					if v = strings.TrimSpace(v); v != "" {
						reqs = append(reqs, v)
					}
				}
			}
			// Parse dsn= (comma-separated)
			var dsns []string
			if d := strings.Trim(attrs["dsn"], `"`); d != "" {
				for _, v := range strings.Split(d, ",") {
					if v = strings.TrimSpace(v); v != "" {
						dsns = append(dsns, v)
					}
				}
			}
			// Parse arch= (comma-separated)
			var archs []string
			if a := strings.Trim(attrs["arch"], `"`); a != "" {
				for _, v := range strings.Split(a, ",") {
					if v = strings.TrimSpace(v); v != "" {
						archs = append(archs, v)
					}
				}
			}
			pendingSpec = &model.TestSpec{
				ID:         id,
				Req:        reqs,
				Dsn:        dsns,
				Arch:       archs,
				FilePath:   filePath,
				LineNumber: lineNum,
				Project:    project,
				Attributes: map[string]string{"aspice": strings.Trim(attrs["aspice"], `"`)},
			}
			continue
		}

		// Next comment line may be the title — attach to pending spec
		if strings.HasPrefix(line, "//") && pendingSpec != nil && pendingSpec.Title == "" {
			title := strings.TrimSpace(strings.TrimPrefix(line, "//"))
			if title != "" && !strings.HasPrefix(title, "[") {
				pendingSpec.Title = title
			}
			continue
		}

		// Detect test function declaration
		if strings.HasPrefix(line, "func Test") && pendingSpec != nil {
			funcName := extractFuncName(line)
			if funcName != "" {
				if pendingSpec.Title == "" {
					pendingSpec.Title = funcName
				}
				specs = append(specs, pendingSpec)
				code := &model.TestCode{
					ID:       fmt.Sprintf("%s::%s", filePath, funcName),
					TestSpec: pendingSpec.ID,
					File:     filePath,
					Function: funcName,
					Language: "go",
					Project:  project,
				}
				codes = append(codes, code)
			}
			pendingSpec = nil
			continue
		}

		// Reset pending spec if a non-comment, non-empty line appears
		if line != "" && !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "func Test") {
			pendingSpec = nil
		}
	}

	return specs, codes, scanner.Err()
}

// extractFuncName extracts the function name from a "func TestXxx(...)" line.
func extractFuncName(line string) string {
	// line looks like: func TestFoo(t *testing.T) {
	after := strings.TrimPrefix(line, "func ")
	paren := strings.Index(after, "(")
	if paren < 0 {
		return ""
	}
	return strings.TrimSpace(after[:paren])
}
