package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gdamore/tcell/v2"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "fishtank: %v\n", err)
		os.Exit(1)
	}
}

func start() error {
	population, err := parseOptions(os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("create terminal screen: %w", err)
	}
	return run(ctx, screen, population)
}

func parseOptions(args []string, output io.Writer) (population, error) {
	var counts population
	flags := flag.NewFlagSet("fish-tank-term", flag.ContinueOnError)
	flags.SetOutput(output)
	addCount := func(name, usage string, target **int) {
		flags.Func(name, usage, func(value string) error {
			count, err := strconv.Atoi(value)
			if err != nil || count < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			*target = &count
			return nil
		})
	}
	addCount("fish", "exact fish count (default: automatic)", &counts.fishCount)
	addCount("plants", "exact plant count (default: automatic)", &counts.plantCount)
	if err := flags.Parse(args); err != nil {
		return population{}, err
	}
	if flags.NArg() != 0 {
		return population{}, fmt.Errorf("unexpected argument: %q", flags.Arg(0))
	}
	return counts, nil
}
