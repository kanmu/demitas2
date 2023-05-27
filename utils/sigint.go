package utils

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/atomic"
)

func TrapInt(proc func() error, teardown0 func()) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Ignore(syscall.SIGURG)
	signal.Ignore(syscall.SIGWINCH)
	stopped := atomic.NewBool(false)

	teardown := func() {
		if stopped.Load() {
			return
		}

		stopped.Store(true)
		teardown0()
	}

	defer teardown()

	go func() {
		<-sig
		teardown()
		os.Exit(130)
	}()

	return proc()
}
