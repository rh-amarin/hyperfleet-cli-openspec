package output

import (
	"testing"
	"time"
)

func TestSpinnerFrame(t *testing.T) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	for i, want := range frames {
		if got := SpinnerFrame(i); got != want {
			t.Errorf("SpinnerFrame(%d) = %q, want %q", i, got, want)
		}
	}
	// wraps at 10
	if SpinnerFrame(10) != SpinnerFrame(0) {
		t.Errorf("SpinnerFrame(10) should equal SpinnerFrame(0)")
	}
	if SpinnerFrame(21) != SpinnerFrame(1) {
		t.Errorf("SpinnerFrame(21) should equal SpinnerFrame(1)")
	}
}

func TestIsActive(t *testing.T) {
	freq := 5

	t.Run("recent timestamp is active", func(t *testing.T) {
		recent := time.Now().Add(-3 * time.Second).Format(time.RFC3339)
		if !IsActive(recent, freq) {
			t.Error("expected active for timestamp 3s ago with freq=5")
		}
	})

	t.Run("stale timestamp is not active", func(t *testing.T) {
		stale := time.Now().Add(-30 * time.Second).Format(time.RFC3339)
		if IsActive(stale, freq) {
			t.Error("expected inactive for timestamp 30s ago with freq=5")
		}
	})

	t.Run("exactly at 2x boundary is not active", func(t *testing.T) {
		boundary := time.Now().Add(-time.Duration(2*freq)*time.Second - time.Millisecond).Format(time.RFC3339)
		if IsActive(boundary, freq) {
			t.Error("expected inactive at exactly 2x frequency boundary")
		}
	})

	t.Run("empty string is not active", func(t *testing.T) {
		if IsActive("", freq) {
			t.Error("expected inactive for empty lastReportTime")
		}
	})

	t.Run("malformed string is not active", func(t *testing.T) {
		if IsActive("not-a-timestamp", freq) {
			t.Error("expected inactive for malformed lastReportTime")
		}
	})

	t.Run("zero frequency uses default watch interval", func(t *testing.T) {
		recent := time.Now().Add(-3 * time.Second).Format(time.RFC3339)
		if !IsActive(recent, 0) {
			t.Error("expected active for timestamp 3s ago with default interval")
		}
		stale := time.Now().Add(-30 * time.Second).Format(time.RFC3339)
		if IsActive(stale, 0) {
			t.Error("expected inactive for timestamp 30s ago with default interval")
		}
	})
}

func TestAdapterActivityPrefix(t *testing.T) {
	recent := time.Now().Add(-1 * time.Second).Format(time.RFC3339)
	got := AdapterActivityPrefix(recent, 2, 0)
	want := SpinnerFrame(2) + " "
	if got != want {
		t.Errorf("active prefix = %q, want %q", got, want)
	}
	if AdapterActivityPrefix("", 0, 5) != "  " {
		t.Error("inactive prefix should be two spaces")
	}
}
