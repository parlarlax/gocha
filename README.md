# gocha

A single-command Go test reporter that combines **unit test results** and **coverage** into one self-contained HTML file — no external dependencies, no CI plugins required.

Inspired by [Mocha](https://mochajs.org/)'s beautiful HTML reports, brought to Go.

## Why

`go test -v` shows test results in the terminal. `go tool cover -html` shows coverage in a browser. Neither shows both. gocha fixes that.

|                     | `go test -v` | `go tool cover -html` | **gocha** |
| ------------------- | ------------ | --------------------- | --------- |
| Test case results   | ✅           | ❌                    | ✅        |
| Coverage % per file | ❌           | ✅                    | ✅        |
| Source code view    | ❌           | ✅                    | ✅        |
| Single HTML file    | ❌           | ✅                    | ✅        |

## Install

```bash
go install github.com/parlarlax/gocha/cmd/gocha@latest
```

## Usage

```bash
go test -v -json ./... | gocha -cover coverage.out
```

This generates `gocha-report.html` in the current directory.

**Full example:**

```bash
go test -v -json -coverprofile=coverage.out ./... | gocha -cover coverage.out -o report.html
```

**Flags:**

| Flag     | Default             | Description                                       |
| -------- | ------------------- | ------------------------------------------------- |
| `-cover` | _(none)_            | Path to coverage profile (`-coverprofile` output) |
| `-o`     | `gocha-report.html` | Output HTML file path                             |

## Report

The generated HTML has three tabs:

- **Tests** — all test cases with PASS / FAIL / SKIP badge and duration
- **Coverage** — per-file coverage percentage with progress bar
- **Source** — source code highlighted green (covered) / red (uncovered)

> Run `gocha` from your **project root** so source file paths in the coverage profile can be resolved.

## Requirements

- Go 1.21+
