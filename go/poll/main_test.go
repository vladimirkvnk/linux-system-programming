package main

import (
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// testingPollFd is a helper function to run Poll with given file descriptors
func testingPollFd(fds []unix.PollFd, timeoutSec int) (int, error) {
	return unix.Poll(fds, timeoutSec*1000)
}

// mockFd is a helper struct to simulate file descriptor behavior for testing
type mockFd struct {
	file     *os.File
	readable bool
	writable bool
}

// createMockFds creates a pipe that can be used for testing Poll
func createMockFds() (mockFd, mockFd, func(), error) {
	r, w, err := os.Pipe()
	if err != nil {
		return mockFd{}, mockFd{}, nil, err
	}

	cleanup := func() {
		_ = r.Close()
		_ = w.Close()
	}

	return mockFd{file: r, readable: false, writable: false},
		mockFd{file: w, readable: false, writable: true},
		cleanup, nil
}

// TestPollTimeout tests the timeout behavior of Poll
func TestPollTimeout(t *testing.T) {
	r, _, cleanup, err := createMockFds()
	if err != nil {
		t.Fatalf("Failed to create mock fds: %v", err)
	}
	defer cleanup()

	// Create a poll with a really short timeout
	fds := []unix.PollFd{
		{
			Fd:     int32(r.file.Fd()),
			Events: unix.POLLIN,
		},
	}

	// Run poll with a very short timeout
	start := time.Now()
	res, err := testingPollFd(fds, 1) // 1 second timeout
	elapsed := time.Since(start)

	// Validate timeout results
	if err != nil {
		t.Fatalf("Poll returned error: %v", err)
	}
	if res != 0 {
		t.Errorf("Expected timeout (result 0), got %d", res)
	}
	if elapsed < 950*time.Millisecond {
		t.Errorf("Timeout occurred too quickly: %v", elapsed)
	}
}

// TestPollReadable tests if Poll correctly identifies readable file descriptors
func TestPollReadable(t *testing.T) {
	r, w, cleanup, err := createMockFds()
	if err != nil {
		t.Fatalf("Failed to create mock fds: %v", err)
	}
	defer cleanup()

	// Write something to make the pipe readable
	_, err = w.file.Write([]byte("test data"))
	if err != nil {
		t.Fatalf("Failed to write to pipe: %v", err)
	}

	// Create a poll for readable events
	fds := []unix.PollFd{
		{
			Fd:     int32(r.file.Fd()),
			Events: unix.POLLIN,
		},
	}

	// Poll should return immediately as the pipe is readable
	res, err := testingPollFd(fds, 5)
	if err != nil {
		t.Fatalf("Poll returned error: %v", err)
	}
	if res != 1 {
		t.Errorf("Expected 1 ready fd, got %d", res)
	}
	if (fds[0].Revents & unix.POLLIN) == 0 {
		t.Error("Expected POLLIN event but got none")
	}
}

// TestPollWritable tests if Poll correctly identifies writable file descriptors
func TestPollWritable(t *testing.T) {
	_, w, cleanup, err := createMockFds()
	if err != nil {
		t.Fatalf("Failed to create mock fds: %v", err)
	}
	defer cleanup()

	// Create a poll for writable events on the write end
	fds := []unix.PollFd{
		{
			Fd:     int32(w.file.Fd()),
			Events: unix.POLLOUT,
		},
	}

	// Poll should return immediately as the pipe is writable
	res, err := testingPollFd(fds, 5)
	if err != nil {
		t.Fatalf("Poll returned error: %v", err)
	}
	if res != 1 {
		t.Errorf("Expected 1 ready fd, got %d", res)
	}
	if (fds[0].Revents & unix.POLLOUT) == 0 {
		t.Error("Expected POLLOUT event but got none")
	}
}

// TestPollMultipleFds tests handling multiple file descriptors simultaneously
func TestPollMultipleFds(t *testing.T) {
	r, w, cleanup, err := createMockFds()
	if err != nil {
		t.Fatalf("Failed to create mock fds: %v", err)
	}
	defer cleanup()

	// Write something to make the pipe readable
	_, err = w.file.Write([]byte("test data"))
	if err != nil {
		t.Fatalf("Failed to write to pipe: %v", err)
	}

	// Create a poll with both read and write fds
	fds := []unix.PollFd{
		{
			Fd:     int32(r.file.Fd()),
			Events: unix.POLLIN,
		},
		{
			Fd:     int32(w.file.Fd()),
			Events: unix.POLLOUT,
		},
	}

	// Poll should detect both events
	res, err := testingPollFd(fds, 5)
	if err != nil {
		t.Fatalf("Poll returned error: %v", err)
	}
	if res != 2 {
		t.Errorf("Expected 2 ready fds, got %d", res)
	}
	if (fds[0].Revents & unix.POLLIN) == 0 {
		t.Error("Expected POLLIN event but got none")
	}
	if (fds[1].Revents & unix.POLLOUT) == 0 {
		t.Error("Expected POLLOUT event but got none")
	}
}

// TestPollError tests error handling in Poll
func TestPollError(t *testing.T) {
	// Create and close a file to get an invalid file descriptor
	f, err := os.Open("/dev/null")
	if err != nil {
		t.Fatalf("Failed to open /dev/null: %v", err)
	}

	// Get the fd and then close it to make it invalid
	fd := f.Fd()
	_ = f.Close()

	// A closed file descriptor should cause an error when polled
	fds := []unix.PollFd{
		{
			Fd:     int32(fd),
			Events: unix.POLLIN,
		},
	}

	// On some platforms, Poll might not return an error for all invalid fds
	// So instead we'll check that the revents includes POLLNVAL
	_, err = testingPollFd(fds, 1)

	// Either we should get an error OR the POLLNVAL flag should be set
	if err == nil && (fds[0].Revents&unix.POLLNVAL) == 0 {
		t.Error("Expected either an error or POLLNVAL flag for invalid file descriptor")
	}
}
