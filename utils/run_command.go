package utils

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"go.uber.org/atomic"
)

func RunCommand(cmdWithArgs []string, silent bool) (string, string, bool, error) {
	cmd := exec.Command(cmdWithArgs[0], cmdWithArgs[1:]...)
	interrupted := atomic.NewBool(false)

	outReader, err := cmd.StdoutPipe()

	if err != nil {
		return "", "", false, err
	}

	errReader, err := cmd.StderrPipe()

	if err != nil {
		return "", "", false, err
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig)
	signal.Ignore(syscall.SIGURG)
	signal.Ignore(syscall.SIGWINCH)
	var bufOut, bufErr strings.Builder

	go func() {
		for {
			s := <-sig
			if s != syscall.SIGCHLD {
				interrupted.Store(true)
				_ = cmd.Process.Signal(s)
			}
		}
	}()

	go func() {
		out := []io.Writer{&bufOut}

		if !silent {
			out = append(out, os.Stdout)
		}

		w := io.MultiWriter(out...)
		_, _ = io.Copy(w, outReader)
		wg.Done()
	}()

	go func() {
		out := []io.Writer{&bufErr}

		if !silent {
			out = append(out, os.Stderr)
		}

		w := io.MultiWriter(out...)
		_, _ = io.Copy(w, errReader)
		wg.Done()
	}()

	err = cmd.Start()

	if err != nil {
		return bufOut.String(), bufErr.String(), interrupted.Load(), err
	}

	err = cmd.Wait()

	if err != nil {
		return bufOut.String(), bufErr.String(), interrupted.Load(), err
	}

	wg.Wait()

	return bufOut.String(), bufErr.String(), interrupted.Load(), nil
}
