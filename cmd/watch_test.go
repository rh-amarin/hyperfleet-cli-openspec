package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunWatch_CallsFnImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls []int
	done := make(chan error, 1)
	go func() {
		done <- runWatch(ctx, &bytes.Buffer{}, 60, func(tick int) error {
			calls = append(calls, tick)
			cancel() // stop after first call
			return nil
		})
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runWatch did not return after context cancel")
	}

	if len(calls) != 1 || calls[0] != 0 {
		t.Errorf("expected fn called once with tick=0, got %v", calls)
	}
}

func TestRunWatch_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	done := make(chan error, 1)
	go func() {
		done <- runWatch(ctx, &bytes.Buffer{}, 60, func(tick int) error {
			callCount++
			if callCount == 1 {
				cancel()
			}
			return nil
		})
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil error on cancel, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runWatch did not return after context cancel")
	}
}

func TestRunWatch_StopsOnFnError(t *testing.T) {
	ctx := context.Background()
	sentinel := errors.New("stop")

	done := make(chan error, 1)
	go func() {
		done <- runWatch(ctx, &bytes.Buffer{}, 60, func(tick int) error {
			return sentinel
		})
	}()

	select {
	case err := <-done:
		if !errors.Is(err, sentinel) {
			t.Errorf("expected sentinel error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runWatch did not return after fn error")
	}
}

func TestRunWatch_UsesAltScreenAndClearsOnRender(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var buf bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- runWatch(ctx, &buf, 60, func(tick int) error {
			cancel()
			return nil
		})
	}()

	<-done
	got := buf.Bytes()
	if !bytes.Contains(got, []byte(ansiAltEnter)) {
		t.Error("expected alternate-screen enter sequence")
	}
	if !bytes.Contains(got, []byte(ansiAltExit)) {
		t.Error("expected alternate-screen exit sequence")
	}
	if !bytes.Contains(got, []byte(ansiHome)) {
		t.Error("expected cursor-home sequence before each render")
	}
}

func TestRunWatchFast_CallsFnImmediatelyWithRefresh(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type call struct {
		tick    int
		refresh bool
	}
	var calls []call
	done := make(chan error, 1)
	go func() {
		done <- runWatchFast(ctx, &bytes.Buffer{}, 60, func(tick int, refresh bool) error {
			calls = append(calls, call{tick, refresh})
			cancel()
			return nil
		})
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runWatchFast did not return after context cancel")
	}

	if len(calls) != 1 || calls[0].tick != 0 || !calls[0].refresh {
		t.Errorf("expected fn called once with tick=0 refresh=true, got %v", calls)
	}
}

func TestRunWatchFast_SpinnerTicksHaveRefreshFalse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type call struct {
		tick    int
		refresh bool
	}
	var calls []call
	done := make(chan error, 1)
	go func() {
		// data interval is 60s so only spinner ticks fire during the test
		done <- runWatchFast(ctx, &bytes.Buffer{}, 60, func(tick int, refresh bool) error {
			calls = append(calls, call{tick, refresh})
			if len(calls) >= 2 {
				cancel()
			}
			return nil
		})
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("runWatchFast did not fire spinner ticks within 3s")
	}

	if len(calls) < 2 {
		t.Fatalf("expected at least 2 calls, got %d", len(calls))
	}
	// first call is the initial data fetch
	if !calls[0].refresh {
		t.Errorf("first call should have refresh=true, got %v", calls[0])
	}
	// subsequent calls within the long data interval are spinner-only
	if calls[1].refresh {
		t.Errorf("second call should have refresh=false (spinner tick), got %v", calls[1])
	}
}

func TestRunWatchFast_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- runWatchFast(ctx, &bytes.Buffer{}, 60, func(tick int, refresh bool) error {
			cancel()
			return nil
		})
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil error on cancel, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runWatchFast did not return after context cancel")
	}
}

func TestRunWatchFast_StopsOnFnError(t *testing.T) {
	ctx := context.Background()
	sentinel := errors.New("stop")

	done := make(chan error, 1)
	go func() {
		done <- runWatchFast(ctx, &bytes.Buffer{}, 60, func(tick int, refresh bool) error {
			return sentinel
		})
	}()

	select {
	case err := <-done:
		if !errors.Is(err, sentinel) {
			t.Errorf("expected sentinel error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runWatchFast did not return after fn error")
	}
}

func TestRunWatchFast_UsesAltScreenAndClearsOnRender(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var buf bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- runWatchFast(ctx, &buf, 60, func(tick int, refresh bool) error {
			cancel()
			return nil
		})
	}()

	<-done
	got := buf.Bytes()
	if !bytes.Contains(got, []byte(ansiAltEnter)) {
		t.Error("expected alternate-screen enter sequence")
	}
	if !bytes.Contains(got, []byte(ansiAltExit)) {
		t.Error("expected alternate-screen exit sequence")
	}
	if !bytes.Contains(got, []byte(ansiHome)) {
		t.Error("expected cursor-home sequence before each render")
	}
}
