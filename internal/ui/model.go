package ui

import (
	"fmt"
	"strings"

	"tboard/internal/domain"
	"tboard/internal/probe"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewMode int

const (
	View3DStack ViewMode = iota
	ViewSplit
	ViewExpandedChart
)

// Model represents the top-level Bubble Tea state.
type Model struct {
	targets    []domain.Target
	cursor     int
	width      int
	height     int
	paused     bool
	addingHost bool
	inputField textinput.Model
	statusMsg  string
	themeIdx   int
	viewMode   ViewMode

	// 3D Isometric Ridgeline Stack Camera Parameters
	depthX    int
	depthY    int
	scaleY    float64
	fillStyle FillStyle

	// Bubbles & ntcharts Components
	spinner  spinner.Model
	progress progress.Model
	help     help.Model
	keys     KeyMap
}

// NewModel constructs a Model initialized with targets and theme index.
func NewModel(initialTargets []domain.Target, themeIdx int) Model {
	ti := textinput.New()
	ti.Placeholder = "example.com:80"
	ti.CharLimit = 64
	ti.Width = 30

	if len(initialTargets) == 0 {
		initialTargets = []domain.Target{
			domain.NewTarget("1.1.1.1", 53, 50, 10, 0),
			domain.NewTarget("8.8.8.8", 53, 50, 10, 1),
			domain.NewTarget("google.com", 443, 50, 10, 2),
			domain.NewTarget("github.com", 443, 50, 10, 3),
			domain.NewTarget("127.0.0.1", 80, 50, 10, 4),
		}
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	prog := progress.New(
		progress.WithScaledGradient("#79740e", "#9d0006"),
		progress.WithoutPercentage(),
	)

	h := help.New()

	m := Model{
		targets:    initialTargets,
		cursor:     0,
		paused:     false,
		inputField: ti,
		themeIdx:   themeIdx,
		viewMode:   View3DStack,
		depthX:     3,
		depthY:     2,
		scaleY:     0.12,
		fillStyle:  FillSolid,
		spinner:    sp,
		progress:   prog,
		help:       h,
		keys:       NewKeyMap(),
		statusMsg:  "3D Z-Axis Ridgeline Stack View Active",
	}

	m.applyThemeStyles()
	return m
}

func (m *Model) applyThemeStyles() {
	th := Themes[m.themeIdx]
	m.spinner.Style = lipgloss.NewStyle().Foreground(th.InputPrompt)
	m.help.Styles.ShortKey = lipgloss.NewStyle().Foreground(th.FooterKey)
	m.help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(th.FooterDesc)
	m.help.Styles.FullKey = lipgloss.NewStyle().Foreground(th.FooterKey)
	m.help.Styles.FullDesc = lipgloss.NewStyle().Foreground(th.FooterDesc)
}

func (m *Model) resizeCharts() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	chartW := m.width - 44
	if chartW < 20 {
		chartW = 20
	}
	chartH := 10
	if m.viewMode == ViewExpandedChart {
		chartW = m.width - 6
		chartH = m.height - 10
		if chartH < 6 {
			chartH = 6
		}
	}

	for i := range m.targets {
		m.targets[i].Chart.Resize(chartW, chartH)
		m.targets[i].Chart.Draw()
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, probe.TickCmd(), m.spinner.Tick)

	for i, t := range m.targets {
		cmds = append(cmds, probe.PingHostCmd(i, t.Host, t.Port))
	}

	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.resizeCharts()

	case probe.TickMsg:
		if !m.paused {
			for i, t := range m.targets {
				cmds = append(cmds, probe.PingHostCmd(i, t.Host, t.Port))
			}
		}
		cmds = append(cmds, probe.TickCmd())

	case probe.PingResultMsg:
		if msg.TargetIdx >= 0 && msg.TargetIdx < len(m.targets) {
			m.targets[msg.TargetIdx].AddResult(msg.Rtt, msg.Err)
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
					host, port := probe.ParseHostPort(val, 80)
					chartW := m.width - 44
					if chartW < 20 {
						chartW = 20
					}
					nt := domain.NewTarget(host, port, chartW, 10, len(m.targets))
					m.targets = append(m.targets, nt)
					newIdx := len(m.targets) - 1
					cmds = append(cmds, probe.PingHostCmd(newIdx, host, port))
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

		case key.Matches(msg, m.keys.ToggleView):
			m.viewMode = (m.viewMode + 1) % 3
			switch m.viewMode {
			case View3DStack:
				m.statusMsg = "Switched to 3D Z-Axis Ridgeline Stack View"
			case ViewSplit:
				m.statusMsg = "Switched to Split Table/Chart View"
			case ViewExpandedChart:
				m.statusMsg = "Switched to Expanded ntcharts View"
			}
			m.resizeCharts()

		case key.Matches(msg, m.keys.CamUp):
			m.depthY++
			if m.depthY > 6 {
				m.depthY = 6
			}
			m.statusMsg = fmt.Sprintf("3D Depth Y: %d", m.depthY)

		case key.Matches(msg, m.keys.CamDown):
			m.depthY--
			if m.depthY < 0 {
				m.depthY = 0
			}
			m.statusMsg = fmt.Sprintf("3D Depth Y: %d", m.depthY)

		case key.Matches(msg, m.keys.CamLeft):
			m.depthX--
			if m.depthX < 0 {
				m.depthX = 0
			}
			m.statusMsg = fmt.Sprintf("3D Slant X: %d", m.depthX)

		case key.Matches(msg, m.keys.CamRight):
			m.depthX++
			if m.depthX > 8 {
				m.depthX = 8
			}
			m.statusMsg = fmt.Sprintf("3D Slant X: %d", m.depthX)

		case key.Matches(msg, m.keys.ScaleUp):
			m.scaleY *= 1.25
			if m.scaleY > 1.5 {
				m.scaleY = 1.5
			}
			m.statusMsg = fmt.Sprintf("3D Height Scale: %.2f", m.scaleY)

		case key.Matches(msg, m.keys.ScaleDown):
			m.scaleY /= 1.25
			if m.scaleY < 0.02 {
				m.scaleY = 0.02
			}
			m.statusMsg = fmt.Sprintf("3D Height Scale: %.2f", m.scaleY)

		case key.Matches(msg, m.keys.ToggleMesh):
			m.fillStyle = (m.fillStyle + 1) % 4
			names := map[FillStyle]string{
				FillSolid:     "Solid Block (█)",
				FillMedium:    "Medium Shade (▓)",
				FillLight:     "Light Shade (▒)",
				FillWireframe: "Wireframe Contour (━)",
			}
			m.statusMsg = fmt.Sprintf("3D Mesh Fill Style: %s", names[m.fillStyle])

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
			m.themeIdx = (m.themeIdx + 1) % len(Themes)
			m.applyThemeStyles()
			m.statusMsg = fmt.Sprintf("Switched theme to: %s", Themes[m.themeIdx].Name)

		case key.Matches(msg, m.keys.Reset):
			for i := range m.targets {
				chartW := m.width - 44
				if chartW < 20 {
					chartW = 20
				}
				m.targets[i].Reset(chartW, 10)
			}
			m.statusMsg = "Stats & Charts reset"

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) renderTable() string {
	th := Themes[m.themeIdx]
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

		badge := RenderBadge(t.Status, th)
		spark := RenderSparkline(t.Samples, sparkWidth, th)

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

func (m Model) renderChartPanel(t domain.Target) string {
	th := Themes[m.themeIdx]

	chartTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(th.DetailTitle).
		Render(fmt.Sprintf("📊 Live Latency Stream (ntcharts) - %s:%d", t.Host, t.Port))

	chartStr := t.Chart.View()
	if strings.TrimSpace(chartStr) == "" {
		chartStr = lipgloss.NewStyle().Foreground(th.StatLabel).Render("Waiting for ping samples...")
	}

	lossPct := t.LossPercentage() / 100.0
	lossBar := m.progress.ViewAs(lossPct)

	minMs, maxMs, avgMs, lastMs := "---", "---", "---", "---"
	if t.MinRtt > 0 {
		minMs = fmt.Sprintf("%d ms", t.MinRtt.Milliseconds())
	}
	if t.MaxRtt > 0 {
		maxMs = fmt.Sprintf("%d ms", t.MaxRtt.Milliseconds())
	}
	if t.AvgRtt > 0 {
		avgMs = fmt.Sprintf("%d ms", t.AvgRtt.Milliseconds())
	}
	if t.LastRtt > 0 && t.Status != "DOWN" {
		lastMs = fmt.Sprintf("%d ms", t.LastRtt.Milliseconds())
	}

	statsBox := fmt.Sprintf(
		"Status: %s  |  Last: %s  |  Min: %s  |  Max: %s  |  Avg: %s\n"+
			"Packets: Sent %d, Recv %d, Lost %d (%.1f%%)\nLoss: %s",
		t.Status, lastMs, minMs, maxMs, avgMs,
		t.Sent, t.Received, t.Sent-t.Received, t.LossPercentage(), lossBar,
	)

	if t.LastError != "" {
		statsBox += fmt.Sprintf("\n%s", lipgloss.NewStyle().Foreground(th.LastError).Render("Error: "+t.LastError))
	}

	panelContent := lipgloss.JoinVertical(lipgloss.Left,
		chartTitle,
		chartStr,
		lipgloss.NewStyle().Foreground(th.CardBorder).Render("──────────────────────────────────────────────────"),
		statsBox,
	)

	borderCol := th.DetailBorder
	if t.Status == "SLOW" {
		borderCol = th.BadgeSlowBg
	} else if t.Status == "DOWN" {
		borderCol = th.BadgeDownBg
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Padding(0, 1).
		Render(panelContent)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing 3D Ping Dashboard..."
	}

	th := Themes[m.themeIdx]
	var doc strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(th.TitleFg).Background(th.TitleBg).Padding(0, 1)
	subTitleStyle := lipgloss.NewStyle().Foreground(th.SubTitle).Italic(true)

	spinStr := ""
	if !m.paused {
		spinStr = m.spinner.View() + " "
	}
	pauseState := ""
	if m.paused {
		pauseState = " [PAUSED]"
	}

	viewName := "3D Waterfall Stack"
	if m.viewMode == ViewSplit {
		viewName = "Split View"
	} else if m.viewMode == ViewExpandedChart {
		viewName = "2D ntcharts View"
	}

	header := titleStyle.Render("📡 VISUAL PING DASHBOARD") + subTitleStyle.Render(fmt.Sprintf("  %s%d targets | %s | Theme: %s%s", spinStr, len(m.targets), viewName, th.Name, pauseState))
	doc.WriteString(header + "\n\n")

	var selectedTarget domain.Target
	if len(m.targets) > 0 && m.cursor >= 0 && m.cursor < len(m.targets) {
		selectedTarget = m.targets[m.cursor]
	}

	switch m.viewMode {
	case View3DStack:
		doc.WriteString(Render3DStackView(m.targets, m.width, m.height, m.depthX, m.depthY, m.scaleY, m.fillStyle, th) + "\n")

	case ViewExpandedChart:
		if len(m.targets) > 0 {
			doc.WriteString(m.renderChartPanel(selectedTarget) + "\n")
		} else {
			doc.WriteString(lipgloss.NewStyle().Foreground(th.LastError).Render("No targets configured. Press 'a' to add a host.") + "\n")
		}

	case ViewSplit:
		doc.WriteString(m.renderTable() + "\n")
		if len(m.targets) > 0 {
			doc.WriteString(m.renderChartPanel(selectedTarget) + "\n")
		}
	}

	if m.addingHost {
		doc.WriteString("\n" + lipgloss.NewStyle().Foreground(th.InputPrompt).Render("Enter host to ping (host:port): ") + m.inputField.View() + "\n")
	} else if m.statusMsg != "" {
		doc.WriteString("\n" + lipgloss.NewStyle().Foreground(th.StatusMsg).Italic(true).Render("Status: "+m.statusMsg) + "\n")
	} else {
		doc.WriteString("\n")
	}

	doc.WriteString(m.help.View(m.keys) + "\n")

	return doc.String()
}
