package testjson_test

import (
	"os"
	"testing"

	"github.com/parlarlax/gocha/internal/testjson"
)

func TestParse_PassingTests(t *testing.T) {
	f, err := os.Open("testdata/pass.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	pkgs, err := testjson.Parse(f)
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

	pkgs, err := testjson.Parse(f)
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
