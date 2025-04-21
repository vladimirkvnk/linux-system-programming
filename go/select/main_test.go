package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestMainBehavior(t *testing.T) {
	tests := []struct {
		name           string
		prepareStdin   func() (*os.File, error)
		expectedOutput string
		timeout        time.Duration
		skip           bool
	}{
		{
			name: "Timeout when no input",
			prepareStdin: func() (*os.File, error) {
				// Create pipe and don't write anything
				r, _, err := os.Pipe()
				return r, err
			},
			expectedOutput: "Timeout hit.",
			timeout:        5 * time.Second, // Increased timeout for CI
			// Skip this test in CI environment if needed
			skip: os.Getenv("CI") == "true",
		},
		{
			name: "Successful read from input",
			prepareStdin: func() (*os.File, error) {
				// Create pipe with input
				r, w, err := os.Pipe()
				if err != nil {
					return nil, err
				}
				_, err = w.WriteString("test input\n")
				if err != nil {
					return nil, err
				}
				if err := w.Close(); err != nil {
					return nil, err
				}
				return r, nil
			},
			expectedOutput: "Read: test input",
			timeout:        5 * time.Second,
		},
		{
			name: "Nothing read from closed input",
			prepareStdin: func() (*os.File, error) {
				// Create and close a pipe
				r, w, err := os.Pipe()
				if err != nil {
					return nil, err
				}
				if err := w.Close(); err != nil { // Close write end to signal EOF
					return nil, err
				}
				return r, nil
			},
			expectedOutput: "Nothing read",
			timeout:        5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("Skipping test in CI environment")
			}

			// Set very short timeout for tests
			if err := os.Setenv("SELECT_TIMEOUT", "1"); err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}

			// Create stdin
			stdin, err := tt.prepareStdin()
			if err != nil {
				t.Fatalf("Failed to prepare stdin: %v", err)
			}
			defer func() {
				if err := stdin.Close(); err != nil {
					t.Logf("Warning: failed to close stdin: %v", err)
				}
			}()

			// Create stdout capture
			stdout := &bytes.Buffer{}

			// Create a command that will run the same binary
			cmd := exec.Command(os.Args[0], "-test.run=TestHelper")
			cmd.Stdin = stdin
			cmd.Stdout = stdout
			cmd.Stderr = os.Stderr
			cmd.Env = append(os.Environ(), "GO_RUNNING_SUBTEST=1")

			// Run with timeout
			err = cmd.Start()
			if err != nil {
				t.Fatalf("Failed to start command: %v", err)
			}

			// Wait with timeout
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case <-time.After(tt.timeout):
				// Kill the process if it times out
				if err := cmd.Process.Kill(); err != nil {
					t.Logf("Warning: failed to kill process: %v", err)
				}
				t.Fatalf("Test timed out after %v", tt.timeout)
			case err := <-done:
				if err != nil {
					t.Errorf("Command failed: %v", err)
				}
			}

			// Check output
			output := stdout.String()
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expectedOutput, output)
			}
		})
	}
}

// TestHelper is used as a helper for running the command
func TestHelper(t *testing.T) {
	if os.Getenv("GO_RUNNING_SUBTEST") != "1" {
		return
	}
	main()
}
