package report

import (
	_ "embed"
	"bufio"
	"fmt"
	"html/template"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/parlarlax/gocha/internal/coverage"
	"github.com/parlarlax/gocha/internal/testjson"
)

//go:embed template.html
var htmlTemplate string

type Stats struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
}

type SourceLine struct {
	Number int
	Code   string
	Class  string
}

type SourceFile struct {
	Name    string
	Percent float64
	Lines   []SourceLine
}

type Data struct {
	GeneratedAt string
	Version     string
	ProjectName string
	Stats       Stats
	Packages    []testjson.PackageResult
	Coverage    *coverage.Report
	SourceFiles []SourceFile
}

func Generate(w io.Writer, packages []testjson.PackageResult, cov *coverage.Report, projectName string) error {
	stats := Stats{}
	for _, pkg := range packages {
		for _, t := range pkg.Tests {
			stats.Total++
			switch {
			case t.Skipped:
				stats.Skipped++
			case t.Passed:
				stats.Passed++
			default:
				stats.Failed++
			}
		}
	}

	sourceFiles, err := buildSourceFiles(cov)
	if err != nil {
		return err
	}

	data := Data{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Version:     buildVersion(),
		ProjectName: projectName,
		Stats:       stats,
		Packages:    packages,
		Coverage:    cov,
		SourceFiles: sourceFiles,
	}

	funcMap := template.FuncMap{
		"coverClass": func(pct float64) string {
			switch {
			case pct >= 80:
				return "cov-high"
			case pct >= 50:
				return "cov-mid"
			default:
				return "cov-low"
			}
		},
		"coverColor": func(pct float64) string {
			switch {
			case pct >= 80:
				return "#4ade80"
			case pct >= 50:
				return "#facc15"
			default:
				return "#f87171"
			}
		},
		"not": func(b bool) bool { return !b },
		"passRate": func(passed, total int) float64 {
			if total == 0 {
				return 0
			}
			return float64(passed) / float64(total) * 100
		},
		"passCount": func(tests []testjson.TestResult) int {
			n := 0
			for _, t := range tests {
				if t.Passed {
					n++
				}
			}
			return n
		},
		"failCount": func(tests []testjson.TestResult) int {
			n := 0
			for _, t := range tests {
				if !t.Passed && !t.Skipped {
					n++
				}
			}
			return n
		},
		"testStatus": func(t testjson.TestResult) string {
			switch {
			case t.Skipped:
				return "skip"
			case t.Passed:
				return "pass"
			default:
				return "fail"
			}
		},
		"lower": strings.ToLower,
		"isSlow": func(t testjson.TestResult) bool {
			return t.Elapsed >= 1.0
		},
		"sortTests": func(tests []testjson.TestResult) []testjson.TestResult {
			sorted := make([]testjson.TestResult, len(tests))
			copy(sorted, tests)
			sort.SliceStable(sorted, func(i, j int) bool {
				// failed first, then skipped, then passed
				statusOrder := func(t testjson.TestResult) int {
					if !t.Passed && !t.Skipped {
						return 0
					}
					if t.Skipped {
						return 1
					}
					return 2
				}
				return statusOrder(sorted[i]) < statusOrder(sorted[j])
			})
			return sorted
		},
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	return tmpl.Execute(w, data)
}

func buildSourceFiles(cov *coverage.Report) ([]SourceFile, error) {
	if cov == nil {
		return nil, nil
	}

	var result []SourceFile

	for _, fc := range cov.Files {
		// build a map of line -> covered/uncovered
		type lineState int
		const (
			stateNeutral   lineState = 0
			stateCovered   lineState = 1
			stateUncovered lineState = 2
		)

		lineMap := map[int]lineState{}
		for _, p := range fc.Profiles {
			for _, block := range p.Blocks {
				state := stateUncovered
				if block.Count > 0 {
					state = stateCovered
				}
				for ln := block.StartLine; ln <= block.EndLine; ln++ {
					existing := lineMap[ln]
					if existing == stateCovered {
						continue
					}
					lineMap[ln] = state
				}
			}
		}

		// try to read source file from GOPATH/module cache — best effort
		src, err := readSourceFile(fc.FileName)
		if err != nil {
			// skip source view if file not readable
			continue
		}

		var lines []SourceLine
		scanner := bufio.NewScanner(strings.NewReader(src))
		lineNum := 1
		for scanner.Scan() {
			class := ""
			switch lineMap[lineNum] {
			case stateCovered:
				class = "covered"
			case stateUncovered:
				class = "uncovered"
			}
			lines = append(lines, SourceLine{
				Number: lineNum,
				Code:   scanner.Text(),
				Class:  class,
			})
			lineNum++
		}

		result = append(result, SourceFile{
			Name:    fc.FileName,
			Percent: fc.Percent,
			Lines:   lines,
		})
	}

	return result, nil
}

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	return "dev"
}

func readSourceFile(name string) (string, error) {
	// try relative to cwd first
	data, err := os.ReadFile(name)
	if err == nil {
		return string(data), nil
	}
	// strip module prefix: find first path segment that looks like a dir
	parts := strings.Split(name, "/")
	for i := 1; i < len(parts); i++ {
		candidate := strings.Join(parts[i:], "/")
		data, err = os.ReadFile(candidate)
		if err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("source not found: %s", name)
}
