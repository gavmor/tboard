package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines keybindings using bubbles/key
type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Add        key.Binding
	Delete     key.Binding
	ToggleView key.Binding
	CamUp      key.Binding
	CamDown    key.Binding
	CamLeft    key.Binding
	CamRight   key.Binding
	ScaleUp    key.Binding
	ScaleDown  key.Binding
	ToggleMesh key.Binding
	Theme      key.Binding
	Pause      key.Binding
	Reset      key.Binding
	Help       key.Binding
	Quit       key.Binding
}

// NewKeyMap creates and initializes all TUI keybindings.
func NewKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "down"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add host"),
		),
		Delete: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete host"),
		),
		ToggleView: key.NewBinding(
			key.WithKeys("tab", "v"),
			key.WithHelp("tab/v", "cycle view mode"),
		),
		CamUp: key.NewBinding(
			key.WithKeys("w", "up"),
			key.WithHelp("w/↑", "3D Z-depth +"),
		),
		CamDown: key.NewBinding(
			key.WithKeys("s", "down"),
			key.WithHelp("s/↓", "3D Z-depth -"),
		),
		CamLeft: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "3D X-slant -"),
		),
		CamRight: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "3D X-slant +"),
		),
		ScaleUp: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+", "3D height +"),
		),
		ScaleDown: key.NewBinding(
			key.WithKeys("-", "_"),
			key.WithHelp("-", "3D height -"),
		),
		ToggleMesh: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "3D fill style"),
		),
		Theme: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "theme"),
		),
		Pause: key.NewBinding(
			key.WithKeys("space", " "),
			key.WithHelp("space", "pause"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ToggleView, k.CamUp, k.CamDown, k.ScaleUp, k.ToggleMesh, k.Theme, k.Help, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ToggleView, k.Add, k.Delete, k.Theme, k.Pause},
		{k.CamUp, k.CamDown, k.CamLeft, k.CamRight},
		{k.ScaleUp, k.ScaleDown, k.ToggleMesh, k.Reset, k.Quit},
	}
}
