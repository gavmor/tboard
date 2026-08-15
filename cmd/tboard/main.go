package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var version = "0.1.0"

// Config File representation
type ConfigFile struct {
	Theme       string         `json:"theme"`
	DefaultPort int            `json:"default_port"`
	Targets     []ConfigTarget `json:"targets"`
}

type ConfigTarget struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Sample records a latency measurement and its exact timestamp.
type Sample struct {
	Rtt       time.Duration // -1 indicates timeout/error
	Timestamp time.Time
}

// Target represents a host monitored by the ping dashboard.
type Target struct {
	Host      string
	Port      int
	Samples   []Sample // Time-stamped history for variable time-axis rendering
	Sent      int
	Received  int
	LastRtt   time.Duration
	MinRtt    time.Duration
	MaxRtt    time.Duration
	AvgRtt    time.Duration
	Status    string // "UP", "SLOW", "DOWN", "PENDING"
	LastError string
}

func newTarget(host string, port int) Target {
	return Target{
		Host:    host,
		Port:    port,
		Samples: make([]Sample, 0, 180),
		Status:  "PENDING",
	}
}

func parseHostPort(raw string, defaultPort int) (string, int) {
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
		// Fallback split if net.SplitHostPort fails (e.g. no ipv6 brackets)
		parts := strings.Split(raw, ":")
		if len(parts) == 2 {
			if p, err := strconv.Atoi(parts[1]); err == nil && p > 0 {
				return parts[0], p
			}
		}
	}

	return raw, defaultPort
}

// Packet loss percentage
func (t Target) LossPercentage() float64 {
	if t.Sent == 0 {
		return 0.0
	}
	lost := t.Sent - t.Received
	return (float64(lost) / float64(t.Sent)) * 100.0
}

// Add a ping measurement to the host history
func (t *Target) AddResult(rtt time.Duration, err error) {
	t.Sent++
	const maxHistory = 180

	sample := Sample{
		Timestamp: time.Now(),
	}

	if err != nil {
		t.LastError = err.Error()
		sample.Rtt = -1
		t.Status = "DOWN"
	} else {
		t.Received++
		t.LastRtt = rtt
		sample.Rtt = rtt
		t.LastError = ""

		if t.MinRtt == 0 || rtt < t.MinRtt {
			t.MinRtt = rtt
		}
		if rtt > t.MaxRtt {
			t.MaxRtt = rtt
		}

		var total time.Duration
		validCount := 0
		for _, s := range t.Samples {
			if s.Rtt >= 0 {
				total += s.Rtt
				validCount++
			}
		}
		if validCount > 0 {
			t.AvgRtt = total / time.Duration(validCount)
		}

		if rtt > 200*time.Millisecond {
			t.Status = "SLOW"
		} else {
			t.Status = "UP"
		}
	}

	t.Samples = append(t.Samples, sample)
	if len(t.Samples) > maxHistory {
		t.Samples = t.Samples[1:]
	}
}

// KeyMap defines keybindings using bubbles/key
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Add    key.Binding
	Delete key.Binding
	Theme  key.Binding
	Pause  key.Binding
	Reset  key.Binding
	Help   key.Binding
	Quit   key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/↓", "down"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add host"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
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

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Add, k.Delete, k.Theme, k.Pause, k.Reset, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Add, k.Delete},
		{k.Theme, k.Pause, k.Reset},
		{k.Help, k.Quit},
	}
}

