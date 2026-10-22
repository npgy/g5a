package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("retry", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	retries := flags.Int("n", 0, "number of retries after the initial attempt")
	wait := flags.Duration("wait", 0, "delay between retries")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	if *retries < 0 || *wait < 0 || flags.NArg() != 1 {
		return fmt.Errorf("usage: go run ./cmd/retry -n <retries> -wait <duration> -- <shell-command>")
	}

	var err error
	for attempt := 0; attempt <= *retries; attempt++ {
		if err = executeShellCommand(flags.Arg(0)); err == nil {
			return nil
		}
		if attempt < *retries {
			time.Sleep(*wait)
		}
	}

	return fmt.Errorf("command failed after %d attempt(s): %w", *retries+1, err)
}

func executeShellCommand(command string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
