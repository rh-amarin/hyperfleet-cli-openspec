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

func TestRunWatch_ClearsScreenBeforeEachCall(t *testing.T) {
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
	if !bytes.Contains(buf.Bytes(), []byte(ansiClear)) {
		t.Error("expected ANSI clear sequence in output")
	}
}