// Color Theme definitions
type Theme struct {
	Name           string
	TitleFg        lipgloss.TerminalColor
	TitleBg        lipgloss.TerminalColor
	SubTitle       lipgloss.TerminalColor
	CardBorder     lipgloss.TerminalColor
	SelCardBorder  lipgloss.TerminalColor
	HostText       lipgloss.TerminalColor
	BadgeUpFg      lipgloss.TerminalColor
	BadgeUpBg      lipgloss.TerminalColor
	BadgeSlowFg    lipgloss.TerminalColor
	BadgeSlowBg    lipgloss.TerminalColor
	BadgeDownFg    lipgloss.TerminalColor
	BadgeDownBg    lipgloss.TerminalColor
	BadgePendingFg lipgloss.TerminalColor
	BadgePendingBg lipgloss.TerminalColor
	StatLabel      lipgloss.TerminalColor
	StatVal        lipgloss.TerminalColor
	SparklineEmpty lipgloss.TerminalColor
	SparklineGreen lipgloss.TerminalColor
	SparklineYel   lipgloss.TerminalColor
	SparklineRed   lipgloss.TerminalColor
	DetailTitle    lipgloss.TerminalColor
	DetailBorder   lipgloss.TerminalColor
	InputPrompt    lipgloss.TerminalColor
	StatusMsg      lipgloss.TerminalColor
	LastError      lipgloss.TerminalColor
	FooterKey      lipgloss.TerminalColor
	FooterDesc     lipgloss.TerminalColor
}

