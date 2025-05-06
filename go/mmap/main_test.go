package main

import (
	"os/exec"
	"testing"
)

func TestSuccesfullMmap(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "test.txt")

	b, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to run an get output: %v", err)
	}

	expected := "1123"
	if string(b) != expected {
		t.Fatalf("expected: %s, got: %s", expected, string(b))
	}
}
