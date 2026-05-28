package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/parlarlax/gocha/internal/coverage"
	"github.com/parlarlax/gocha/internal/report"
	"github.com/parlarlax/gocha/internal/testjson"
)

func main() {
	coverFlag := flag.String("cover", "", "path to coverage profile (consumer mode only)")
	outFlag := flag.String("o", "gocha-report.html", "output HTML file")
	titleFlag := flag.String("title", "", "project name shown in report header (default: auto-detect from go.mod)")
	mdFlag := flag.String("md", "", "also write markdown summary to this file (use - for stdout)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `gocha — Go test reporter

Usage:
  gocha [flags] -- [go test flags] <packages>   runner mode
  go test -json ./... | gocha [flags]            consumer mode

Flags:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  gocha -- ./...
  gocha -- -race -timeout 60s ./...
  go test -json -coverprofile=c.out ./... | gocha -cover c.out
  go test -json ./... | gocha -title "My App"
  gocha -- ./... && gocha -md summary.md
`)
	}
	flag.Parse()

	args := flag.Args()

	var (
		packages  []testjson.PackageResult
		covReport *coverage.Report
		duration  time.Duration
		testErr   error
	)

	if len(args) > 0 {
		packages, covReport, duration, testErr = runTests(args)
	} else {
		var err error
		packages, duration, err = testjson.Parse(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gocha: parse test output: %v\n", err)
			os.Exit(1)
		}
		if *coverFlag != "" {
			covReport, err = coverage.Parse(*coverFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "gocha: parse coverage: %v\n", err)
				os.Exit(1)
			}
		}
	}

	title := *titleFlag
	if title == "" {
		title = detectTitle()
	}

	if err := writeReport(*outFlag, packages, covReport, title, duration); err != nil {
		fmt.Fprintf(os.Stderr, "gocha: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "gocha: report written to %s\n", *outFlag)

	if *mdFlag != "" {
		if err := writeMD(*mdFlag, packages, covReport, title, duration); err != nil {
			fmt.Fprintf(os.Stderr, "gocha: write markdown: %v\n", err)
		} else if *mdFlag != "-" {
			fmt.Fprintf(os.Stderr, "gocha: markdown written to %s\n", *mdFlag)
		}
	}

	if testErr != nil {
		os.Exit(1)
	}
}

func runTests(args []string) ([]testjson.PackageResult, *coverage.Report, time.Duration, error) {
	tmpCover, err := os.CreateTemp("", "gocha-cover-*.out")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gocha: create temp file: %v\n", err)
		os.Exit(1)
	}
	tmpCover.Close()
	defer os.Remove(tmpCover.Name())

	goArgs := append([]string{"test", "-json", "-coverprofile=" + tmpCover.Name()}, args...)
	cmd := exec.Command("go", goArgs...)
	cmd.Stderr = os.Stderr

	var buf bytes.Buffer
	cmd.Stdout = &buf

	runErr := cmd.Run()

	packages, duration, err := testjson.Parse(&buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gocha: parse test output: %v\n", err)
		os.Exit(1)
	}

	var covReport *coverage.Report
	if _, err := os.Stat(tmpCover.Name()); err == nil {
		covReport, err = coverage.Parse(tmpCover.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "gocha: parse coverage: %v\n", err)
		}
	}

	return packages, covReport, duration, runErr
}

func writeReport(outFile string, packages []testjson.PackageResult, covReport *coverage.Report, title string, duration time.Duration) error {
	out, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer out.Close()
	return report.Generate(out, packages, covReport, title, duration)
}

func writeMD(path string, packages []testjson.PackageResult, cov *coverage.Report, title string, duration time.Duration) error {
	var w io.Writer
	if path == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	return report.GenerateMarkdown(w, packages, cov, title, duration)
}

func detectTitle() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			mod := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			parts := strings.Split(mod, "/")
			return parts[len(parts)-1]
		}
	}
	return ""
}