var themes = []Theme{
	{
		Name:           "Adaptive (Auto)",
		TitleFg:        lipgloss.AdaptiveColor{Light: "#fbf1c7", Dark: "#FAFAFA"},
		TitleBg:        lipgloss.AdaptiveColor{Light: "#8f3f71", Dark: "#7D56F4"},
		SubTitle:       lipgloss.AdaptiveColor{Light: "#665c54", Dark: "#8A8A8A"},
		CardBorder:     lipgloss.AdaptiveColor{Light: "#d5c4a1", Dark: "#3C3C3C"},
		SelCardBorder:  lipgloss.AdaptiveColor{Light: "#8f3f71", Dark: "#7D56F4"},
		HostText:       lipgloss.AdaptiveColor{Light: "#282828", Dark: "#FFFFFF"},
		BadgeUpFg:      lipgloss.AdaptiveColor{Light: "#fbf1c7", Dark: "#000000"},
		BadgeUpBg:      lipgloss.AdaptiveColor{Light: "#79740e", Dark: "#50FA7B"},
		BadgeSlowFg:    lipgloss.AdaptiveColor{Light: "#282828", Dark: "#000000"},
		BadgeSlowBg:    lipgloss.AdaptiveColor{Light: "#b57614", Dark: "#F1FA8C"},
		BadgeDownFg:    lipgloss.AdaptiveColor{Light: "#fbf1c7", Dark: "#FFFFFF"},
		BadgeDownBg:    lipgloss.AdaptiveColor{Light: "#9d0006", Dark: "#FF5555"},
		BadgePendingFg: lipgloss.AdaptiveColor{Light: "#fbf1c7", Dark: "#000000"},
		BadgePendingBg: lipgloss.AdaptiveColor{Light: "#7c6f64", Dark: "#6272A4"},
		StatLabel:      lipgloss.AdaptiveColor{Light: "#665c54", Dark: "#6272A4"},
		StatVal:        lipgloss.AdaptiveColor{Light: "#282828", Dark: "#F8F8F2"},
		SparklineEmpty: lipgloss.AdaptiveColor{Light: "#d5c4a1", Dark: "#44475A"},
		SparklineGreen: lipgloss.AdaptiveColor{Light: "#79740e", Dark: "#50FA7B"},
		SparklineYel:   lipgloss.AdaptiveColor{Light: "#b57614", Dark: "#F1FA8C"},
		SparklineRed:   lipgloss.AdaptiveColor{Light: "#9d0006", Dark: "#FF5555"},
		DetailTitle:    lipgloss.AdaptiveColor{Light: "#8f3f71", Dark: "#BD93F9"},
		DetailBorder:   lipgloss.AdaptiveColor{Light: "#7c6f64", Dark: "#6272A4"},
		InputPrompt:    lipgloss.AdaptiveColor{Light: "#79740e", Dark: "#50FA7B"},
		StatusMsg:      lipgloss.AdaptiveColor{Light: "#b57614", Dark: "#F1FA8C"},
		LastError:      lipgloss.AdaptiveColor{Light: "#9d0006", Dark: "#FF5555"},
		FooterKey:      lipgloss.AdaptiveColor{Light: "#8f3f71", Dark: "#BD93F9"},
		FooterDesc:     lipgloss.AdaptiveColor{Light: "#665c54", Dark: "#6272A4"},
	},
	{
		Name:           "Gruvbox Light",
		TitleFg:        lipgloss.Color("#fbf1c7"),
		TitleBg:        lipgloss.Color("#8f3f71"),
		SubTitle:       lipgloss.Color("#665c54"),
		CardBorder:     lipgloss.Color("#d5c4a1"),
		SelCardBorder:  lipgloss.Color("#8f3f71"),
		HostText:       lipgloss.Color("#282828"),
		BadgeUpFg:      lipgloss.Color("#fbf1c7"),
		BadgeUpBg:      lipgloss.Color("#79740e"),
		BadgeSlowFg:    lipgloss.Color("#282828"),
		BadgeSlowBg:    lipgloss.Color("#b57614"),
		BadgeDownFg:    lipgloss.Color("#fbf1c7"),
		BadgeDownBg:    lipgloss.Color("#9d0006"),
		BadgePendingFg: lipgloss.Color("#fbf1c7"),
		BadgePendingBg: lipgloss.Color("#7c6f64"),
		StatLabel:      lipgloss.Color("#665c54"),
		StatVal:        lipgloss.Color("#282828"),
		SparklineEmpty: lipgloss.Color("#d5c4a1"),
		SparklineGreen: lipgloss.Color("#79740e"),
		SparklineYel:   lipgloss.Color("#b57614"),
		SparklineRed:   lipgloss.Color("#9d0006"),
		DetailTitle:    lipgloss.Color("#8f3f71"),
		DetailBorder:   lipgloss.Color("#7c6f64"),
		InputPrompt:    lipgloss.Color("#79740e"),
		StatusMsg:      lipgloss.Color("#b57614"),
		LastError:      lipgloss.Color("#9d0006"),
		FooterKey:      lipgloss.Color("#8f3f71"),
		FooterDesc:     lipgloss.Color("#665c54"),
	},
	{
		Name:           "Dracula Dark",
		TitleFg:        lipgloss.Color("#FAFAFA"),
		TitleBg:        lipgloss.Color("#7D56F4"),
		SubTitle:       lipgloss.Color("#8A8A8A"),
		CardBorder:     lipgloss.Color("#3C3C3C"),
		SelCardBorder:  lipgloss.Color("#7D56F4"),
		HostText:       lipgloss.Color("#FFFFFF"),
		BadgeUpFg:      lipgloss.Color("#000000"),
		BadgeUpBg:      lipgloss.Color("#50FA7B"),
		BadgeSlowFg:    lipgloss.Color("#000000"),
		BadgeSlowBg:    lipgloss.Color("#F1FA8C"),
		BadgeDownFg:    lipgloss.Color("#FFFFFF"),
		BadgeDownBg:    lipgloss.Color("#FF5555"),
		BadgePendingFg: lipgloss.Color("#000000"),
		BadgePendingBg: lipgloss.Color("#6272A4"),
		StatLabel:      lipgloss.Color("#6272A4"),
		StatVal:        lipgloss.Color("#F8F8F2"),
		SparklineEmpty: lipgloss.Color("#44475A"),
		SparklineGreen: lipgloss.Color("#50FA7B"),
		SparklineYel:   lipgloss.Color("#F1FA8C"),
		SparklineRed:   lipgloss.Color("#FF5555"),
		DetailTitle:    lipgloss.Color("#BD93F9"),
		DetailBorder:   lipgloss.Color("#6272A4"),
		InputPrompt:    lipgloss.Color("#50FA7B"),
		StatusMsg:      lipgloss.Color("#F1FA8C"),
		LastError:      lipgloss.Color("#FF5555"),
		FooterKey:      lipgloss.Color("#BD93F9"),
		FooterDesc:     lipgloss.Color("#6272A4"),
	},
}

