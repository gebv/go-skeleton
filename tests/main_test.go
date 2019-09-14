package tests

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

var (
	Ctx context.Context
)

func TestMain(m *testing.M) {
	log.SetPrefix("testmain: ")
	log.SetFlags(0)

	flag.Parse()

	if testing.Short() {
		log.Print("-short flag is passed, skipping integration tests.")
		os.Exit(0)
	}

	var cancel context.CancelFunc
	Ctx, cancel = context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, unix.SIGTERM, unix.SIGINT)
	go func() {
		s := <-signals
		log.Printf("Got %s, shutting down...", unix.SignalName(s.(unix.Signal)))
		cancel()

		s = <-signals
		log.Panicf("Got %s, exiting!", unix.SignalName(s.(unix.Signal)))
	}()

	var exitCode int
	defer func() {
		if p := recover(); p != nil {
			panic(p)
		}
		os.Exit(exitCode)
	}()

	// TODO: start infra

	// sleep after start infra if need
	time.Sleep(time.Millisecond * 250)

	// TODO: start clients

	exitCode = m.Run()
	cancel()
}

func runMake(arg string) error {
	args := []string{"-C", "..", arg}
	cmd := exec.Command("make", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func runBin(ctx context.Context, bin string, args ...string) error {
	cmd := exec.Command(filepath.Join("..", "bin", bin), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), `GORACE="halt_on_error=1"`)
	log.Print(strings.Join(cmd.Args, " "))
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed bin start (does not wait for it to complete)")
	}

	go func() {
		<-ctx.Done()
		_ = cmd.Process.Signal(unix.SIGTERM)
	}()

	if err := cmd.Wait(); err != nil {
		return errors.Wrap(err, "failed waits for the command to exit")
	}

	return nil
}
