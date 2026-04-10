package lifecycle

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeHTTPDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (f fakeHTTPDoer) Do(req *http.Request) (*http.Response, error) {
	return f.do(req)
}

func TestRequestFlyMachineStop_SucceedsOn2xx(t *testing.T) {
	t.Parallel()

	client := fakeHTTPDoer{do: func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("method = %s, want %s", req.Method, http.MethodPost)
		}
		if req.URL.Path != "/v1/apps/test-app/machines/test-machine/stop" {
			t.Fatalf("path = %s", req.URL.Path)
		}

		return &http.Response{
			StatusCode: http.StatusAccepted,
			Status:     "202 Accepted",
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}}

	status, err := requestFlyMachineStop(context.Background(), client, "test-app", "test-machine")
	if err != nil {
		t.Fatalf("requestFlyMachineStop() error = %v", err)
	}
	if status != "202 Accepted" {
		t.Fatalf("status = %q, want %q", status, "202 Accepted")
	}
}

func TestRequestFlyMachineStop_ReturnsErrorOnNon2xx(t *testing.T) {
	t.Parallel()

	client := fakeHTTPDoer{do: func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Status:     "502 Bad Gateway",
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}}

	_, err := requestFlyMachineStop(context.Background(), client, "test-app", "test-machine")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestManagerOnIdleTimeout_ReschedulesAfterStopFailure(t *testing.T) {
	t.Parallel()

	m, err := New(time.Hour)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	called := 0
	m.stopMachineFn = func() error {
		called++
		return errors.New("stop failed")
	}
	m.timerVersion = 1
	m.idleTimer = time.NewTimer(time.Hour)
	defer m.idleTimer.Stop()

	m.onIdleTimeout(1)
	if called != 1 {
		t.Fatalf("stopMachineFn calls = %d, want 1", called)
	}
	if m.idleTimer == nil {
		t.Fatal("expected idle timer to be rescheduled")
	}
	m.idleTimer.Stop()
}

func TestRequestFlyMachineStop_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := fakeHTTPDoer{do: func(req *http.Request) (*http.Response, error) {
		if err := req.Context().Err(); err == nil {
			t.Fatal("expected canceled request context")
		}
		return nil, req.Context().Err()
	}}

	_, err := requestFlyMachineStop(ctx, client, "test-app", "test-machine")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestManagerResetIdleTimer_ConcurrentCallsDoNotDuplicateTimeouts(t *testing.T) {
	t.Parallel()

	m, err := New(25 * time.Millisecond)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var calls atomic.Int32
	fired := make(chan struct{}, 1)
	m.stopMachineFn = func() error {
		if calls.Add(1) == 1 {
			fired <- struct{}{}
		}
		return nil
	}

	const resets = 16
	var wg sync.WaitGroup
	for range resets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.ResetIdleTimer()
		}()
	}
	wg.Wait()

	select {
	case <-fired:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected idle timeout to fire")
	}

	time.Sleep(50 * time.Millisecond)

	if got := calls.Load(); got != 1 {
		t.Fatalf("stopMachineFn calls = %d, want 1", got)
	}
	if got := m.timerVersion; got != resets {
		t.Fatalf("timerVersion = %d, want %d", got, resets)
	}
	if m.idleTimer != nil {
		m.idleTimer.Stop()
	}
}

func TestManagerStop_IsIdempotent(t *testing.T) {
	t.Parallel()

	m, err := New(time.Hour)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	m.SetRunning(true)
	m.ResetIdleTimer()
	firstVersion := m.timerVersion

	m.Stop()
	m.Stop()

	if m.IsRunning() {
		t.Fatal("expected manager to be stopped")
	}
	if m.idleTimer != nil {
		t.Fatal("expected idle timer to be cleared")
	}
	if m.timerVersion != firstVersion+2 {
		t.Fatalf("timerVersion = %d, want %d", m.timerVersion, firstVersion+2)
	}
}

func TestManagerResetIdleTimer_AfterStopStartsFreshTimer(t *testing.T) {
	t.Parallel()

	m, err := New(20 * time.Millisecond)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	fired := make(chan struct{}, 1)
	m.stopMachineFn = func() error {
		fired <- struct{}{}
		return nil
	}

	m.ResetIdleTimer()
	m.Stop()
	m.ResetIdleTimer()

	select {
	case <-fired:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected idle timeout to fire after reset")
	}

	if m.idleTimer != nil {
		m.idleTimer.Stop()
	}
}
