package testjson

import (
	"bufio"
	"encoding/json"
	"io"
	"time"
)

type Action string

const (
	ActionRun    Action = "run"
	ActionPass   Action = "pass"
	ActionFail   Action = "fail"
	ActionSkip   Action = "skip"
	ActionOutput Action = "output"
	ActionStart  Action = "start"
)

type Event struct {
	Time    time.Time `json:"Time"`
	Action  Action    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Output  string    `json:"Output"`
	Elapsed float64   `json:"Elapsed"`
}

type TestResult struct {
	Package string
	Name    string
	Passed  bool
	Skipped bool
	Elapsed float64
}

type PackageResult struct {
	Name    string
	Passed  bool
	Skipped bool
	Elapsed float64
	Tests   []TestResult
}

func Parse(r io.Reader) ([]PackageResult, error) {
	packages := map[string]*PackageResult{}
	order := []string{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}

		pkg, ok := packages[e.Package]
		if !ok {
			pkg = &PackageResult{Name: e.Package}
			packages[e.Package] = pkg
			order = append(order, e.Package)
		}

		switch e.Action {
		case ActionPass:
			if e.Test == "" {
				pkg.Passed = true
				pkg.Elapsed = e.Elapsed
			} else {
				pkg.Tests = append(pkg.Tests, TestResult{
					Package: e.Package,
					Name:    e.Test,
					Passed:  true,
					Elapsed: e.Elapsed,
				})
			}
		case ActionFail:
			if e.Test == "" {
				pkg.Passed = false
				pkg.Elapsed = e.Elapsed
			} else {
				pkg.Tests = append(pkg.Tests, TestResult{
					Package: e.Package,
					Name:    e.Test,
					Passed:  false,
					Elapsed: e.Elapsed,
				})
			}
		case ActionSkip:
			if e.Test == "" {
				pkg.Skipped = true
			} else {
				pkg.Tests = append(pkg.Tests, TestResult{
					Package: e.Package,
					Name:    e.Test,
					Skipped: true,
					Elapsed: e.Elapsed,
				})
			}
		}
	}

	results := make([]PackageResult, 0, len(order))
	for _, name := range order {
		results = append(results, *packages[name])
	}
	return results, scanner.Err()
}