func parseThemeName(name string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "auto", "adaptive":
		return 0
	case "gruvbox", "gruvbox-light", "light":
		return 1
	case "dracula", "dracula-dark", "dark":
		return 2
	default:
		return 1
	}
}

func colorToFGANSI(c lipgloss.TerminalColor) string {
	switch v := c.(type) {
	case lipgloss.Color:
		hex := strings.TrimPrefix(string(v), "#")
		if len(hex) == 6 {
			var r, g, b uint8
			fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
			return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
		}
	case lipgloss.AdaptiveColor:
		hex := v.Light
		if lipgloss.HasDarkBackground() {
			hex = v.Dark
		}
		hex = strings.TrimPrefix(hex, "#")
		if len(hex) == 6 {
			var r, g, b uint8
			fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
			return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
		}
	}
	return ""
}

func colorToBGANSI(c lipgloss.TerminalColor) string {
	switch v := c.(type) {
	case lipgloss.Color:
		hex := strings.TrimPrefix(string(v), "#")
		if len(hex) == 6 {
			var r, g, b uint8
			fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
			return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
		}
	case lipgloss.AdaptiveColor:
		hex := v.Light
		if lipgloss.HasDarkBackground() {
			hex = v.Dark
		}
		hex = strings.TrimPrefix(hex, "#")
		if len(hex) == 6 {
			var r, g, b uint8
			fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
			return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
		}
	}
	return ""
}

func renderSparkline(samples []Sample, width int, th Theme) string {
	emptyFG := colorToFGANSI(th.SparklineEmpty)
	greenFG := colorToFGANSI(th.SparklineGreen)
	yelFG := colorToFGANSI(th.SparklineYel)
	redFG := colorToFGANSI(th.SparklineRed)

	if len(samples) == 0 {
		return emptyFG + strings.Repeat("░", width) + "\x1b[39m"
	}

	blocks := []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	const compression = 15.0

	now := time.Now()
	oldestTime := samples[0].Timestamp
	maxAge := now.Sub(oldestTime)
	if maxAge < time.Second {
		maxAge = time.Second
	}

	type bucket struct {
		maxRtt   time.Duration
		hasError bool
		count    int
	}
	buckets := make([]bucket, width)
	denom := math.Log(1.0 + compression)

	for _, s := range samples {
		age := now.Sub(s.Timestamp)
		if age < 0 {
			age = 0
		}
		ageRatio := float64(age) / float64(maxAge)
		if ageRatio > 1.0 {
			ageRatio = 1.0
		}

		normDist := math.Log(1.0+compression*ageRatio) / denom
		col := int(math.Round(float64(width-1) * (1.0 - normDist)))

		if col < 0 {
			col = 0
		}
		if col >= width {
			col = width - 1
		}

		b := &buckets[col]
		b.count++
		if s.Rtt < 0 {
			b.hasError = true
		} else if s.Rtt > b.maxRtt {
			b.maxRtt = s.Rtt
		}
	}

	lastValid := bucket{maxRtt: 0, hasError: false, count: 0}
	for i := 0; i < width; i++ {
		if buckets[i].count > 0 {
			lastValid = buckets[i]
		} else if lastValid.count > 0 {
			buckets[i] = lastValid
		}
	}

	var sb strings.Builder
	for i := 0; i < width; i++ {
		b := buckets[i]
		if b.count == 0 {
			sb.WriteString(emptyFG)
			sb.WriteRune('░')
			continue
		}

		if b.hasError {
			sb.WriteString(redFG)
			sb.WriteRune('✕')
			continue
		}

		ms := float64(b.maxRtt.Microseconds()) / 1000.0
		idx := int(math.Min(7, math.Max(0, (ms/150.0)*7)))

		if ms < 40 {
			sb.WriteString(greenFG)
		} else if ms < 120 {
			sb.WriteString(yelFG)
		} else {
			sb.WriteString(redFG)
		}

		sb.WriteRune(blocks[idx])
	}

	sb.WriteString("\x1b[39m")
	return sb.String()
}

