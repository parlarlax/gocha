package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/parlarlax/gocha/internal/coverage"
	"github.com/parlarlax/gocha/internal/report"
	"github.com/parlarlax/gocha/internal/testjson"
)

func main() {
	coverFlag := flag.String("cover", "", "path to coverage profile (coverprofile)")
	outFlag := flag.String("o", "gocha-report.html", "output HTML file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: go test -json ./... | gocha [flags]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	packages, err := testjson.Parse(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gocha: parse test output: %v\n", err)
		os.Exit(1)
	}

	var covReport *coverage.Report
	if *coverFlag != "" {
		covReport, err = coverage.Parse(*coverFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gocha: parse coverage: %v\n", err)
			os.Exit(1)
		}
	}

	out, err := os.Create(*outFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gocha: create output: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	if err := report.Generate(out, packages, covReport); err != nil {
		fmt.Fprintf(os.Stderr, "gocha: generate report: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "gocha: report written to %s\n", *outFlag)
}
