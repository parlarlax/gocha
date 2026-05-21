package coverage_test

import (
	"testing"

	"github.com/parlarlax/gocha/internal/coverage"
)

func TestParse(t *testing.T) {
	r, err := coverage.Parse("testdata/coverage.out")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(r.Files) != 2 {
		t.Fatalf("got %d files, want 2", len(r.Files))
	}

	byName := map[string]coverage.FileCoverage{}
	for _, f := range r.Files {
		byName[f.FileName] = f
	}

	calc, ok := byName["go-test-coverage-demo/calc/calc.go"]
	if !ok {
		t.Fatal("missing calc/calc.go")
	}
	if calc.Percent != 100.0 {
		t.Errorf("calc coverage: got %.1f%%, want 100.0%%", calc.Percent)
	}

	main, ok := byName["go-test-coverage-demo/main.go"]
	if !ok {
		t.Fatal("missing main.go")
	}
	if main.Percent != 0.0 {
		t.Errorf("main coverage: got %.1f%%, want 0.0%%", main.Percent)
	}

	// total is weighted average: 6 covered out of 13 statements
	if r.Total < 45 || r.Total > 48 {
		t.Errorf("total coverage: got %.1f%%, want ~46%%", r.Total)
	}
}

func TestParse_NotFound(t *testing.T) {
	_, err := coverage.Parse("testdata/nonexistent.out")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
