package main

import (
	"log"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

const PR_SET_CHILD_SUBREAPER = 36

func initSubreap() {
	if err := unix.Prctl(PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		log.Printf("set subreap failed: %v\n", err)
		return
	}

	// ignore SIGCHLD to reap children automatically
	signal.Ignore(syscall.SIGCHLD)
}