func renderBadge(status string, th Theme) string {
	var fg, bg lipgloss.TerminalColor
	var text string

	switch status {
	case "UP":
		fg, bg, text = th.BadgeUpFg, th.BadgeUpBg, " UP "
	case "SLOW":
		fg, bg, text = th.BadgeSlowFg, th.BadgeSlowBg, "SLOW"
	case "DOWN":
		fg, bg, text = th.BadgeDownFg, th.BadgeDownBg, "DOWN"
	default:
		fg, bg, text = th.BadgePendingFg, th.BadgePendingBg, "WAIT"
	}

	fgAnsi := colorToFGANSI(fg)
	bgAnsi := colorToBGANSI(bg)
	return fgAnsi + bgAnsi + "\x1b[1m" + text + "\x1b[22;39;49m"
}

// Messages
type tickMsg time.Time
type pingResultMsg struct {
	index int
	rtt   time.Duration
	err   error
}

// Commands
func tickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func pingHostCmd(index int, host string, port int) tea.Cmd {
	return func() tea.Msg {
		address := net.JoinHostPort(host, strconv.Itoa(port))
		start := time.Now()
		conn, err := net.DialTimeout("tcp", address, 1500*time.Millisecond)
		rtt := time.Since(start)
		if err != nil {
			return pingResultMsg{index: index, rtt: 0, err: err}
		}
		conn.Close()
		return pingResultMsg{index: index, rtt: rtt, err: nil}
	}
}

// Model
type model struct {
	targets    []Target
	cursor     int
	width      int
	height     int
	paused     bool
	addingHost bool
	inputField textinput.Model
	statusMsg  string
	themeIdx   int

	// Bubbles Components
	spinner  spinner.Model
	progress progress.Model
	help     help.Model
	keys     keyMap
}

