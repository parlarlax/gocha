package report

import (
	"fmt"
	"io"
	"time"

	"github.com/parlarlax/gocha/internal/coverage"
	"github.com/parlarlax/gocha/internal/testjson"
)

const maxFailedMD = 50

func GenerateMarkdown(w io.Writer, packages []testjson.PackageResult, cov *coverage.Report, title string, duration time.Duration) error {
	stats := Stats{}
	var failed []testjson.TestResult
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
				failed = append(failed, t)
			}
		}
	}

	if stats.Failed > 0 {
		fmt.Fprintf(w, "## ❌ %d/%d passed", stats.Passed, stats.Total)
	} else {
		fmt.Fprintf(w, "## ✅ %d/%d passed", stats.Passed, stats.Total)
	}
	if d := formatDuration(duration); d != "" {
		fmt.Fprintf(w, " · %s", d)
	}
	if cov != nil {
		fmt.Fprintf(w, " · %.1f%% coverage", cov.Total)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "| Passed | Failed | Skipped | Total |")
	fmt.Fprintln(w, "|--------|--------|---------|-------|")
	fmt.Fprintf(w, "| %d | %d | %d | %d |\n\n", stats.Passed, stats.Failed, stats.Skipped, stats.Total)

	if len(failed) > 0 {
		fmt.Fprintln(w, "<details>")
		fmt.Fprintf(w, "<summary>%d failed test(s)</summary>\n\n", len(failed))
		for i, t := range failed {
			if i >= maxFailedMD {
				fmt.Fprintf(w, "_...and %d more_\n", len(failed)-maxFailedMD)
				break
			}
			fmt.Fprintf(w, "- `%s/%s` — %.3fs\n", t.Package, t.Name, t.Elapsed)
		}
		fmt.Fprintf(w, "\n</details>\n\n")
	}

	if cov != nil {
		fmt.Fprintf(w, "**Coverage: %.1f%%**\n\n", cov.Total)
		if len(cov.Files) > 0 {
			fmt.Fprintln(w, "| File | Coverage | Statements |")
			fmt.Fprintln(w, "|------|----------|------------|")
			for _, f := range cov.Files {
				fmt.Fprintf(w, "| `%s` | %.1f%% | %d/%d |\n", f.FileName, f.Percent, f.Covered, f.Total)
			}
			fmt.Fprintln(w)
		}
	}

	return nil
}
