package lifecycle

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
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

	m := New(time.Hour)
	called := 0
	m.stopMachineFn = func() error {
		called++
		return errors.New("stop failed")
	}

	m.onIdleTimeout()
	if called != 1 {
		t.Fatalf("stopMachineFn calls = %d, want 1", called)
	}
	if m.idleTimer == nil {
		t.Fatal("expected idle timer to be rescheduled")
	}
	m.idleTimer.Stop()
}
