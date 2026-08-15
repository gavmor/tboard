package ui

import (
	"testing"

	"tboard/internal/domain"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderBadge(t *testing.T) {
	statuses := []string{"UP", "SLOW", "DOWN", "PENDING"}
	th := Themes[1]
	for _, st := range statuses {
		badge := RenderBadge(st, th)
		if badge == "" {
			t.Errorf("expected non-empty badge for status %s", st)
		}
	}
}

func TestRender3DStackView(t *testing.T) {
	targets := []domain.Target{
		domain.NewTarget("1.1.1.1", 53, 50, 10, 0),
		domain.NewTarget("8.8.8.8", 53, 50, 10, 1),
	}
	th := Themes[1]
	stackView := Render3DStackView(targets, 100, 30, 3, 2, 0.12, FillSolid, th)
	if stackView == "" {
		t.Error("expected non-empty 3D Stack View output")
	}
}

func TestModel3DStackControls(t *testing.T) {
	m := NewModel(nil, 1)
	m.width = 100
	m.height = 30

	if m.viewMode != View3DStack {
		t.Errorf("expected initial viewMode View3DStack, got %v", m.viewMode)
	}

	// Test 3D depth controls ('w')
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updated.(Model)
	if m.depthY != 3 {
		t.Errorf("expected depthY 3 after pressing 'w', got %d", m.depthY)
	}

	// Test 3D mesh fill style toggle ('m')
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = updated.(Model)
	if m.fillStyle != FillMedium {
		t.Errorf("expected fillStyle FillMedium after pressing 'm', got %v", m.fillStyle)
	}

	// Test View Mode toggle key ('v')
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = updated.(Model)
	if m.viewMode != ViewSplit {
		t.Errorf("expected viewMode ViewSplit after pressing 'v', got %v", m.viewMode)
	}
}
