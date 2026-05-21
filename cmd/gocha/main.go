package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/parlarlax/gocha/internal/coverage"
	"github.com/parlarlax/gocha/internal/report"
	"github.com/parlarlax/gocha/internal/testjson"
)

func main() {
	coverFlag := flag.String("cover", "", "path to coverage profile (consumer mode only)")
	outFlag := flag.String("o", "gocha-report.html", "output HTML file")
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
`)
	}
	flag.Parse()

	args := flag.Args() // positional args after flags (or after --)

	var (
		packages  []testjson.PackageResult
		covReport *coverage.Report
		testErr   error
	)

	if len(args) > 0 {
		// runner mode: gocha [flags] -- [go test flags] <packages>
		packages, covReport, testErr = runTests(args, *outFlag)
	} else {
		// consumer mode: pipe stdin
		var err error
		packages, err = testjson.Parse(os.Stdin)
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

	if err := writeReport(*outFlag, packages, covReport); err != nil {
		fmt.Fprintf(os.Stderr, "gocha: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "gocha: report written to %s\n", *outFlag)

	if testErr != nil {
		os.Exit(1)
	}
}

func runTests(args []string, _ string) ([]testjson.PackageResult, *coverage.Report, error) {
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

	packages, err := testjson.Parse(&buf)
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

	return packages, covReport, runErr
}

func writeReport(outFile string, packages []testjson.PackageResult, covReport *coverage.Report) error {
	out, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer out.Close()
	return report.Generate(out, packages, covReport)
}
