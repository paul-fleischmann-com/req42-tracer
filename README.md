# req42-tracer — Requirements Tracing Tool

[![Test](https://github.com/paul-fleischmann-com/req42-tracer/actions/workflows/test.yml/badge.svg)](https://github.com/paul-fleischmann-com/req42-tracer/actions/workflows/test.yml)

A Go CLI tool for tracing requirements across AsciiDoc documentation, architecture models (Bausteinsicht), and test specifications with ASPICE PAM 4.0 compliance validation.

## MVP Features

### Core Commands

- **`req42-tracer init`** — Initialize new project with templates (interactive or automated)
- **`req42-tracer trace`** — Display traceability matrix showing requirement → architecture → test links
- **`req42-tracer gaps`** — Analyze and report orphan artifacts and missing implementations
- **`req42-tracer aspice`** — Validate ASPICE PAM 4.0 compliance, show process coverage
- **`req42-tracer validate`** — Check project structure and validate all references

### Architecture

- **AsciiDoc Parser** — Extracts [req], [arch], [test-spec] blocks with attributes
- **Bausteinsicht Integration** — Loads architecture models as JSONC with nested hierarchy
- **Traceability Graph** — Bidirectional dependency graph with explicit and name-based linking
- **ASPICE Checker** — Validates SWE.1-6, SYS.1-3 processes with best practice checks
- **Test Result Parsers** — JUnit XML and go-test JSON support with platform detection
- **Report Generator** — Text, Markdown, and JSON output formats

### Hybrid Test-Result Tracing

Combines three approaches for matching test results to specifications:
1. **Explicit annotations** in test code comments (primary)
2. **Name-based matching** using heuristics (fallback)
3. **Custom metadata** in test reports (optional)

## Coverage Rules (ASPICE PAM 4.0)

req42-tracer enforces the full ASPICE SWE.1–SWE.6 traceability chain. Each block type maps to one ASPICE process:

```asciidoc
[req,id=SWR-001,aspice=SWE.1]
== System shall parse AsciiDoc blocks           ← SWE.1 Software Requirements

[arch,id=comp.parser,req=SWR-001,aspice=SWE.2,impl=src/internal/parser/]
== Parser Component                              ← SWE.2 Architectural Design

[dsn,id=comp.parser.tokenizer,arch=comp.parser,aspice=SWE.3,impl=src/internal/parser/tokenizer.go]
== Tokenizer Unit                                ← SWE.3 Detailed Design

[test-spec,id=TS-UNIT-001,dsn=comp.parser.tokenizer,aspice=SWE.4]
== Unit Test: Tokenizer                          ← SWE.4 Unit Verification

[test-spec,id=TS-INT-001,arch=comp.parser,aspice=SWE.5]
== Integration Test: Parser Component            ← SWE.5 Integration Verification

[test-spec,id=TS-SWR-001,req=SWR-001,aspice=SWE.6]
== SW Qualification Test: Parser extracts all req blocks  ← SWE.6 Qualification Test
```

### Traceability chain

```
[req,SWE.1] ──req=──▶ [arch,SWE.2] ──arch=──▶ [dsn,SWE.3] ──impl=──▶ src/...
     │                     │                        │
     └──req=──▶ [test-spec/SWE.6]                  └──dsn=──▶ [test-spec/SWE.4]
                [test-spec/SWE.5] ◀──arch=──────────┘
```

### Vollständige ASPICE-Traceability-Kette (C-Projekt-Beispiel)

```asciidoc
[req,id=SWR-GPIO-001,aspice=SWE.1]
== System shall initialize GPIO hardware
The HAL shall configure pin direction and pull resistors.

[arch,id=comp.hal.gpio,req=SWR-GPIO-001,aspice=SWE.2,
 impl=src/hal/gpio.c:gpio_init:23]
== GPIO Hardware Abstraction Layer

[test-spec,id=TS-GPIO-001,req=SWR-GPIO-001,aspice=SWE.4-BP4]
== Test: gpio_init configures pin direction correctly
```

Im C-Quellcode (`src/hal/gpio.c`):
```c
/**
 * @req SWR-GPIO-001
 * @arch comp.hal.gpio
 * @aspice SWE.3
 */
void gpio_init(uint8_t pin, gpio_direction_t dir) { ... }
```

### Coverage rules

| Was | Wo | ASPICE | Pflicht |
|---|---|---|---|
| `req=SWR-001` | `[arch]`-Block | SWE.2 BP4 | ✅ |
| `impl=file.c:func:line` | `[arch]`-Block | SWE.3 BP5 | ✅ |
| `req=SWR-001` | `[test-spec]`-Block | SWE.4 BP4 | ✅ |
| TestResult PASS | JUnit/Unity/CUnit XML | SWE.4 BP5 | ✅ |
| `derives=SYS-001` | `[req]`-Block | SWE.1 BP5 | optional |

> **Wichtig:** `impl=` gehört zum **`[arch]`-Block**, nicht zum `[req]`-Block.

#### impl= Formate

| Format | Beispiel | Bedeutung |
|---|---|---|
| Go-Package | `src/internal/parser` | Go-Paketpfad |
| Datei | `src/hal/gpio.c` | Quelldatei |
| Datei:Funktion | `src/hal/gpio.c:gpio_init` | Funktion in Datei |
| Datei:Funktion:Zeile | `src/hal/gpio.c:gpio_init:23` | Exakte Zeile (klickbar im HTML-Report) |

### Gap messages

| Message | Meaning | Fix |
|---|---|---|
| `orphan requirement` | `[req]` has no `[arch]` with `req=` | Add `[arch,req=SWR-XXX]` |
| `missing impl` | `[arch]`/`[dsn]` has no `impl=` | Add `impl=src/...` |
| `untested architecture element (SWE.5)` | `[arch]` has no `[test-spec]` with `arch=` | Add Integration Test |
| `orphan design element (SWE.3)` | `[dsn]` has no `arch=` parent | Add `arch=` to `[dsn]` |
| `untested detailed design (SWE.4)` | `[dsn]` has no `[test-spec]` with `dsn=` | Add Unit Test |
| `untraced test result` | `[test-spec]` has no matching TestResult in JUnit/go-test | Check CI output |

## Quick Start

```bash
# Initialize a project
req42-tracer init --name=MyProject --module=github.com/user/myproject --interactive=false

# Validate structure
req42-tracer validate

# Show traceability
req42-tracer trace --format=markdown

# Analyze gaps
req42-tracer gaps

# Check ASPICE compliance
req42-tracer aspice
```

## Project Structure

```
internal/
  ├── model/       # Type definitions, config loader
  ├── parser/      # AsciiDoc + Bausteinsicht parsers
  ├── graph/       # Traceability builder and analyzer
  ├── aspice/      # ASPICE PAM 4.0 registry and checker
  ├── testresult/  # JUnit/go-test JSON loaders
  ├── report/      # CLI report generators
  └── templates/   # Project initialization templates

cmd/req42-tracer/
  ├── root.go      # CLI root command
  ├── init.go      # Project initialization
  ├── trace.go     # Traceability matrix
  ├── gaps.go      # Gap analysis
  ├── aspice.go    # ASPICE compliance
  └── validate.go  # Validation
```

## Roadmap (Phase 2)

- [ ] HTML Report Generator with interactive Graph/Matrix/ASPICE views
- [ ] Watch Mode with live-reload browser preview
- [ ] LSP Server for IDE integration (autocomplete, hover, diagnostics)
- [ ] Advanced filtering and search
- [ ] Custom validation rules

## Implementation Status

✅ Complete:
- Core type system
- AsciiDoc and Bausteinsicht parsing
- Traceability graph construction
- ASPICE PAM 4.0 validation
- Test result parsing
- CLI reports (text/markdown/json)
- Project initialization

⏳ Pending:
- HTML report generation
- Watch mode
- LSP server
