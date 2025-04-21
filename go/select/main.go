package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

const (
	defaultTimeout = 1 // seconds
	bufLen         = 1024
)

func main() {
	// Get timeout from environment for testing
	timeoutSec := defaultTimeout
	if val, err := strconv.Atoi(os.Getenv("SELECT_TIMEOUT")); err == nil && val > 0 {
		timeoutSec = val
	}

	// Check for CI environment - use a shorter timeout
	if os.Getenv("CI") == "true" {
		timeoutSec = 1 // Always use 1 second in CI
	}

	stdinFd := int(os.Stdin.Fd())

	// Set stdin to non-blocking mode for testing
	if err := unix.SetNonblock(stdinFd, true); err != nil {
		log.Printf("Failed to set non-blocking mode: %v", err)
		return
	}
	defer func() {
		if err := unix.SetNonblock(stdinFd, false); err != nil {
			log.Printf("Failed to restore blocking mode: %v", err)
		}
	}()

	readfs := unix.FdSet{}
	readfs.Zero()
	readfs.Set(stdinFd)

	timeVal := unix.Timeval{
		Sec:  int64(timeoutSec),
		Usec: 0,
	}

	res, err := unix.Select(stdinFd+1, &readfs, nil, nil, &timeVal)
	if err != nil {
		log.Printf("Select error: %v", err)
		return
	}
	if res == 0 {
		fmt.Printf("Timeout hit. %d seconds elapsed.\n", timeoutSec)
		return
	}

	if !readfs.IsSet(stdinFd) {
		log.Println("Unexpected state: stdin not ready")
		return
	}

	buf := make([]uint8, bufLen)
	len, err := unix.Read(stdinFd, buf)
	if err != nil {
		log.Printf("Read error: %v", err)
		return
	}
	if len == 0 {
		fmt.Println("Nothing read")
		return
	}

	str := string(buf[:len])
	if _, err := fmt.Printf("Read: %s", str); err != nil {
		log.Printf("Print error: %v", err)
	}
}
