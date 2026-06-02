package testresult

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

// Unity XML structures — Unity's own test-runner output format.

type unityTestRun struct {
	XMLName xml.Name          `xml:"TestRun"`
	Results []unityTestResult `xml:"TestResult>TestCase"`
	Summary unitySummary      `xml:"Summary"`
}

type unityTestResult struct {
	TestName   string `xml:"TestName"`
	SourceFile string `xml:"SourceFile"`
	SourceLine int    `xml:"SourceLine"`
	Message    string `xml:"Message"`
}

type unitySummary struct {
	Tests    int `xml:"Tests"`
	Failures int `xml:"Failures"`
	Ignored  int `xml:"Ignored"`
}

// ParseUnity parses a Unity XML test result file.
func ParseUnity(filePath, project, platform string) ([]*model.TestResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Unity file %s: %w", filePath, err)
	}

	var run unityTestRun
	if err := xml.Unmarshal(data, &run); err != nil {
		return nil, fmt.Errorf("failed to parse Unity XML %s: %w", filePath, err)
	}

	now := time.Now()
	var results []*model.TestResult
	for _, tc := range run.Results {
		status := "passed"
		errMsg := ""
		msg := tc.Message
		if msg != "OK" && msg != "" {
			status = "failed"
			errMsg = msg
		}

		pkg := tc.SourceFile
		result := &model.TestResult{
			ID:         shortPackage(pkg) + "::" + tc.TestName,
			Package:    pkg,
			TestName:   tc.TestName,
			FullName:   pkg + "::" + tc.TestName,
			Status:     status,
			Error:      errMsg,
			Timestamp:  now,
			Project:    project,
			Platform:   platform,
			Attributes: make(map[string]string),
		}
		results = append(results, result)
	}
	return results, nil
}
