package report

import (
	_ "embed"
	"bufio"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/parlarlax/gocha/internal/coverage"
	"github.com/parlarlax/gocha/internal/testjson"
)

//go:embed template.html
var htmlTemplate string

const topSlowestN = 10

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
	Duration    string
	GitBranch   string
	GitCommit   string
	GitDirty    bool
	Packages    []testjson.PackageResult
	Coverage    *coverage.Report
	SourceFiles []SourceFile
	Slowest     []testjson.TestResult
}

func Generate(w io.Writer, packages []testjson.PackageResult, cov *coverage.Report, projectName string, duration time.Duration) error {
	stats := Stats{}
	var allTests []testjson.TestResult
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
			if !t.Skipped {
				allTests = append(allTests, t)
			}
		}
	}

	sort.Slice(allTests, func(i, j int) bool {
		return allTests[i].Elapsed > allTests[j].Elapsed
	})
	if len(allTests) > topSlowestN {
		allTests = allTests[:topSlowestN]
	}

	sourceFiles, err := buildSourceFiles(cov)
	if err != nil {
		return err
	}

	branch, commit, dirty := gitInfo()

	data := Data{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Version:     buildVersion(),
		ProjectName: projectName,
		Stats:       stats,
		Duration:    formatDuration(duration),
		GitBranch:   branch,
		GitCommit:   commit,
		GitDirty:    dirty,
		Packages:    packages,
		Coverage:    cov,
		SourceFiles: sourceFiles,
		Slowest:     allTests,
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
		"fmtElapsed": func(elapsed float64) string {
			if elapsed < 0.001 {
				return "< 1ms"
			}
			if elapsed < 1.0 {
				return fmt.Sprintf("%.0fms", elapsed*1000)
			}
			return fmt.Sprintf("%.3fs", elapsed)
		},
		"lower": strings.ToLower,
		"isSlow": func(t testjson.TestResult) bool {
			return t.Elapsed >= 1.0
		},
		"sortTests": func(tests []testjson.TestResult) []testjson.TestResult {
			sorted := make([]testjson.TestResult, len(tests))
			copy(sorted, tests)
			sort.SliceStable(sorted, func(i, j int) bool {
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

func gitInfo() (branch, commit string, dirty bool) {
	branch = runGit("rev-parse", "--abbrev-ref", "HEAD")
	commit = runGit("rev-parse", "--short", "HEAD")
	dirty = runGit("status", "--porcelain") != ""
	return
}

func runGit(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	m := int(d.Minutes())
	s := d.Seconds() - float64(m)*60
	return fmt.Sprintf("%dm %.2fs", m, s)
}

func buildSourceFiles(cov *coverage.Report) ([]SourceFile, error) {
	if cov == nil {
		return nil, nil
	}

	var result []SourceFile

	for _, fc := range cov.Files {
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

		src, err := readSourceFile(fc.FileName)
		if err != nil {
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
	data, err := os.ReadFile(name)
	if err == nil {
		return string(data), nil
	}
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
