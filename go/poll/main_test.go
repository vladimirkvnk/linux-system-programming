package main

import (
	"os/exec"
	"testing"
)

func TestStdoutWritable(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go")

	cl, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to pipe stdin: %v", err)
	}
	defer func() {
		_ = cl.Close()
	}()

	b, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to run an get output: %v", err)
	}

	expected := "Stdout is writeable\n"
	if string(b) != expected {
		t.Fatalf("expected: %s, got: %s", expected, string(b))
	}
}

func TestStoudWritableStdinReadable(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go")

	b, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to run an get output: %v", err)
	}

	expected := "Stdin is readable\nStdout is writeable\n"
	if string(b) != expected {
		t.Fatalf("expected: %s, got: %s", expected, string(b))
	}
}
