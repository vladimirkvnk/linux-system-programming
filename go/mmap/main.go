package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

func main() {
	if len(os.Args) != 2 {
		panic(fmt.Sprintf("usage: %s <file>\n", os.Args[0]))
	}

	fd, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to open a file: %v", err))
	}

	st, err := fd.Stat()
	if err != nil {
		panic(fmt.Sprintf("failed to stat file: %v", err))
	}

	if !st.Mode().IsRegular() {
		panic(fmt.Sprintf("%s is not a file", os.Args[1]))
	}

	data, err := unix.Mmap(int(fd.Fd()), 0, int(st.Size()), unix.PROT_READ, unix.MAP_SHARED)
	if err != nil {
		panic(fmt.Sprintf("mmap failed: %v", err))
	}

	if err := fd.Close(); err != nil {
		panic(fmt.Sprintf("file close failed: %v", err))
	}

	for _, char := range strings.Fields(string(data)) {
		fmt.Print(char)
	}

	if err := unix.Munmap(data); err != nil {
		panic(fmt.Sprintf("munmap failed: %v", err))
	}

	os.Exit(0)
}
