// Package lifecycle manages Fly.io machine lifecycle: idle timeout,
// running state, and machine shutdown via the Fly Machines API.
package lifecycle

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const flySocket = "/.fly/api"

// Manager handles Fly.io machine lifecycle: idle timeout, running state, and shutdown.
type Manager struct {
	running       atomic.Bool
	idleTimeout   time.Duration
	timerMu       sync.Mutex
	idleTimer     *time.Timer
	stopMachineFn func() error
}

// New creates a lifecycle Manager with the specified idle timeout.
func New(idleTimeout time.Duration) *Manager {
	m := &Manager{idleTimeout: idleTimeout}
	m.stopMachineFn = m.StopMachine
	return m
}

// IsRunning returns whether a task is currently in progress.
func (m *Manager) IsRunning() bool {
	return m.running.Load()
}

// SetRunning sets the running state.
func (m *Manager) SetRunning(v bool) {
	m.running.Store(v)
}

// ResetIdleTimer (re)starts the idle timeout. Call from the /run handler
// and at server startup.
func (m *Manager) ResetIdleTimer() {
	m.timerMu.Lock()
	defer m.timerMu.Unlock()
	if m.idleTimer != nil {
		m.idleTimer.Stop()
	}
	m.idleTimer = time.AfterFunc(m.idleTimeout, m.onIdleTimeout)
}

func (m *Manager) onIdleTimeout() {
	if m.IsRunning() {
		m.ResetIdleTimer()
		return
	}
	slog.Info("idle timeout, stopping machine", "timeout", m.idleTimeout)
	stopMachine := m.stopMachineFn
	if stopMachine == nil {
		stopMachine = m.StopMachine
	}
	if err := stopMachine(); err != nil {
		slog.Error("failed to stop idle machine", "error", err)
		m.ResetIdleTimer()
	}
}

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// StopMachine stops this Fly machine via the Machines API Unix socket.
// Falls back to os.Exit(0) when not running on Fly.io.
func (m *Manager) StopMachine() error {
	appName := os.Getenv("FLY_APP_NAME")
	machineID := os.Getenv("FLY_MACHINE_ID")
	if appName == "" || machineID == "" {
		slog.Info("not on Fly.io, exiting normally")
		os.Exit(0)
		return nil
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var dialer net.Dialer
				return dialer.DialContext(ctx, "unix", flySocket)
			},
		},
		Timeout: 10 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status, err := requestFlyMachineStop(ctx, client, appName, machineID)
	if err != nil {
		return fmt.Errorf("request machine stop: %w", err)
	}

	slog.Info("machine stop requested", "status", status)
	return nil
}

func requestFlyMachineStop(ctx context.Context, client httpDoer, appName, machineID string) (string, error) {
	url := fmt.Sprintf("http://flaps/v1/apps/%s/machines/%s/stop", appName, machineID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", fmt.Errorf("build stop request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send stop request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("unexpected stop status: %s", resp.Status)
	}

	return resp.Status, nil
}
