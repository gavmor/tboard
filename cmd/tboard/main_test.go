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
	target := newTarget("1.1.1.1", 53, 50, 10, 0)
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
	target := newTarget("google.com", 443, 50, 10, 0)

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
	th := themes[1]
	for _, st := range statuses {
		badge := renderBadge(st, th)
		if badge == "" {
			t.Errorf("expected non-empty badge for status %s", st)
		}
	}
}

func TestRender3DStackView(t *testing.T) {
	m := initialModel(nil, 1)
	m.width = 100
	m.height = 30
	stackView := m.render3DStackView()
	if stackView == "" {
		t.Error("expected non-empty 3D Stack View output")
	}
}

func TestModel3DStackControls(t *testing.T) {
	m := initialModel(nil, 1)
	m.width = 100
	m.height = 30

	if m.viewMode != View3DStack {
		t.Errorf("expected initial viewMode View3DStack, got %v", m.viewMode)
	}

	// Test 3D depth controls ('w')
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updated.(model)
	if m.depthY != 3 {
		t.Errorf("expected depthY 3 after pressing 'w', got %d", m.depthY)
	}

	// Test 3D mesh fill style toggle ('m')
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = updated.(model)
	if m.fillStyle != FillMedium {
		t.Errorf("expected fillStyle FillMedium after pressing 'm', got %v", m.fillStyle)
	}

	// Test View Mode toggle key ('v')
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = updated.(model)
	if m.viewMode != ViewSplit {
		t.Errorf("expected viewMode ViewSplit after pressing 'v', got %v", m.viewMode)
	}
}
