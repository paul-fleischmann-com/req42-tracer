package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

var (
	reAtReq    = regexp.MustCompile(`@req\s+(\S+)`)
	reAtArch   = regexp.MustCompile(`@arch\s+(\S+)`)
	reAtAspice = regexp.MustCompile(`@aspice\s+(\S+)`)

	// inline style: "/* req: SWR-001, arch: comp.hal.gpio */"
	reInlineReq  = regexp.MustCompile(`\breq:\s*([\w.\-]+(?:\s*,\s*[\w.\-]+)*)`)
	reInlineArch = regexp.MustCompile(`\barch:\s*([\w.\-]+)`)

	// C function declaration: captures the function name before the first '('.
	// Matches common return-type prefixes and plain identifiers.
	reCFunc = regexp.MustCompile(`^(?:(?:static|inline|extern|const|unsigned|signed|void|int|char|bool|float|double|size_t|uint\w*|int\w*|\w+_t)\s+)*(\w+)\s*\(`)
)

// ParseCSourceDirs walks all directories in each CSourceConfig, parses .c/.h files
// for @req / @arch / @aspice comment annotations, and returns a map of
//
//	archElementID → "filepath:funcName:lineNumber"
//
// impl-ref patches ready to be applied to the traceability graph.
func ParseCSourceDirs(sources []model.CSourceConfig, project string) (map[string]string, error) {
	implRefs := make(map[string]string)
	for _, src := range sources {
		for _, dir := range src.Paths {
			anns, err := walkCDir(dir)
			if err != nil {
				// non-fatal: keep going for other dirs
				fmt.Fprintf(os.Stderr, "Warning: C parser: %s: %v\n", dir, err)
				continue
			}
			for _, ann := range anns {
				if ann.FuncName == "" {
					continue
				}
				implRef := fmt.Sprintf("%s:%s:%d", ann.FilePath, ann.FuncName, ann.Line)
				for _, archID := range ann.Archs {
					if _, exists := implRefs[archID]; !exists {
						implRefs[archID] = implRef
					}
				}
			}
		}
	}
	return implRefs, nil
}

// cAnnotation holds the traceability markers extracted from a single C comment block.
type cAnnotation struct {
	FilePath string
	FuncName string // function that immediately follows the comment block
	Line     int    // line number of the function declaration
	Reqs     []string
	Archs    []string
	ASPICE   string
}

// walkCDir recursively walks dir and returns annotations from all .c/.h files.
func walkCDir(dir string) ([]cAnnotation, error) {
	var all []cAnnotation
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".c" && ext != ".h" {
			return nil
		}
		anns, parseErr := parseCFile(path)
		if parseErr == nil {
			all = append(all, anns...)
		}
		return nil
	})
	return all, err
}

// parseCFile extracts cAnnotation entries from a single C/H file.
// A pending annotation is committed when a function declaration is found on the
// next non-blank, non-comment line.
func parseCFile(filePath string) ([]cAnnotation, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var results []cAnnotation
	var pending *cAnnotation
	inBlock := false // inside a /* ... */ comment

	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)

		// Track block comment state
		if !inBlock && strings.Contains(raw, "/*") {
			inBlock = true
		}

		if inBlock {
			if pending == nil {
				pending = &cAnnotation{FilePath: filePath}
			}
			extractCTags(trimmed, pending)
			if strings.Contains(raw, "*/") {
				inBlock = false
			}
			continue
		}

		// Single-line // comments
		if strings.HasPrefix(trimmed, "//") {
			if pending == nil {
				pending = &cAnnotation{FilePath: filePath}
			}
			extractCTags(trimmed[2:], pending)
			continue
		}

		// Empty line — reset pending if it has no tags yet
		if trimmed == "" {
			if pending != nil && len(pending.Reqs) == 0 && len(pending.Archs) == 0 {
				pending = nil
			}
			continue
		}

		// Non-comment, non-empty: look for a function declaration
		if pending != nil {
			if m := reCFunc.FindStringSubmatch(trimmed); m != nil {
				if len(pending.Reqs) > 0 || len(pending.Archs) > 0 {
					pending.FuncName = m[1]
					pending.Line = lineNum
					results = append(results, *pending)
				}
			}
			pending = nil
			continue
		}
	}

	return results, scanner.Err()
}

// extractCTags parses @req / @arch / @aspice and inline "req: / arch:" markers
// from a single comment line and appends them to the annotation.
func extractCTags(line string, ann *cAnnotation) {
	if m := reAtReq.FindStringSubmatch(line); m != nil {
		ann.Reqs = appendUnique(ann.Reqs, m[1])
	}
	if m := reAtArch.FindStringSubmatch(line); m != nil {
		ann.Archs = appendUnique(ann.Archs, m[1])
	}
	if m := reAtAspice.FindStringSubmatch(line); m != nil && ann.ASPICE == "" {
		ann.ASPICE = m[1]
	}
	// inline style
	if m := reInlineReq.FindStringSubmatch(line); m != nil {
		for _, r := range strings.Split(m[1], ",") {
			if v := strings.TrimSpace(r); v != "" {
				ann.Reqs = appendUnique(ann.Reqs, v)
			}
		}
	}
	if m := reInlineArch.FindStringSubmatch(line); m != nil {
		ann.Archs = appendUnique(ann.Archs, m[1])
	}
}

func appendUnique(slice []string, v string) []string {
	for _, s := range slice {
		if s == v {
			return slice
		}
	}
	return append(slice, v)
}
