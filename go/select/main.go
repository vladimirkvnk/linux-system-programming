package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

const (
	defaultTimeoutSec = 3
	bufLen            = 1024
)

func main() {
	readfs := unix.FdSet{}
	readfs.Zero()
	readfs.Set(unix.Stdin)

	timeVal := unix.Timeval{
		Sec: defaultTimeoutSec,
	}

	result, err := unix.Select(unix.Stdin+1, &readfs, nil, nil, &timeVal)
	if err != nil {
		log.Fatalf("select error: %v", err)
	}

	if result == 0 {
		fmt.Println("hit timeout")
		return
	}

	if !readfs.IsSet(unix.Stdin) {
		fmt.Println("this should never happen!")
	}

	buf := make([]uint8, bufLen)

	len, err := unix.Read(unix.Stdin, buf)
	if err != nil {
		log.Fatalf("read from stdin: %v", err)
	}

	if len == 0 {
		fmt.Println("nothing read")
		return
	}

	if _, err := fmt.Printf("read: %s\n", string(buf[0:len])); err != nil {
		log.Fatalf("failed to convert bytes to string after read: %b", buf[0:len])
	}
}
