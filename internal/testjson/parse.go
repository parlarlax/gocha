package testjson

import (
	"bufio"
	"bytes"
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

const maxOutputBytes = 16 * 1024

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
	Output  string
}

type PackageResult struct {
	Name    string
	Passed  bool
	Skipped bool
	Elapsed float64
	Tests   []TestResult
}

func Parse(r io.Reader) ([]PackageResult, time.Duration, error) {
	packages := map[string]*PackageResult{}
	order := []string{}
	outBufs := map[string][]byte{} // "pkg\x00test" -> buffered output

	var firstTime, lastTime time.Time

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}

		if !e.Time.IsZero() {
			if firstTime.IsZero() {
				firstTime = e.Time
			}
			lastTime = e.Time
		}

		pkg, ok := packages[e.Package]
		if !ok {
			pkg = &PackageResult{Name: e.Package}
			packages[e.Package] = pkg
			order = append(order, e.Package)
		}

		switch e.Action {
		case ActionOutput:
			if e.Test != "" {
				key := e.Package + "\x00" + e.Test
				buf := append(outBufs[key], []byte(e.Output)...)
				if len(buf) > maxOutputBytes {
					buf = buf[len(buf)-maxOutputBytes:]
					if idx := bytes.IndexByte(buf, '\n'); idx >= 0 {
						buf = buf[idx+1:]
					}
				}
				outBufs[key] = buf
			}
		case ActionPass:
			if e.Test == "" {
				pkg.Passed = true
				pkg.Elapsed = e.Elapsed
			} else {
				key := e.Package + "\x00" + e.Test
				pkg.Tests = append(pkg.Tests, TestResult{
					Package: e.Package,
					Name:    e.Test,
					Passed:  true,
					Elapsed: e.Elapsed,
					Output:  string(outBufs[key]),
				})
				delete(outBufs, key)
			}
		case ActionFail:
			if e.Test == "" {
				pkg.Passed = false
				pkg.Elapsed = e.Elapsed
			} else {
				key := e.Package + "\x00" + e.Test
				pkg.Tests = append(pkg.Tests, TestResult{
					Package: e.Package,
					Name:    e.Test,
					Passed:  false,
					Elapsed: e.Elapsed,
					Output:  string(outBufs[key]),
				})
				delete(outBufs, key)
			}
		case ActionSkip:
			if e.Test == "" {
				pkg.Skipped = true
			} else {
				key := e.Package + "\x00" + e.Test
				pkg.Tests = append(pkg.Tests, TestResult{
					Package: e.Package,
					Name:    e.Test,
					Skipped: true,
					Elapsed: e.Elapsed,
					Output:  string(outBufs[key]),
				})
				delete(outBufs, key)
			}
		}
	}

	var duration time.Duration
	if !firstTime.IsZero() && !lastTime.IsZero() {
		duration = lastTime.Sub(firstTime)
	}

	results := make([]PackageResult, 0, len(order))
	for _, name := range order {
		results = append(results, *packages[name])
	}
	return results, duration, scanner.Err()
}
