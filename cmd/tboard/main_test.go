package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		input        string
		defaultPort  int
		expectedHost string
		expectedPort int
	}{
		{"google.com", 80, "google.com", 80},
		{"1.1.1.1:53", 80, "1.1.1.1", 53},
		{"github.com:443", 80, "github.com", 443},
		{"10.0.0.1", 8080, "10.0.0.1", 8080},
	}

	for _, tt := range tests {
		h, p := parseHostPort(tt.input, tt.defaultPort)
		if h != tt.expectedHost || p != tt.expectedPort {
			t.Errorf("parseHostPort(%q, %d) = (%q, %d), expected (%q, %d)",
				tt.input, tt.defaultPort, h, p, tt.expectedHost, tt.expectedPort)
		}
	}
}

func TestLoadConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	content := `
theme: dracula
default_port: 443
targets:
  - host: 1.1.1.1:53
  - host: custom.domain.org
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}

	cfg, err := loadConfigFile(configPath)
	if err != nil {
		t.Fatalf("expected no error loading config file, got %v", err)
	}

	if cfg.Theme != "dracula" {
		t.Errorf("expected theme 'dracula', got %s", cfg.Theme)
	}
	if cfg.DefaultPort != 443 {
		t.Errorf("expected default port 443, got %d", cfg.DefaultPort)
	}
}

func TestTargetLossPercentage(t *testing.T) {
	target := newTarget("1.1.1.1", 53)
	if target.LossPercentage() != 0.0 {
		t.Errorf("expected 0%% loss for 0 sent, got %f", target.LossPercentage())
	}

	target.AddResult(10*time.Millisecond, nil)
	if target.LossPercentage() != 0.0 {
		t.Errorf("expected 0%% loss, got %f", target.LossPercentage())
	}

	target.AddResult(0, errors.New("timeout"))
	if target.LossPercentage() != 50.0 {
		t.Errorf("expected 50%% loss, got %f", target.LossPercentage())
	}
}

func TestTargetAddResult(t *testing.T) {
	target := newTarget("google.com", 443)

	target.AddResult(30*time.Millisecond, nil)
	if target.Status != "UP" {
		t.Errorf("expected status UP, got %s", target.Status)
	}
	if target.MinRtt != 30*time.Millisecond {
		t.Errorf("expected MinRtt 30ms, got %v", target.MinRtt)
	}

	target.AddResult(250*time.Millisecond, nil)
	if target.Status != "SLOW" {
		t.Errorf("expected status SLOW, got %s", target.Status)
	}
	if target.MaxRtt != 250*time.Millisecond {
		t.Errorf("expected MaxRtt 250ms, got %v", target.MaxRtt)
	}

	target.AddResult(0, errors.New("connection refused"))
	if target.Status != "DOWN" {
		t.Errorf("expected status DOWN, got %s", target.Status)
	}
	if target.LastError != "connection refused" {
		t.Errorf("expected LastError 'connection refused', got %s", target.LastError)
	}
}

func TestRenderBadge(t *testing.T) {
	statuses := []string{"UP", "SLOW", "DOWN", "PENDING"}
	th := themes[1] // Gruvbox Light
	for _, st := range statuses {
		badge := renderBadge(st, th)
		if badge == "" {
			t.Errorf("expected non-empty badge for status %s", st)
		}
	}
}

func TestRenderSparklineLogarithmicTime(t *testing.T) {
	now := time.Now()
	samples := []Sample{
		{Rtt: 10 * time.Millisecond, Timestamp: now.Add(-60 * time.Second)},
		{Rtt: 50 * time.Millisecond, Timestamp: now.Add(-30 * time.Second)},
		{Rtt: -1, Timestamp: now.Add(-10 * time.Second)},
		{Rtt: 200 * time.Millisecond, Timestamp: now},
	}
	th := themes[1]
	sparkline := renderSparkline(samples, 20, th)
	if sparkline == "" {
		t.Error("expected non-empty sparkline rendering for timestamped samples")
	}
}

func TestRenderTable(t *testing.T) {
	m := initialModel(nil, 1)
	tbl := m.renderTable()
	if tbl == "" {
		t.Error("expected non-empty table output")
	}
}

func TestModelBubblesComponents(t *testing.T) {
	m := initialModel(nil, 1)
	if len(m.targets) != 5 {
		t.Fatalf("expected 5 default targets, got %d", len(m.targets))
	}

	// Test theme key 't'
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = updated.(model)
	if m.themeIdx != 2 {
		t.Errorf("expected themeIdx 2 after pressing 't', got %d", m.themeIdx)
	}

	// Test pause key
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(model)
	if !m.paused {
		t.Error("expected model to be paused")
	}

	// Test help toggle key '?'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(model)
	if !m.help.ShowAll {
		t.Error("expected full help to be visible after pressing '?'")
	}
}
