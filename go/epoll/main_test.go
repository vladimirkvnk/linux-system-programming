package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestEpollEchoServer(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("epoll is Linux-only")
	}

	port := getFreePort(t)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	bin := buildServerBinary(t)

	var output bytes.Buffer

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port))
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	t.Cleanup(func() {
		killProcess(t, cmd)
	})

	waitForTCP(t, addr, 3*time.Second, &output)

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("failed to connect: %v\nserver output:\n%s", err, output.String())
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("failed to close connection: %v", err)
		}
	}()

	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set deadline: %v", err)
	}

	msg := []byte("hello epoll\n")

	if _, err := conn.Write(msg); err != nil {
		t.Fatalf("failed to write: %v\nserver output:\n%s", err, output.String())
	}

	got := make([]byte, len(msg))

	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("failed to read echo response: %v\nserver output:\n%s", err, output.String())
	}

	if string(got) != string(msg) {
		t.Fatalf("expected %q, got %q\nserver output:\n%s", msg, got, output.String())
	}
}

func buildServerBinary(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	bin := filepath.Join(dir, "epoll-server")

	cmd := exec.Command("go", "build", "-race", "-o", bin, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build server: %v\n%s", err, string(out))
	}

	return bin
}

func waitForTCP(t *testing.T, addr string, timeout time.Duration, output *bytes.Buffer) {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("server did not start on %s within %s\nserver output:\n%s", addr, timeout, output.String())
}

func getFreePort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	defer func() {
		if err := ln.Close(); err != nil {
			t.Fatalf("failed to close listener: %v", err)
		}
	}()

	return ln.Addr().(*net.TCPAddr).Port
}

func killProcess(t *testing.T, cmd *exec.Cmd) {
	t.Helper()

	if cmd.Process == nil {
		return
	}

	_ = cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
		return
	case <-time.After(500 * time.Millisecond):
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-done
	}
}
