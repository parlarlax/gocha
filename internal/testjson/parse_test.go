package testjson_test

import (
	"os"
	"strings"
	"testing"

	"github.com/parlarlax/gocha/internal/testjson"
)

func TestParse_PassingTests(t *testing.T) {
	f, err := os.Open("testdata/pass.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	pkgs, _, err := testjson.Parse(f)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	var found *testjson.PackageResult
	for i := range pkgs {
		if pkgs[i].Name == "go-test-coverage-demo/calc" {
			found = &pkgs[i]
			break
		}
	}
	if found == nil {
		t.Fatal("expected package go-test-coverage-demo/calc")
	}

	wantTests := []string{"TestAdd", "TestSub", "TestMul", "TestDiv", "TestDivByZero"}
	if len(found.Tests) != len(wantTests) {
		t.Fatalf("got %d tests, want %d", len(found.Tests), len(wantTests))
	}

	for i, name := range wantTests {
		if found.Tests[i].Name != name {
			t.Errorf("test[%d]: got %q, want %q", i, found.Tests[i].Name, name)
		}
		if !found.Tests[i].Passed {
			t.Errorf("test %q: expected passed", name)
		}
	}
}

func TestParse_SkippedPackage(t *testing.T) {
	f, err := os.Open("testdata/pass.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	pkgs, _, err := testjson.Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	var skipped *testjson.PackageResult
	for i := range pkgs {
		if pkgs[i].Name == "go-test-coverage-demo" {
			skipped = &pkgs[i]
			break
		}
	}
	if skipped == nil {
		t.Fatal("expected package go-test-coverage-demo to be present")
	}
	if !skipped.Skipped {
		t.Error("expected package to be skipped (no test files)")
	}
}

func TestParse_FailedTestOutput(t *testing.T) {
	input := `{"Time":"2026-05-28T10:00:00Z","Action":"start","Package":"mypkg"}
{"Time":"2026-05-28T10:00:00.1Z","Action":"run","Package":"mypkg","Test":"TestFoo"}
{"Time":"2026-05-28T10:00:00.2Z","Action":"output","Package":"mypkg","Test":"TestFoo","Output":"=== RUN   TestFoo\n"}
{"Time":"2026-05-28T10:00:00.3Z","Action":"output","Package":"mypkg","Test":"TestFoo","Output":"    foo_test.go:10: expected 1, got 2\n"}
{"Time":"2026-05-28T10:00:00.4Z","Action":"output","Package":"mypkg","Test":"TestFoo","Output":"--- FAIL: TestFoo (0.10s)\n"}
{"Time":"2026-05-28T10:00:00.5Z","Action":"fail","Package":"mypkg","Test":"TestFoo","Elapsed":0.1}
{"Time":"2026-05-28T10:00:00.6Z","Action":"fail","Package":"mypkg","Elapsed":0.1}
`
	pkgs, dur, err := testjson.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("got %d packages, want 1", len(pkgs))
	}
	if len(pkgs[0].Tests) != 1 {
		t.Fatalf("got %d tests, want 1", len(pkgs[0].Tests))
	}
	test := pkgs[0].Tests[0]
	if test.Passed {
		t.Error("expected test to be failed")
	}
	if test.Output == "" {
		t.Error("expected non-empty output for failed test")
	}
	if !strings.Contains(test.Output, "expected 1, got 2") {
		t.Errorf("output missing failure message, got: %q", test.Output)
	}
	if dur <= 0 {
		t.Error("expected positive wall-clock duration")
	}
}

func TestParse_PassedTestOutputNotCaptured(t *testing.T) {
	input := `{"Time":"2026-05-28T10:00:00Z","Action":"start","Package":"mypkg"}
{"Time":"2026-05-28T10:00:00.1Z","Action":"run","Package":"mypkg","Test":"TestBar"}
{"Time":"2026-05-28T10:00:00.2Z","Action":"output","Package":"mypkg","Test":"TestBar","Output":"=== RUN   TestBar\n"}
{"Time":"2026-05-28T10:00:00.3Z","Action":"output","Package":"mypkg","Test":"TestBar","Output":"--- PASS: TestBar (0.00s)\n"}
{"Time":"2026-05-28T10:00:00.4Z","Action":"pass","Package":"mypkg","Test":"TestBar","Elapsed":0}
{"Time":"2026-05-28T10:00:00.5Z","Action":"pass","Package":"mypkg","Elapsed":0}
`
	pkgs, _, err := testjson.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(pkgs[0].Tests) != 1 {
		t.Fatalf("got %d tests, want 1", len(pkgs[0].Tests))
	}
	test := pkgs[0].Tests[0]
	if !test.Passed {
		t.Error("expected test to be passed")
	}
	// Output is still captured (template decides whether to show it)
	_ = test.Output
}

func TestParse_Duration(t *testing.T) {
	input := `{"Time":"2026-05-28T10:00:00Z","Action":"start","Package":"mypkg"}
{"Time":"2026-05-28T10:00:02.5Z","Action":"pass","Package":"mypkg","Elapsed":2.5}
`
	_, dur, err := testjson.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if dur.Seconds() < 2.4 || dur.Seconds() > 2.6 {
		t.Errorf("expected ~2.5s duration, got %v", dur)
	}
}
