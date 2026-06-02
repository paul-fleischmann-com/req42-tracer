package testresult

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"github.com/paulefl/req42-tracer/src/internal/model"
)

// CUnit XML structures — CUnit Automated test output (CUnitAutomated.xml).

type cunitReport struct {
	XMLName xml.Name      `xml:"CUNIT_TEST_RUN_REPORT"`
	Listing cunitListing  `xml:"CUNIT_RESULT_LISTING"`
}

type cunitListing struct {
	Suites []cunitSuite `xml:"CUNIT_RUN_SUITE"`
}

type cunitSuite struct {
	Name     string        `xml:"SUITE_NAME"`
	Successes []cunitSuccess `xml:"CUNIT_RUN_TEST_SUCCESS"`
	Failures  []cunitFailure `xml:"CUNIT_RUN_TEST_FAILURE"`
}

type cunitSuccess struct {
	Name string `xml:"TEST_NAME"`
}

type cunitFailure struct {
	Name      string `xml:"TEST_NAME"`
	Condition string `xml:"CONDITION"`
	File      string `xml:"FILE"`
	Line      int    `xml:"LINE"`
}

// ParseCUnit parses a CUnit Automated XML test result file.
func ParseCUnit(filePath, project, platform string) ([]*model.TestResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CUnit file %s: %w", filePath, err)
	}

	var report cunitReport
	if err := xml.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to parse CUnit XML %s: %w", filePath, err)
	}

	now := time.Now()
	var results []*model.TestResult
	for _, suite := range report.Listing.Suites {
		for _, s := range suite.Successes {
			results = append(results, &model.TestResult{
				ID:         shortPackage(suite.Name) + "::" + s.Name,
				Package:    suite.Name,
				TestName:   s.Name,
				FullName:   suite.Name + "::" + s.Name,
				Status:     "passed",
				Timestamp:  now,
				Project:    project,
				Platform:   platform,
				Attributes: make(map[string]string),
			})
		}
		for _, f := range suite.Failures {
			errMsg := f.Condition
			results = append(results, &model.TestResult{
				ID:         shortPackage(suite.Name) + "::" + f.Name,
				Package:    suite.Name,
				TestName:   f.Name,
				FullName:   suite.Name + "::" + f.Name,
				Status:     "failed",
				Error:      errMsg,
				Timestamp:  now,
				Project:    project,
				Platform:   platform,
				Attributes: make(map[string]string),
			})
		}
	}
	return results, nil
}
