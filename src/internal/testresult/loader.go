package testresult

import (
	"fmt"
	"strings"

	"github.com/paulefl/req42-tracer/src/internal/model"
	"github.com/paulefl/req42-tracer/src/internal/plugin"
)

// Loader loads test results from various formats.
type Loader struct {
	project  string
	platform string
}

// NewLoader creates a new test result loader.
func NewLoader(project, platform string) *Loader {
	return &Loader{
		project:  project,
		platform: platform,
	}
}

// Load loads test results from a file based on format.
func (l *Loader) Load(filePath, format string) ([]*model.TestResult, error) {
	switch strings.ToLower(format) {
	case "junit":
		return ParseJUnit(filePath, l.project, l.platform)
	case "go-test", "go-test-json":
		return ParseGoTest(filePath, l.project, l.platform)
	default:
		return nil, fmt.Errorf("unsupported test result format: %s", format)
	}
}

// LoadPlugin loads test results via an external plugin binary.
func (l *Loader) LoadPlugin(filePath, pluginPath string) ([]*model.TestResult, error) {
	results, err := plugin.RunTestResultParser(pluginPath, filePath, l.project)
	if err != nil {
		return nil, err
	}
	for _, r := range results {
		r.Project = l.project
		r.Platform = l.platform
	}
	return results, nil
}

// LoadAll loads test results from multiple files based on configuration.
func LoadAll(graph *model.TraceabilityGraph, config *model.Config) error {
	for _, testSource := range config.TestResults {
		loader := NewLoader("software", "linux") // Default to linux
		var results []*model.TestResult
		var err error
		if strings.ToLower(testSource.Format) == "plugin" {
			results, err = loader.LoadPlugin(testSource.Path, testSource.PluginPath)
		} else {
			results, err = loader.Load(testSource.Path, testSource.Format)
		}
		if err != nil {
			// Log warning but continue
			continue
		}

		for _, result := range results {
			result.Project = "software"
			// Map to test platform if available
			if strings.Contains(result.Platform, "window") {
				result.Platform = "windows"
			} else if strings.Contains(result.Platform, "darwin") || strings.Contains(result.Platform, "macos") {
				result.Platform = "macos"
			} else {
				result.Platform = "linux"
			}

			result.ID = shortPackage(result.Package) + "::" + result.TestName
			graph.TestResults[result.ID] = result
		}
	}

	return nil
}
