package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

const timeoutSec = 5

func main() {
	fds := []unix.PollFd{
		{
			Fd:     int32(unix.Stdin),
			Events: unix.POLLIN,
		},
		{
			Fd:     int32(unix.Stdout),
			Events: unix.POLLOUT,
		},
	}

	res, err := unix.Poll(fds, timeoutSec*1000)
	if err != nil {
		log.Panicf("Poll error: %s", err)
	}
	if res == 0 {
		fmt.Printf("Timeout hit. %d seconds elapsed.\n", timeoutSec)
	}
	if (fds[0].Revents & unix.POLLIN) != 0 {
		fmt.Println("Stdin is readable")
	}
	if (fds[1].Revents & unix.POLLOUT) != 0 {
		fmt.Println("Stdout is writeable")
	}
}
