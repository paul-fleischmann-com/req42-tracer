package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

// ADocParser parses AsciiDoc files and extracts REQ42, ARC42 blocks.
type ADocParser struct {
	filePath string
}

// NewADocParser creates a new parser for a single .adoc file.
func NewADocParser(filePath string) *ADocParser {
	return &ADocParser{filePath: filePath}
}

// ParseRequirements extracts all [req] blocks from the file.
func (p *ADocParser) ParseRequirements(project string) ([]*model.Requirement, error) {
	var requirements []*model.Requirement
	lineNum := 1

	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[req,") {
			// Handle multi-line blocks
			fullBlock := line
			for !strings.Contains(fullBlock, "]") && scanner.Scan() {
				fullBlock += " " + strings.TrimSpace(scanner.Text())
				lineNum++
			}
			req, err := p.parseReqBlock(fullBlock, lineNum, project, scanner)
			if err == nil && req != nil {
				requirements = append(requirements, req)
			}
		}
		lineNum++
	}

	return requirements, nil
}

// ParseArchElements extracts all [arch] blocks from the file.
func (p *ADocParser) ParseArchElements(project string) ([]*model.ArchElement, error) {
	var archElements []*model.ArchElement
	lineNum := 1

	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[arch,") {
			// Handle multi-line blocks
			fullBlock := line
			for !strings.Contains(fullBlock, "]") && scanner.Scan() {
				fullBlock += " " + strings.TrimSpace(scanner.Text())
				lineNum++
			}
			arch, err := p.parseArchBlock(fullBlock, lineNum, project, scanner)
			if err == nil && arch != nil {
				archElements = append(archElements, arch)
			}
		}
		lineNum++
	}

	return archElements, nil
}

// ParseTestSpecs extracts all [test-spec] blocks from the file.
func (p *ADocParser) ParseTestSpecs(project string) ([]*model.TestSpec, error) {
	var testSpecs []*model.TestSpec
	lineNum := 1

	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[test-spec,") {
			// Handle multi-line blocks
			fullBlock := line
			for !strings.Contains(fullBlock, "]") && scanner.Scan() {
				fullBlock += " " + strings.TrimSpace(scanner.Text())
				lineNum++
			}
			spec, err := p.parseTestSpecBlock(fullBlock, lineNum, project, scanner)
			if err == nil && spec != nil {
				testSpecs = append(testSpecs, spec)
			}
		}
		lineNum++
	}

	return testSpecs, nil
}

// parseReqBlock parses a single [req,...] block.
func (p *ADocParser) parseReqBlock(blockLine string, lineNum int, project string, scanner *bufio.Scanner) (*model.Requirement, error) {
	attrs := extractAttributes(blockLine)

	req := &model.Requirement{
		ID:         attrs["id"],
		Project:    project,
		FilePath:   p.filePath,
		LineNumber: lineNum,
		Attributes: attrs,
		Priority:   "medium",
		Status:     "draft",
	}

	if req.ID == "" {
		return nil, fmt.Errorf("requirement missing id at line %d", lineNum)
	}

	if v, ok := attrs["version"]; ok {
		if versionInt, err := strconv.Atoi(v); err == nil {
			req.Version = versionInt
		}
	}

	if priority, ok := attrs["priority"]; ok {
		req.Priority = priority
	}

	if aspice, ok := attrs["aspice"]; ok {
		req.ASPICE = aspice
	}

	if status, ok := attrs["status"]; ok {
		req.Status = status
	}

	if reviewed, ok := attrs["reviewed-by"]; ok {
		req.ReviewedBy = reviewed
	}

	if date, ok := attrs["reviewed-date"]; ok {
		if t, err := time.Parse("2006-01-02", date); err == nil {
			req.ReviewedDate = t
		}
	}

	if derives, ok := attrs["derives"]; ok {
		req.Derives = parseIDList(derives)
	}

	// Read next lines for title and text
	if scanner.Scan() {
		titleLine := scanner.Text()
		if strings.HasPrefix(titleLine, "==") {
			req.Title = strings.TrimSpace(strings.TrimPrefix(titleLine, "=="))
		}
	}

	return req, nil
}

// parseArchBlock parses a single [arch,...] block.
func (p *ADocParser) parseArchBlock(blockLine string, lineNum int, project string, scanner *bufio.Scanner) (*model.ArchElement, error) {
	attrs := extractAttributes(blockLine)

	arch := &model.ArchElement{
		ID:         attrs["id"],
		Project:    project,
		FilePath:   p.filePath,
		LineNumber: lineNum,
		Attributes: attrs,
	}

	if arch.ID == "" {
		return nil, fmt.Errorf("arch element missing id at line %d", lineNum)
	}

	if parent, ok := attrs["parent"]; ok {
		arch.Parent = parent
	}

	if aspice, ok := attrs["aspice"]; ok {
		arch.ASPICE = aspice
	}

	if req, ok := attrs["req"]; ok {
		arch.Req = parseIDList(req)
	}

	if impl, ok := attrs["impl"]; ok {
		arch.Impl = impl
	}

	if testSpec, ok := attrs["test-spec"]; ok {
		arch.TestSpec = testSpec
	}

	// Read next lines for title and text
	if scanner.Scan() {
		titleLine := scanner.Text()
		// Support both == (level 1) and === (level 2+)
		trimmed := strings.TrimSpace(titleLine)
		if strings.HasPrefix(trimmed, "===") {
			arch.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "==="))
		} else if strings.HasPrefix(trimmed, "==") {
			arch.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "=="))
		}
	}

	return arch, nil
}

