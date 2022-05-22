package utils

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func RunCommand(cmdWithArgs []string, silent bool) (string, string, error) {
	cmd := exec.Command(cmdWithArgs[0], cmdWithArgs[1:]...)

	outReader, err := cmd.StdoutPipe()

	if err != nil {
		return "", "", err
	}

	errReader, err := cmd.StderrPipe()

	if err != nil {
		return "", "", err
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig)
	signal.Ignore(syscall.SIGURG)
	var bufOut, bufErr strings.Builder

	go func() {
		for {
			s := <-sig
			_ = cmd.Process.Signal(s)
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
		return bufOut.String(), bufErr.String(), err
	}

	err = cmd.Wait()

	if err != nil {
		return bufOut.String(), bufErr.String(), err
	}

	wg.Wait()

	return bufOut.String(), bufErr.String(), nil
}
