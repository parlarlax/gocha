package sample_test

import (
	"sample"
	"testing"
)

func TestAdd(t *testing.T) {
	if got := sample.Add(1, 2); got != 3 {
		t.Errorf("Add(1,2) = %d, want 3", got)
	}
}

func TestFail(t *testing.T) {
	if sample.Fail() {
		t.Error("expected false")
	}
}
