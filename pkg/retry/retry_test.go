package retry

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRetryDoDoesNotUseTimeAfterInDelaySelect(t *testing.T) {
	content, err := os.ReadFile("retry.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "time.After(") {
		t.Fatal("retry delay should use time.NewTimer with Stop/drain instead of time.After")
	}
}

func TestDoReturnsContextErrorWhenCancelledDuringDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	err := Do(ctx, Config{MaxAttempts: 2, BaseDelay: time.Hour, MaxDelay: time.Hour, Multiplier: 2}, func() error {
		attempts++
		cancel()
		return errors.New("temporary failure")
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}
