package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {

	// start log driver
	pio, err := driveIO()
	if err != nil {
		log.Fatal(err)
	}

	// start app
	cmd := exec.Command("./app-example")
	cmd.Stdout = pio.out.w
	cmd.Stderr = pio.err.w

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func newPipe() (*pipe, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	return &pipe{
		r: r,
		w: w,
	}, nil
}

type pipe struct {
	r *os.File
	w *os.File
}

type binaryIO struct {
	cmd      *exec.Cmd
	out, err *pipe
}

func (p *pipe) Close() error {
	if err := p.w.Close(); err != nil {
	}
	if err := p.r.Close(); err != nil {
	}
	return fmt.Errorf("pipe close error")
}

func driveIO() (_ *binaryIO, err error) {

	var closers []func() error

	// app out pipe
	out, err := newPipe()
	if err != nil {
		return nil, err
	}
	closers = append(closers, out.Close)

	// app err pipe
	serr, err := newPipe()
	if err != nil {
		return nil, err
	}
	closers = append(closers, serr.Close)

	// drive ready pipe
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	closers = append(closers, r.Close, w.Close)

	cmd := exec.Command("./drive-example")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.ExtraFiles = append(cmd.ExtraFiles, out.r, serr.r, w)

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	closers = append(closers, func() error { return cmd.Process.Kill() })

	// close our side of the pipe after start
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close write pipe after start: %w", err)
	}

	// wait for the logging binary to be ready
	b := make([]byte, 1)
	if _, err := r.Read(b); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read from logging binary: %w", err)
	}

	return &binaryIO{
		cmd: cmd,
		out: out,
		err: serr,
	}, nil
}
