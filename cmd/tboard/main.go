package main

import (
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Target represents a host monitored by the ping dashboard.
type Target struct {
	Host      string
	Port      int
	Rtts      []time.Duration // History of RTTs (-1 indicates timeout/error)
	Sent      int
	Received  int
	LastRtt   time.Duration
	MinRtt    time.Duration
	MaxRtt    time.Duration
	AvgRtt    time.Duration
	Status    string // "UP", "SLOW", "DOWN"
	LastError string
}

func newTarget(host string, port int) Target {
	return Target{
		Host:   host,
		Port:   port,
		Rtts:   make([]time.Duration, 0, 30),
		Status: "PENDING",
	}
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
	const maxHistory = 30
	if err != nil {
		t.LastError = err.Error()
		t.Rtts = append(t.Rtts, -1)
		t.Status = "DOWN"
	} else {
		t.Received++
		t.LastRtt = rtt
		t.Rtts = append(t.Rtts, rtt)
		t.LastError = ""

		if t.MinRtt == 0 || rtt < t.MinRtt {
			t.MinRtt = rtt
		}
		if rtt > t.MaxRtt {
			t.MaxRtt = rtt
		}

		var total time.Duration
		validCount := 0
		for _, r := range t.Rtts {
			if r >= 0 {
				total += r
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

	if len(t.Rtts) > maxHistory {
		t.Rtts = t.Rtts[1:]
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
			key.WithHelp("j/k/↑↓", "select"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
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
	return []key.Binding{k.Up, k.Add, k.Delete, k.Theme, k.Pause, k.Reset, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Add, k.Delete},
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
	width      int
	height     int
	paused     bool
	addingHost bool
	inputField textinput.Model
	statusMsg  string
	themeIdx   int

	// Bubbles Components
	table    table.Model
	spinner  spinner.Model
	progress progress.Model
	help     help.Model
	keys     keyMap
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "example.com:80"
	ti.CharLimit = 64
	ti.Width = 30

	defaultTargets := []Target{
		newTarget("1.1.1.1", 53),
		newTarget("8.8.8.8", 53),
		newTarget("google.com", 443),
		newTarget("github.com", 443),
		newTarget("127.0.0.1", 80),
	}

	// Spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(themes[1].InputPrompt)

	// Table
	cols := []table.Column{
		{Title: "STATUS", Width: 8},
		{Title: "HOST", Width: 18},
		{Title: "PORT", Width: 6},
		{Title: "SPARKLINE (LATENCY)", Width: 22},
		{Title: "LAST RTT", Width: 10},
		{Title: "AVG RTT", Width: 10},
		{Title: "LOSS %", Width: 10},
	}

	tbl := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	// Progress
	prog := progress.New(
		progress.WithScaledGradient("#79740e", "#9d0006"),
		progress.WithoutPercentage(),
	)

	// Help
	h := help.New()

	m := model{
		targets:    defaultTargets,
		paused:     false,
		inputField: ti,
		themeIdx:   1, // Default Gruvbox Light
		table:      tbl,
		spinner:    sp,
		progress:   prog,
		help:       h,
		keys:       newKeyMap(),
	}

	m.applyThemeStyles()
	m.updateTableRows()
	return m
}

func (m *model) applyThemeStyles() {
	th := themes[m.themeIdx]

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(th.CardBorder).
		BorderBottom(true).
		Bold(true).
		Foreground(th.DetailTitle)
	s.Selected = s.Selected.
		Foreground(th.HostText).
		Background(th.CardBorder).
		Bold(true)
	m.table.SetStyles(s)

	m.spinner.Style = lipgloss.NewStyle().Foreground(th.InputPrompt)
	m.help.Styles.ShortKey = lipgloss.NewStyle().Foreground(th.FooterKey)
	m.help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(th.FooterDesc)
	m.help.Styles.FullKey = lipgloss.NewStyle().Foreground(th.FooterKey)
	m.help.Styles.FullDesc = lipgloss.NewStyle().Foreground(th.FooterDesc)
}

func (m *model) updateTableRows() {
	th := themes[m.themeIdx]
	rows := make([]table.Row, len(m.targets))
	for i, t := range m.targets {
		badge := renderBadge(t.Status, th)
		spark := renderSparkline(t.Rtts, 20, th)

		rttStr := "---"
		if t.LastRtt > 0 && t.Status != "DOWN" {
			rttStr = fmt.Sprintf("%d ms", t.LastRtt.Milliseconds())
		}

		avgStr := "---"
		if t.AvgRtt > 0 {
			avgStr = fmt.Sprintf("%d ms", t.AvgRtt.Milliseconds())
		}

		lossStr := fmt.Sprintf("%.1f%%", t.LossPercentage())

		rows[i] = table.Row{
			badge,
			t.Host,
			strconv.Itoa(t.Port),
			spark,
			rttStr,
			avgStr,
			lossStr,
		}
	}
	m.table.SetRows(rows)
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
			m.updateTableRows()
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
					host := val
					port := 80
					if strings.Contains(val, ":") {
						parts := strings.Split(val, ":")
						host = parts[0]
						if p, err := strconv.Atoi(parts[1]); err == nil {
							port = p
						}
					}
					m.targets = append(m.targets, newTarget(host, port))
					m.updateTableRows()
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

		case key.Matches(msg, m.keys.Add):
			m.addingHost = true
			m.inputField.Focus()
			return m, textinput.Blink

		case key.Matches(msg, m.keys.Delete):
			cursor := m.table.Cursor()
			if len(m.targets) > 0 && cursor < len(m.targets) {
				removed := m.targets[cursor].Host
				m.targets = append(m.targets[:cursor], m.targets[cursor+1:]...)
				m.updateTableRows()
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
			m.updateTableRows()
			m.statusMsg = fmt.Sprintf("Switched theme to: %s", themes[m.themeIdx].Name)

		case key.Matches(msg, m.keys.Reset):
			for i := range m.targets {
				m.targets[i].Sent = 0
				m.targets[i].Received = 0
				m.targets[i].Rtts = nil
				m.targets[i].MinRtt = 0
				m.targets[i].MaxRtt = 0
				m.targets[i].AvgRtt = 0
				m.targets[i].Status = "PENDING"
			}
			m.updateTableRows()
			m.statusMsg = "Stats reset"

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		default:
			// Let table handle navigation (up/down/j/k)
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func renderSparkline(rtts []time.Duration, width int, th Theme) string {
	if len(rtts) == 0 {
		return lipgloss.NewStyle().Foreground(th.SparklineEmpty).Render(strings.Repeat("░", width))
	}
	blocks := []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var sb strings.Builder

	slice := rtts
	if len(slice) > width {
		slice = slice[len(slice)-width:]
	}

	for i := 0; i < width-len(slice); i++ {
		sb.WriteString(lipgloss.NewStyle().Foreground(th.SparklineEmpty).Render("░"))
	}

	for _, rtt := range slice {
		if rtt < 0 {
			sb.WriteString(lipgloss.NewStyle().Foreground(th.SparklineRed).Render("✕"))
			continue
		}

		ms := float64(rtt.Microseconds()) / 1000.0
		idx := int(math.Min(7, math.Max(0, (ms/150.0)*7)))

		var color lipgloss.TerminalColor
		if ms < 40 {
			color = th.SparklineGreen
		} else if ms < 120 {
			color = th.SparklineYel
		} else {
			color = th.SparklineRed
		}

		sb.WriteString(lipgloss.NewStyle().Foreground(color).Render(string(blocks[idx])))
	}

	return sb.String()
}

func renderBadge(status string, th Theme) string {
	badgeStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)

	switch status {
	case "UP":
		return badgeStyle.Foreground(th.BadgeUpFg).Background(th.BadgeUpBg).Render("UP")
	case "SLOW":
		return badgeStyle.Foreground(th.BadgeSlowFg).Background(th.BadgeSlowBg).Render("SLOW")
	case "DOWN":
		return badgeStyle.Foreground(th.BadgeDownFg).Background(th.BadgeDownBg).Render("DOWN")
	default:
		return badgeStyle.Foreground(th.BadgePendingFg).Background(th.BadgePendingBg).Render("WAIT")
	}
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
	doc.WriteString(m.table.View() + "\n")

	// Details Panel for Selected Target
	cursor := m.table.Cursor()
	if len(m.targets) > 0 && cursor < len(m.targets) {
		selected := m.targets[cursor]
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

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running ping dashboard: %v\n", err)
		os.Exit(1)
	}
}
