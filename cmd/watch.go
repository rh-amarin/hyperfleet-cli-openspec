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

const (
	ansiAltEnter   = "\033[?1049h" // switch to alternate screen buffer
	ansiAltExit    = "\033[?1049l" // restore primary screen buffer
	ansiHome       = "\033[H"      // move cursor to top-left
	ansiClearEnd   = "\033[J"      // clear from cursor to end of screen
	ansiClear      = ansiHome + ansiClearEnd // kept for test compatibility
)

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

// render clears from the top and calls fn, then erases any leftover lines from a
// previous render that was taller than the current one.
func render(out io.Writer, fn func() error) error {
	fmt.Fprint(out, ansiHome)
	if err := fn(); err != nil {
		return err
	}
	fmt.Fprint(out, ansiClearEnd)
	return nil
}

// runWatch enters the alternate screen buffer, then calls fn(0) immediately and
// fn(tick) on every s-second tick. Stops when ctx is cancelled or fn returns an error.
// A cancellation due to context is treated as a clean exit (returns nil).
func runWatch(ctx context.Context, out io.Writer, s int, fn func(tick int) error) error {
	fmt.Fprint(out, ansiAltEnter)
	defer fmt.Fprint(out, ansiAltExit)

	ticker := time.NewTicker(time.Duration(s) * time.Second)
	defer ticker.Stop()

	for tick := 0; ; tick++ {
		if err := render(out, func() error { return fn(tick) }); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// runWatchFast enters the alternate screen buffer, then decouples the spinner
// refresh rate (500 ms) from the data-fetch interval (every s seconds).
// fn receives refresh=true on data ticks and refresh=false on spinner-only ticks,
// allowing the caller to re-render from cached data without an API call.
func runWatchFast(ctx context.Context, out io.Writer, s int, fn func(tick int, refresh bool) error) error {
	fmt.Fprint(out, ansiAltEnter)
	defer fmt.Fprint(out, ansiAltExit)

	const spinnerInterval = 500 * time.Millisecond
	spinnerTicker := time.NewTicker(spinnerInterval)
	dataTicker := time.NewTicker(time.Duration(s) * time.Second)
	defer spinnerTicker.Stop()
	defer dataTicker.Stop()

	tick := 0
	if err := render(out, func() error { return fn(tick, true) }); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-dataTicker.C:
			tick++
			if err := render(out, func() error { return fn(tick, true) }); err != nil {
				return err
			}
		case <-spinnerTicker.C:
			tick++
			if err := render(out, func() error { return fn(tick, false) }); err != nil {
				return err
			}
		}
	}
}
