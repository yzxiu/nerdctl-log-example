package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/fahedouch/go-logrotate"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

func main() {

	fmt.Println("log drive start!!!")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		sigCh = make(chan os.Signal, 32)
		errCh = make(chan error, 1)
	)
	signal.Notify(sigCh, unix.SIGTERM)

	// cmd.ExtraFiles = append(cmd.ExtraFiles, out.r, serr.r, w)
	out := os.NewFile(3, "CONTAINER_STDOUT")
	serr := os.NewFile(4, "CONTAINER_STDERR")
	wait := os.NewFile(5, "CONTAINER_WAIT")

	go func() {
		errCh <- logger(ctx, out, serr, wait.Close)
	}()

	for {
		select {
		case <-sigCh:
			cancel()
		case err := <-errCh:
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Println("log drive exit 0")
			os.Exit(0)
		}
	}

}

func logger(_ context.Context, out *os.File, serr *os.File, ready func() error) error {

	// Notify the shim that it is ready
	// call wait.Close
	// r will receive io.EOF error
	if err := ready(); err != nil {
		return err
	}

	// log path
	jsonFilePath := "app.log"
	l := &logrotate.Logger{
		Filename: jsonFilePath,
	}
	return Encode(l, out, serr)
}

// Entry is compatible with Docker "json-file" logs
type Entry struct {
	Log    string    `json:"log,omitempty"`    // line, including "\r\n"
	Stream string    `json:"stream,omitempty"` // "stdout" or "stderr"
	Time   time.Time `json:"time"`             // e.g. "2020-12-11T20:29:41.939902251Z"
}

func Encode(w io.WriteCloser, stdout, stderr io.Reader) error {
	enc := json.NewEncoder(w)
	var encMu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)
	f := func(r io.Reader, name string) {
		defer wg.Done()
		br := bufio.NewReader(r)
		e := &Entry{
			Stream: name,
		}
		for {
			line, err := br.ReadString(byte('\n'))
			if err != nil {
				logrus.WithError(err).Errorf("failed to read line from %q", name)
				return
			}
			e.Log = line
			e.Time = time.Now().UTC()
			encMu.Lock()
			encErr := enc.Encode(e)
			encMu.Unlock()
			if encErr != nil {
				logrus.WithError(err).Errorf("failed to encode JSON")
				return
			}
		}
	}
	go f(stdout, "stdout")
	go f(stderr, "stderr")
	wg.Wait()
	return nil
}
