// Package lifecycle manages Fly.io machine lifecycle: idle timeout,
// running state, and machine shutdown via the Fly Machines API.
package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// ErrNotOnFly is returned by StopMachine when the process is not running on Fly.io
// (i.e. FLY_APP_NAME or FLY_MACHINE_ID environment variables are not set).
var ErrNotOnFly = errors.New("not running on Fly.io")

const flySocket = "/.fly/api"

// Manager handles Fly.io machine lifecycle: idle timeout, running state, and shutdown.
type Manager struct {
	stateMu       sync.Mutex
	running       bool
	idleTimeout   time.Duration
	idleTimer     *time.Timer
	timerVersion  uint64
	stopMachineFn func() error
}

// New creates a lifecycle Manager with the specified idle timeout.
func New(idleTimeout time.Duration) (*Manager, error) {
	if idleTimeout <= 0 {
		return nil, fmt.Errorf("idle timeout must be positive: %s", idleTimeout)
	}

	m := &Manager{idleTimeout: idleTimeout}
	m.stopMachineFn = m.StopMachine
	return m, nil
}

// IsRunning returns whether a task is currently in progress.
func (m *Manager) IsRunning() bool {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()

	return m.running
}

// SetRunning sets the running state.
func (m *Manager) SetRunning(v bool) {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()

	m.running = v
}

// ResetIdleTimer (re)starts the idle timeout. Call from the /run handler
// and at server startup.
func (m *Manager) ResetIdleTimer() {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()

	m.timerVersion++
	version := m.timerVersion
	if m.idleTimer != nil {
		m.idleTimer.Stop()
	}
	m.idleTimer = time.AfterFunc(m.idleTimeout, func() {
		m.onIdleTimeout(version)
	})
}

func (m *Manager) onIdleTimeout(version uint64) {
	m.stateMu.Lock()
	if version != m.timerVersion || m.idleTimer == nil {
		m.stateMu.Unlock()
		return
	}

	running := m.running
	if !running {
		m.idleTimer = nil
	}
	m.stateMu.Unlock()

	if running {
		m.ResetIdleTimer()
		return
	}
	slog.Info("idle timeout, stopping machine", "timeout", m.idleTimeout)
	stopMachine := m.stopMachineFn
	if stopMachine == nil {
		stopMachine = m.StopMachine
	}
	if err := stopMachine(); err != nil {
		if errors.Is(err, ErrNotOnFly) {
			slog.Info("not on Fly.io, idle timeout handler returning")
			return
		}
		slog.Error("failed to stop idle machine", "error", err)
		m.ResetIdleTimer()
	}
}

// Stop stops idle timeout handling and marks the manager as not running.
func (m *Manager) Stop() {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()

	m.running = false
	m.timerVersion++
	if m.idleTimer != nil {
		m.idleTimer.Stop()
		m.idleTimer = nil
	}
}

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// StopMachine stops this Fly machine via the Machines API Unix socket.
// Returns ErrNotOnFly when not running on Fly.io (FLY_APP_NAME or FLY_MACHINE_ID unset).
func (m *Manager) StopMachine() error {
	appName := os.Getenv("FLY_APP_NAME")
	machineID := os.Getenv("FLY_MACHINE_ID")
	if appName == "" || machineID == "" {
		slog.Info("not on Fly.io, skipping machine stop")
		return ErrNotOnFly
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
