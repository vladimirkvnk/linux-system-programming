package main

import (
	"os/exec"
	"testing"
)

func TestSelectNothingReadFromStdin(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go")
	cmd.Stdin = nil

	b, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to run and get output: %v", err)
	}

	expected := "nothing read\n"
	got := string(b)
	if expected != got {
		t.Fatalf("expected: %s, got: %s", expected, got)
	}
}

func TestReadString(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe failed: %v", err)
	}
	defer func() {
		_ = stdin.Close()
	}()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe failed: %v", err)
	}
	defer func() {
		_ = stdout.Close()
	}()

	_, err = stdin.Write([]byte("simple_string"))
	if err != nil {
		t.Fatalf("failed to write to stdin: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start command: %v", err)
	}

	buf := make([]uint8, 256)

	len, err := stdout.Read(buf)
	if err != nil {
		t.Fatalf("failed to read from stdin: %v", err)
	}

	expected := "read: simple_string\n"
	got := string(buf[0:len])
	if expected != got {
		t.Fatalf("expected: %s, got: %s", expected, got)
	}

	if err := cmd.Wait(); err != nil {
		t.Fatalf("failed to wait: %v", err)
	}
}