func initialModel(targets []Target, themeIdx int) model {
	ti := textinput.New()
	ti.Placeholder = "example.com:80"
	ti.CharLimit = 64
	ti.Width = 30

	if len(targets) == 0 {
		targets = []Target{
			newTarget("1.1.1.1", 53),
			newTarget("8.8.8.8", 53),
			newTarget("google.com", 443),
			newTarget("github.com", 443),
			newTarget("127.0.0.1", 80),
		}
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	prog := progress.New(
		progress.WithScaledGradient("#79740e", "#9d0006"),
		progress.WithoutPercentage(),
	)

	h := help.New()

	m := model{
		targets:    targets,
		cursor:     0,
		paused:     false,
		inputField: ti,
		themeIdx:   themeIdx,
		spinner:    sp,
		progress:   prog,
		help:       h,
		keys:       newKeyMap(),
	}

	m.applyThemeStyles()
	return m
}

func (m *model) applyThemeStyles() {
	th := themes[m.themeIdx]
	m.spinner.Style = lipgloss.NewStyle().Foreground(th.InputPrompt)
	m.help.Styles.ShortKey = lipgloss.NewStyle().Foreground(th.FooterKey)
	m.help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(th.FooterDesc)
	m.help.Styles.FullKey = lipgloss.NewStyle().Foreground(th.FooterKey)
	m.help.Styles.FullDesc = lipgloss.NewStyle().Foreground(th.FooterDesc)
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, tickCmd(), m.spinner.Tick)

	for i, t := range m.targets {
		cmds = append(cmds, pingHostCmd(i, t.Host, t.Port))
	}

	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.progress.Width = msg.Width - 25
		if m.progress.Width < 10 {
			m.progress.Width = 10
		}

	case tickMsg:
		if !m.paused {
			for i, t := range m.targets {
				cmds = append(cmds, pingHostCmd(i, t.Host, t.Port))
			}
		}
		cmds = append(cmds, tickCmd())

	case pingResultMsg:
		if msg.index >= 0 && msg.index < len(m.targets) {
			m.targets[msg.index].AddResult(msg.rtt, msg.err)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.addingHost {
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(m.inputField.Value())
				if val != "" {
					host, port := parseHostPort(val, 80)
					m.targets = append(m.targets, newTarget(host, port))
					newIdx := len(m.targets) - 1
					cmds = append(cmds, pingHostCmd(newIdx, host, port))
					m.statusMsg = fmt.Sprintf("Added target %s:%d", host, port)
				}
				m.inputField.Reset()
				m.addingHost = false
				m.inputField.Blur()

			case "esc":
				m.inputField.Reset()
				m.addingHost = false
				m.inputField.Blur()

			default:
				var cmd tea.Cmd
				m.inputField, cmd = m.inputField.Update(msg)
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.targets)-1 {
				m.cursor++
			}

		case key.Matches(msg, m.keys.Add):
			m.addingHost = true
			m.inputField.Focus()
			return m, textinput.Blink

		case key.Matches(msg, m.keys.Delete):
			if len(m.targets) > 0 && m.cursor < len(m.targets) {
				removed := m.targets[m.cursor].Host
				m.targets = append(m.targets[:m.cursor], m.targets[m.cursor+1:]...)
				if m.cursor >= len(m.targets) && m.cursor > 0 {
					m.cursor--
				}
				m.statusMsg = fmt.Sprintf("Removed target %s", removed)
			}

		case key.Matches(msg, m.keys.Pause):
			m.paused = !m.paused
			if m.paused {
				m.statusMsg = "Pings paused"
			} else {
				m.statusMsg = "Resumed pings"
			}

		case key.Matches(msg, m.keys.Theme):
			m.themeIdx = (m.themeIdx + 1) % len(themes)
			m.applyThemeStyles()
			m.statusMsg = fmt.Sprintf("Switched theme to: %s", themes[m.themeIdx].Name)

		case key.Matches(msg, m.keys.Reset):
			for i := range m.targets {
				m.targets[i].Sent = 0
				m.targets[i].Received = 0
				m.targets[i].Samples = nil
				m.targets[i].MinRtt = 0
				m.targets[i].MaxRtt = 0
				m.targets[i].AvgRtt = 0
				m.targets[i].Status = "PENDING"
			}
			m.statusMsg = "Stats reset"

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) renderTable() string {
	th := themes[m.themeIdx]
	var sb strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(th.DetailTitle)
	header := fmt.Sprintf("%-8s %-20s %-6s %-22s %-10s %-10s %-8s",
		"STATUS", "HOST", "PORT", "SPARKLINE (LATENCY)", "LAST RTT", "AVG RTT", "LOSS %",
	)
	sb.WriteString(headerStyle.Render(header) + "\n")

	divider := lipgloss.NewStyle().Foreground(th.CardBorder).Render(strings.Repeat("─", 88))
	sb.WriteString(divider + "\n")

	sparkWidth := 20
	if m.width > 90 {
		sparkWidth = 22
	}

	for i, t := range m.targets {
		cursorStr := " "
		rowStyle := lipgloss.NewStyle()
		if i == m.cursor {
			cursorStr = "❯"
			rowStyle = lipgloss.NewStyle().
				Background(th.CardBorder).
				Foreground(th.HostText).
				Bold(true)
		}

		badge := renderBadge(t.Status, th)
		spark := renderSparkline(t.Samples, sparkWidth, th)

		rttStr := "---"
		if t.LastRtt > 0 && t.Status != "DOWN" {
			rttStr = fmt.Sprintf("%d ms", t.LastRtt.Milliseconds())
		}

		avgStr := "---"
		if t.AvgRtt > 0 {
			avgStr = fmt.Sprintf("%d ms", t.AvgRtt.Milliseconds())
		}

		lossStr := fmt.Sprintf("%.1f%%", t.LossPercentage())

		rowContent := fmt.Sprintf("%s %s %-20s %-6d %s %-10s %-10s %-8s",
			cursorStr,
			badge,
			t.Host,
			t.Port,
			spark,
			rttStr,
			avgStr,
			lossStr,
		)

		sb.WriteString(rowStyle.Render(rowContent) + "\n")
	}

	return sb.String()
}

func (m model) View() string {
	th := themes[m.themeIdx]
	var doc strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(th.TitleFg).Background(th.TitleBg).Padding(0, 1)
	subTitleStyle := lipgloss.NewStyle().Foreground(th.SubTitle).Italic(true)
	statLabel := lipgloss.NewStyle().Foreground(th.StatLabel)
	statVal := lipgloss.NewStyle().Bold(true).Foreground(th.StatVal)

	// Header with animated Spinner
	spinStr := ""
	if !m.paused {
		spinStr = m.spinner.View() + " "
	}
	pauseState := ""
	if m.paused {
		pauseState = " [PAUSED]"
	}
	header := titleStyle.Render("📡 VISUAL PING DASHBOARD") + subTitleStyle.Render(fmt.Sprintf("  %s%d targets | Theme: %s%s", spinStr, len(m.targets), th.Name, pauseState))
	doc.WriteString(header + "\n\n")

	// Table View
	doc.WriteString(m.renderTable() + "\n")

	// Details Panel for Selected Target
	if len(m.targets) > 0 && m.cursor < len(m.targets) {
		selected := m.targets[m.cursor]
		doc.WriteString("\n")

		detailTitle := lipgloss.NewStyle().Bold(true).Foreground(th.DetailTitle).Render(fmt.Sprintf("Target Details: %s:%d", selected.Host, selected.Port))

		lossVal := selected.LossPercentage()
		lossProgressBar := m.progress.ViewAs(lossVal / 100.0)

		stats := fmt.Sprintf("%s %s   %s %s   %s %s   %s %d/%d\n%s %s %.1f%%",
			statLabel.Render("Min:"), statVal.Render(fmt.Sprintf("%dms", selected.MinRtt.Milliseconds())),
			statLabel.Render("Max:"), statVal.Render(fmt.Sprintf("%dms", selected.MaxRtt.Milliseconds())),
			statLabel.Render("Avg:"), statVal.Render(fmt.Sprintf("%dms", selected.AvgRtt.Milliseconds())),
			statLabel.Render("Sent/Recv:"), selected.Sent, selected.Received,
			statLabel.Render("Loss:"), lossProgressBar, lossVal,
		)

		if selected.LastError != "" {
			stats += "\n" + lipgloss.NewStyle().Foreground(th.LastError).Render("Last Error: "+selected.LastError)
		}

		detailWidth := m.width - 4
		if detailWidth < 40 {
			detailWidth = 40
		}

		detailBox := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(th.DetailBorder).
			Padding(0, 1).
			Width(detailWidth).
			Render(detailTitle + "\n" + stats)

		doc.WriteString(detailBox + "\n")
	}

	// Add Host Input Modal / Line
	if m.addingHost {
		doc.WriteString("\n" + lipgloss.NewStyle().Foreground(th.InputPrompt).Render("Enter host to ping (host:port): ") + m.inputField.View() + "\n")
	} else if m.statusMsg != "" {
		doc.WriteString("\n" + lipgloss.NewStyle().Foreground(th.StatusMsg).Italic(true).Render("Status: "+m.statusMsg) + "\n")
	} else {
		doc.WriteString("\n")
	}

	// Adaptive Bubbles Help Footer
	doc.WriteString(m.help.View(m.keys) + "\n")

	return doc.String()
}

func loadConfigFile(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg ConfigFile
	if err := json.Unmarshal(data, &cfg); err == nil {
		return &cfg, nil
	}

	// Simple YAML / Line fallback parser if not JSON
	lines := strings.Split(string(data), "\n")
	cfg = ConfigFile{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if strings.HasPrefix(line, "theme:") {
			cfg.Theme = strings.TrimSpace(strings.TrimPrefix(line, "theme:"))
		} else if strings.HasPrefix(line, "default_port:") {
			if p, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "default_port:"))); err == nil {
				cfg.DefaultPort = p
			}
		} else if strings.HasPrefix(line, "- host:") || strings.HasPrefix(line, "- ") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				host := strings.TrimSpace(parts[1])
				port := 80
				if len(parts) >= 3 {
					if p, err := strconv.Atoi(strings.TrimSpace(parts[2])); err == nil {
						port = p
					}
				}
				if host != "" {
					cfg.Targets = append(cfg.Targets, ConfigTarget{Host: host, Port: port})
				}
			}
		}
	}

	return &cfg, nil
}

