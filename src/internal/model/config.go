package model

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// Config represents the .req42.yaml configuration file.
type Config struct {
	Projects       map[string]*ProjectConfig `yaml:"projects"`
	DefaultProject string                    `yaml:"default-project"` // optional; derived from first projects key if empty
	GoSrcDir       string                    `yaml:"go-src-dir"`      // optional; root dir scanned for *_test.go annotations
	CSources       []CSourceConfig           `yaml:"c-sources"`       // optional; C/H dirs scanned for @req/@arch annotations
	Bausteinsicht  struct {
		Model string `yaml:"model"`
	} `yaml:"bausteinsicht"`
	TestResults []TestResultSource `yaml:"test-results"`
	Plugins     PluginConfig       `yaml:"plugins"`
	Rules      map[string]string `yaml:"rules"`       // error, warning, off
	RuleParams map[string]int    `yaml:"rule-params"` // numeric thresholds per rule
	ASPICE struct {
		AutoDerive    bool     `yaml:"auto-derive"`
		Processes     []string `yaml:"processes"`
		ProcessRules  map[string]map[string]string `yaml:"process-rules"`
	} `yaml:"aspice"`
	Reports struct {
		HTML struct {
			Output          string `yaml:"output"`
			IncludeGraph    bool   `yaml:"include-graph"`
			IncludeMatrix   bool   `yaml:"include-matrix"`
			IncludeASPICE   bool   `yaml:"include-aspice"`
			Theme           string `yaml:"theme"`
			SourceBaseURL   string `yaml:"source-base-url"` // base URL for clickable impl= links (e.g. GitHub blob URL)
		} `yaml:"html"`
		CLI struct {
			Format string `yaml:"format"` // text, markdown, json
		} `yaml:"cli"`
	} `yaml:"reports"`
}

// CSourceConfig configures a set of C/H source directories to scan for traceability annotations.
type CSourceConfig struct {
	Paths []string `yaml:"paths"` // directories containing .c/.h files
}

// TestResultSource configures a single test result input.
type TestResultSource struct {
	Format     string `yaml:"format"`           // junit, go-test-json, plugin
	Path       string `yaml:"path"`
	PluginPath string `yaml:"plugin,omitempty"` // only for format: plugin
}

// PluginSource configures a single parser plugin with an input file.
type PluginSource struct {
	Type       string `yaml:"type"`
	Path       string `yaml:"path,omitempty"`
	PluginPath string `yaml:"plugin"`
}

// PluginConfig holds all plugin configurations grouped by category.
type PluginConfig struct {
	Parsers      []PluginSource `yaml:"parsers"`
	Preprocessors []PluginSource `yaml:"preprocessors"`
	Enrichers    []PluginSource `yaml:"enrichers"`
	Validators   []PluginSource `yaml:"validators"`
	Linkers      []PluginSource `yaml:"linkers"`
	Reporters    []PluginSource `yaml:"reporters"`
	CIAdapters   []PluginSource `yaml:"ci_adapters"`
	SyncAdapters []PluginSource `yaml:"sync_adapters"`
	Notifiers    []PluginSource `yaml:"notifiers"`
}

// ProjectConfig represents a single project in the configuration.
type ProjectConfig struct {
	Path string `yaml:"path"`
	Docs string `yaml:"docs"`
}

// LoadConfig loads the .req42.yaml configuration file.
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Set defaults
	if config.Projects == nil {
		config.Projects = make(map[string]*ProjectConfig)
	}
	if config.Rules == nil {
		config.Rules = make(map[string]string)
	}
	if config.ASPICE.ProcessRules == nil {
		config.ASPICE.ProcessRules = make(map[string]map[string]string)
	}

	// Default ASPICE processes if not specified
	if len(config.ASPICE.Processes) == 0 {
		config.ASPICE.Processes = []string{"SWE.1", "SWE.2", "SWE.3", "SWE.5"}
	}

	// Default report paths
	if config.Reports.HTML.Output == "" {
		config.Reports.HTML.Output = "reports/traceability-report.html"
	}
	if config.Reports.CLI.Format == "" {
		config.Reports.CLI.Format = "text"
	}

	return config, nil
}

// GetDefaultProject returns the configured default project name.
// Priority: explicit default-project field → alphabetically first key in projects map → "software".
func (c *Config) GetDefaultProject() string {
	if c.DefaultProject != "" {
		return c.DefaultProject
	}
	keys := make([]string, 0, len(c.Projects))
	for name := range c.Projects {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	if len(keys) > 0 {
		return keys[0]
	}
	return "software"
}

// SetDefault sets a default value for a rule if not already set.
func (c *Config) SetDefault(rule, value string) {
	if _, exists := c.Rules[rule]; !exists {
		c.Rules[rule] = value
	}
}

// GetRule returns the rule level for a given rule name.
func (c *Config) GetRule(ruleName string) string {
	if level, exists := c.Rules[ruleName]; exists {
		return level
	}
	return "warning" // Default to warning if not specified
}
