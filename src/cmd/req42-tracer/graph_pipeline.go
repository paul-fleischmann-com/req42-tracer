package main

import (
	"fmt"
	"os"

	"github.com/paulefl/req42-tracer/src/internal/graph"
	"github.com/paulefl/req42-tracer/src/internal/model"
	"github.com/paulefl/req42-tracer/src/internal/parser"
	"github.com/paulefl/req42-tracer/src/internal/plugin"
	"github.com/paulefl/req42-tracer/src/internal/testresult"
)

// buildGraph parses requirements, architecture, Bausteinsicht model, and Go test
// annotations, then derives ASPICE levels and builds trace links.
// reqDir and arcDir are the directories containing AsciiDoc sources.
// goSrcDir is the root directory scanned for *_test.go files (empty = skip).
// Returns the built traceability graph ready for analysis or reporting.
func buildGraph(config *model.Config, reqDir, arcDir, project string, verbose bool) (*model.TraceabilityGraph, error) {
	builder := graph.NewBuilder()

	if req, err := parser.ParseAllFromDir(reqDir, project); err == nil {
		if err := builder.MergeGraph(req); err != nil {
			return nil, fmt.Errorf("requirements merge from %s: %w", reqDir, err)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "Parsed requirements from %s\n", reqDir)
		}
	} else if verbose {
		fmt.Fprintf(os.Stderr, "Warning: no requirements found in %s: %v\n", reqDir, err)
	}

	if arch, err := parser.ParseAllFromDir(arcDir, project); err == nil {
		if err := builder.MergeGraph(arch); err != nil {
			return nil, fmt.Errorf("architecture merge from %s: %w", arcDir, err)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "Parsed architecture from %s\n", arcDir)
		}
	} else if verbose {
		fmt.Fprintf(os.Stderr, "Warning: no architecture found in %s: %v\n", arcDir, err)
	}

	if bPath := config.Bausteinsicht.Model; bPath != "" {
		loadBausteinsicht(builder, bPath, project, verbose)
	}

	// C source parsing: apply @arch → impl= patches onto arch elements.
	if len(config.CSources) > 0 {
		if implRefs, err := parser.ParseCSourceDirs(config.CSources, project); err == nil {
			builder.ApplyImplRefs(implRefs)
			if verbose {
				fmt.Fprintf(os.Stderr, "Applied %d C impl refs from c-sources\n", len(implRefs))
			}
		} else if verbose {
			fmt.Fprintf(os.Stderr, "Warning: C source parsing: %v\n", err)
		}
	}

	// Parse Go test files for [test-spec] annotations → explicit TestCode entries
	if goSrc := config.GoSrcDir; goSrc != "" {
		if goGraph, err := parser.ParseGoTestFiles(goSrc, project); err == nil {
			if mergeErr := builder.MergeGraph(goGraph); mergeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Go test code merge: %v\n", mergeErr)
			} else if verbose {
				fmt.Fprintf(os.Stderr, "Parsed Go test annotations from %s\n", goSrc)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Warning: could not parse Go test files: %v\n", err)
		}
	}

	// Plugin parsers: requirement, arch, test-spec, design
	runParserPlugins(builder, config, project, verbose)

	// Load test results before BuildLinks so name-based linking can match them.
	if err := testresult.LoadAll(builder.GetGraph(), config); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: could not load test results: %v\n", err)
	}

	// Enricher plugins: annotate elements with external metadata
	runEnricherPlugins(builder.GetGraph(), config, verbose)

	builder.DeriveASPICELevels()

	// Linker plugins: add custom trace links before BuildLinks
	runLinkerPlugins(builder, config, verbose)

	if err := builder.BuildLinks(); err != nil {
		return nil, err
	}

	return builder.GetGraph(), nil
}

func runParserPlugins(builder *graph.Builder, config *model.Config, project string, verbose bool) {
	for _, src := range config.Plugins.Parsers {
		var mergeErr error
		switch src.Type {
		case "requirement-parser":
			reqs, err := plugin.RunRequirementParser(src.PluginPath, src.Path, project)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: plugin %s: %v\n", src.PluginPath, err)
				continue
			}
			g := model.EmptyGraph()
			for _, r := range reqs {
				g.Requirements[r.ID] = r
			}
			mergeErr = builder.MergeGraph(g)
		case "arch-parser":
			elems, err := plugin.RunArchParser(src.PluginPath, src.Path, project)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: plugin %s: %v\n", src.PluginPath, err)
				continue
			}
			g := model.EmptyGraph()
			for _, e := range elems {
				g.ArchElements[e.ID] = e
			}
			mergeErr = builder.MergeGraph(g)
		case "test-spec-parser":
			specs, err := plugin.RunTestSpecParser(src.PluginPath, src.Path, project)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: plugin %s: %v\n", src.PluginPath, err)
				continue
			}
			g := model.EmptyGraph()
			for _, s := range specs {
				g.TestSpecs[s.ID] = s
			}
			mergeErr = builder.MergeGraph(g)
		case "design-parser":
			elems, err := plugin.RunDesignParser(src.PluginPath, src.Path, project)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: plugin %s: %v\n", src.PluginPath, err)
				continue
			}
			g := model.EmptyGraph()
			for _, e := range elems {
				g.DesignElements[e.ID] = e
			}
			mergeErr = builder.MergeGraph(g)
		default:
			fmt.Fprintf(os.Stderr, "Warning: unknown plugin type %q, skipping\n", src.Type)
			continue
		}
		if mergeErr != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: plugin merge %s: %v\n", src.PluginPath, mergeErr)
		} else if verbose {
			fmt.Fprintf(os.Stderr, "Plugin %s (%s) loaded: %s\n", src.PluginPath, src.Type, src.Path)
		}
	}
}

func runEnricherPlugins(g *model.TraceabilityGraph, config *model.Config, verbose bool) {
	for _, src := range config.Plugins.Enrichers {
		patches, err := plugin.RunEnricher(src.PluginPath, g)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: enricher %s: %v\n", src.PluginPath, err)
			continue
		}
		plugin.ApplyEnrichments(g, patches)
		if verbose {
			fmt.Fprintf(os.Stderr, "Enricher %s applied %d patches\n", src.PluginPath, len(patches))
		}
	}
}

func runLinkerPlugins(builder *graph.Builder, config *model.Config, verbose bool) {
	for _, src := range config.Plugins.Linkers {
		links, err := plugin.RunLinker(src.PluginPath, builder.GetGraph())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: linker %s: %v\n", src.PluginPath, err)
			continue
		}
		g := model.EmptyGraph()
		g.Links = links
		if mergeErr := builder.MergeGraph(g); mergeErr != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: linker merge %s: %v\n", src.PluginPath, mergeErr)
		} else if verbose {
			fmt.Fprintf(os.Stderr, "Linker %s added %d links\n", src.PluginPath, len(links))
		}
	}
}
