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
	running     atomic.Bool
	idleTimeout time.Duration
	timerMu     sync.Mutex
	idleTimer   *time.Timer
}

// New creates a lifecycle Manager with the specified idle timeout.
func New(idleTimeout time.Duration) *Manager {
	return &Manager{idleTimeout: idleTimeout}
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
	m.StopMachine()
}

// StopMachine stops this Fly machine via the Machines API Unix socket.
// Falls back to os.Exit(0) when not running on Fly.io.
func (m *Manager) StopMachine() {
	appName := os.Getenv("FLY_APP_NAME")
	machineID := os.Getenv("FLY_MACHINE_ID")
	if appName == "" || machineID == "" {
		slog.Info("not on Fly.io, exiting normally")
		os.Exit(0)
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", flySocket)
			},
		},
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("http://flaps/v1/apps/%s/machines/%s/stop", appName, machineID)
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		slog.Error("failed to stop machine, falling back to exit", "error", err)
		os.Exit(0)
		return
	}
	defer resp.Body.Close()
	slog.Info("machine stop requested", "status", resp.StatusCode)
}
