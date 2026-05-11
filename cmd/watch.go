package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const ansiClear = "\033[H\033[2J"

// watchContext returns a context that is cancelled when SIGINT or SIGTERM is received.
func watchContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(ch)
	}()
	return ctx, cancel
}

// runWatch calls fn(0) immediately, then calls fn(tick) on every s-second tick.
// The screen is cleared before each call. Stops when ctx is cancelled or fn returns an error.
// A cancellation due to context is treated as a clean exit (returns nil).
func runWatch(ctx context.Context, out io.Writer, s int, fn func(tick int) error) error {
	ticker := time.NewTicker(time.Duration(s) * time.Second)
	defer ticker.Stop()

	for tick := 0; ; tick++ {
		fmt.Fprint(out, ansiClear)
		if err := fn(tick); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
