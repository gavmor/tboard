package probe

import (
	"net"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// PingResultMsg carries latency result or error from a probe.
type PingResultMsg struct {
	TargetIdx int
	Rtt       time.Duration
	Err       error
}

// TickMsg signals timer tick for polling loop.
type TickMsg time.Time

// ParseHostPort extracts host and port from raw input string, applying defaultPort if omitted.
func ParseHostPort(raw string, defaultPort int) (string, int) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", defaultPort
	}

	if strings.Contains(raw, ":") {
		host, portStr, err := net.SplitHostPort(raw)
		if err == nil {
			if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
				return host, p
			}
		}
		parts := strings.Split(raw, ":")
		if len(parts) == 2 {
			if p, err := strconv.Atoi(parts[1]); err == nil && p > 0 {
				return parts[0], p
			}
		}
	}

	return raw, defaultPort
}

// PingHostCmd returns a tea.Cmd that probes TCP connection latency.
func PingHostCmd(targetIdx int, host string, port int) tea.Cmd {
	return func() tea.Msg {
		address := net.JoinHostPort(host, strconv.Itoa(port))
		start := time.Now()
		conn, err := net.DialTimeout("tcp", address, 1500*time.Millisecond)
		rtt := time.Since(start)
		if err != nil {
			return PingResultMsg{TargetIdx: targetIdx, Rtt: 0, Err: err}
		}
		conn.Close()
		return PingResultMsg{TargetIdx: targetIdx, Rtt: rtt, Err: nil}
	}
}

// TickCmd returns a tea.Cmd timer tick.
func TickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
