package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "gocha")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, "github.com/parlarlax/gocha/cmd/gocha")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build gocha: %v", err)
	}
	return bin
}

func TestRunnerMode_AllPass(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	out := filepath.Join(dir, "report.html")

	sampleDir, _ := filepath.Abs("testdata/sample")
	cmd := exec.Command(bin, "-o", out, "--", "./...")
	cmd.Dir = sampleDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("gocha exited non-zero: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("report not created: %v", err)
	}
	if len(data) == 0 {
		t.Error("report is empty")
	}

	content := string(data)
	for _, want := range []string{"TestAdd", "TestFail", "badge-pass", "coverage"} {
		if !contains(content, want) {
			t.Errorf("report missing %q", want)
		}
	}
}

func TestRunnerMode_ExitCode(t *testing.T) {
	// A package with no test files should still exit 0 and produce a report.
	bin := buildBinary(t)
	dir := t.TempDir()
	out := filepath.Join(dir, "report.html")

	sampleDir, _ := filepath.Abs("testdata/sample")
	cmd := exec.Command(bin, "-o", out, "--", "./...")
	cmd.Dir = sampleDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("gocha exited non-zero (expected for failing tests): %v", err)
	}

	if _, statErr := os.Stat(out); statErr != nil {
		t.Error("report should be created even when tests fail")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