// parseTestSpecBlock parses a single [test-spec,...] block.
func (p *ADocParser) parseTestSpecBlock(blockLine string, lineNum int, project string, scanner *bufio.Scanner) (*model.TestSpec, error) {
	attrs := extractAttributes(blockLine)

	spec := &model.TestSpec{
		ID:         attrs["id"],
		Project:    project,
		FilePath:   p.filePath,
		LineNumber: lineNum,
		Attributes: attrs,
	}

	if spec.ID == "" {
		return nil, fmt.Errorf("test-spec missing id at line %d", lineNum)
	}

	if req, ok := attrs["req"]; ok {
		spec.Req = parseIDList(req)
	}

	if arch, ok := attrs["arch"]; ok {
		spec.Arch = parseIDList(arch)
	}

	if dsn, ok := attrs["dsn"]; ok {
		spec.Dsn = parseIDList(dsn)
	}

	// Read next lines for title and text
	if scanner.Scan() {
		titleLine := scanner.Text()
		if strings.HasPrefix(titleLine, "==") {
			spec.Title = strings.TrimSpace(strings.TrimPrefix(titleLine, "=="))
		}
	}

	return spec, nil
}

// ParseDesignElements extracts all [dsn] blocks from the file.
func (p *ADocParser) ParseDesignElements(project string) ([]*model.DesignElement, error) {
	var elements []*model.DesignElement
	lineNum := 1

	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[dsn,") {
			fullBlock := line
			for !strings.Contains(fullBlock, "]") && scanner.Scan() {
				fullBlock += " " + strings.TrimSpace(scanner.Text())
				lineNum++
			}
			elem, err := p.parseDsnBlock(fullBlock, lineNum, project, scanner)
			if err == nil && elem != nil {
				elements = append(elements, elem)
			}
		}
		lineNum++
	}

	return elements, nil
}

// parseDsnBlock parses a single [dsn,...] block.
func (p *ADocParser) parseDsnBlock(blockLine string, lineNum int, project string, scanner *bufio.Scanner) (*model.DesignElement, error) {
	attrs := extractAttributes(blockLine)

	elem := &model.DesignElement{
		ID:         attrs["id"],
		Project:    project,
		FilePath:   p.filePath,
		LineNumber: lineNum,
		Attributes: attrs,
	}

	if elem.ID == "" {
		return nil, fmt.Errorf("dsn element missing id at line %d", lineNum)
	}

	if arch, ok := attrs["arch"]; ok {
		elem.Arch = arch
	}

	if aspice, ok := attrs["aspice"]; ok {
		elem.ASPICE = aspice
	}

	if impl, ok := attrs["impl"]; ok {
		elem.Impl = impl
	}

	if scanner.Scan() {
		titleLine := scanner.Text()
		trimmed := strings.TrimSpace(titleLine)
		if strings.HasPrefix(trimmed, "===") {
			elem.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "==="))
		} else if strings.HasPrefix(trimmed, "==") {
			elem.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "=="))
		}
	}

	return elem, nil
}

// extractAttributes parses block attributes from a line like [type,attr1=val1,attr2=val2]
// Multi-value attributes (req=A,B,C) are supported: parts without '=' are appended
// to the previous attribute's value as comma-separated continuations.
func extractAttributes(line string) map[string]string {
	attrs := make(map[string]string)

	// Find the block tag: [...]
	start := strings.Index(line, "[")
	end := strings.LastIndex(line, "]")
	if start < 0 || end < 0 || start >= end {
		return attrs
	}

	content := line[start+1 : end]
	parts := strings.Split(content, ",")

	var lastKey string
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if i == 0 {
			// First part is the block type (req, arch, test-spec)
			attrs["type"] = part
			lastKey = ""
			continue
		}

		if idx := strings.Index(part, "="); idx > 0 {
			key := strings.TrimSpace(part[:idx])
			value := strings.TrimSpace(part[idx+1:])
			// Remove quotes if present
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
			attrs[key] = value
			lastKey = key
		} else if lastKey != "" && part != "" {
			// No '=' — continuation of previous multi-value attribute (e.g. req=A,B,C)
			attrs[lastKey] = attrs[lastKey] + "," + part
		}
	}

	return attrs
}

// parseIDList parses a comma-separated list of IDs.
func parseIDList(s string) []string {
	var ids []string
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			ids = append(ids, part)
		}
	}
	return ids
}


// ParseAllFromDir parses all .adoc files in a directory.
func ParseAllFromDir(dirPath, project string) (*model.TraceabilityGraph, error) {
	graph := &model.TraceabilityGraph{
		Requirements:   make(map[string]*model.Requirement),
		ArchElements:   make(map[string]*model.ArchElement),
		DesignElements: make(map[string]*model.DesignElement),
		TestSpecs:      make(map[string]*model.TestSpec),
		TestCodes:      make(map[string]*model.TestCode),
		TestResults:    make(map[string]*model.TestResult),
		Links:          []*model.TraceLink{},
	}

	// Walk directory for .adoc files
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".adoc") {
			return nil
		}

		parser := NewADocParser(path)

		// Parse requirements
		reqs, err := parser.ParseRequirements(project)
		if err == nil {
			for _, req := range reqs {
				graph.Requirements[req.ID] = req
			}
		}

		// Parse architecture elements
		archs, err := parser.ParseArchElements(project)
		if err == nil {
			for _, arch := range archs {
				graph.ArchElements[arch.ID] = arch
			}
		}

		// Parse design elements (SWE.3)
		dsns, err := parser.ParseDesignElements(project)
		if err == nil {
			for _, dsn := range dsns {
				graph.DesignElements[dsn.ID] = dsn
			}
		}

		// Parse test specifications
		specs, err := parser.ParseTestSpecs(project)
		if err == nil {
			for _, spec := range specs {
				graph.TestSpecs[spec.ID] = spec
			}
		}

		return nil
	})

	return graph, err
}