func printUsage() {
	fmt.Printf("tboard v%s - Interactive visual ping dashboard TUI\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  tboard [flags] [host[:port] ...]")
	fmt.Println("\nExamples:")
	fmt.Println("  tboard 1.1.1.1 8.8.8.8 google.com:443")
	fmt.Println("  tboard -c ~/.config/tboard/config.yaml")
	fmt.Println("  tboard -t dracula 1.1.1.1 github.com")
	fmt.Println("\nFlags:")
	fmt.Println("  -c, --config <file>   Path to config file (.json or .yaml)")
	fmt.Println("  -p, --port <port>     Default port when omitted (default: 80)")
	fmt.Println("  -t, --theme <name>    Initial theme: gruvbox-light, dracula, auto")
	fmt.Println("  -v, --version         Show version information")
	fmt.Println("  -h, --help            Show help message")
}

func main() {
	var (
		configFile  string
		defaultPort int
		themeName   string
		showVersion bool
		showHelp    bool
	)

	flag.StringVar(&configFile, "c", "", "Path to config file")
	flag.StringVar(&configFile, "config", "", "Path to config file")
	flag.IntVar(&defaultPort, "p", 80, "Default target port")
	flag.IntVar(&defaultPort, "port", 80, "Default target port")
	flag.StringVar(&themeName, "t", "gruvbox-light", "Initial theme")
	flag.StringVar(&themeName, "theme", "gruvbox-light", "Initial theme")
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	flag.Usage = printUsage
	flag.Parse()

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("tboard v%s\n", version)
		os.Exit(0)
	}

	var targets []Target
	selectedThemeIdx := parseThemeName(themeName)

	// Attempt auto-loading config file if explicit config flag is not passed
	if configFile == "" {
		homeDir, _ := os.UserHomeDir()
		candidates := []string{
			"tboard.yaml",
			"tboard.json",
			filepath.Join(homeDir, ".config", "tboard", "config.yaml"),
			filepath.Join(homeDir, ".config", "tboard", "config.json"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				configFile = c
				break
			}
		}
	}

	// Read config file if specified/found
	if configFile != "" {
		cfg, err := loadConfigFile(configFile)
		if err == nil {
			if cfg.DefaultPort > 0 {
				defaultPort = cfg.DefaultPort
			}
			if cfg.Theme != "" && themeName == "gruvbox-light" {
				selectedThemeIdx = parseThemeName(cfg.Theme)
			}
			for _, ct := range cfg.Targets {
				p := ct.Port
				if p <= 0 {
					p = defaultPort
				}
				targets = append(targets, newTarget(ct.Host, p))
			}
		} else if configFile != "" && flag.Lookup("c").Value.String() != "" {
			fmt.Fprintf(os.Stderr, "Error loading config file %s: %v\n", configFile, err)
			os.Exit(1)
		}
	}

	// Positional CLI host arguments (override/append targets)
	posArgs := flag.Args()
	if len(posArgs) > 0 {
		cliTargets := make([]Target, 0, len(posArgs))
		for _, arg := range posArgs {
			host, port := parseHostPort(arg, defaultPort)
			cliTargets = append(cliTargets, newTarget(host, port))
		}
		targets = cliTargets
	}

	p := tea.NewProgram(initialModel(targets, selectedThemeIdx), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running ping dashboard: %v\n", err)
		os.Exit(1)
	}
}
