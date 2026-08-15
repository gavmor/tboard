package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"tboard/internal/domain"

	"github.com/charmbracelet/lipgloss"
)

type FillStyle int

const (
	FillSolid FillStyle = iota
	FillMedium
	FillLight
	FillWireframe
)

type Cell3D struct {
	Ch    rune
	Style lipgloss.Style
}

// RenderSparkline renders a logarithmic time-axis sparkline with peak-hold bucket aggregation.
func RenderSparkline(samples []domain.Sample, width int, th Theme) string {
	emptyFG := ColorToFGANSI(th.SparklineEmpty)
	greenFG := ColorToFGANSI(th.SparklineGreen)
	yelFG := ColorToFGANSI(th.SparklineYel)
	redFG := ColorToFGANSI(th.SparklineRed)

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

// RenderBadge builds a styled status badge.
func RenderBadge(status string, th Theme) string {
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

	fgAnsi := ColorToFGANSI(fg)
	bgAnsi := ColorToBGANSI(bg)
	return fgAnsi + bgAnsi + "\x1b[1m" + text + "\x1b[22;39;49m"
}

// Render3DStackView renders pseudo-3D Z-axis stacked ridgeline waterfall chart with Painter's Algorithm occlusion.
func Render3DStackView(targets []domain.Target, width, height, depthX, depthY int, scaleY float64, fillStyle FillStyle, th Theme) string {
	canvasW := width - 4
	if canvasW < 40 {
		canvasW = 40
	}
	canvasH := height - 10
	if canvasH < 12 {
		canvasH = 12
	}

	grid := make([][]Cell3D, canvasH)
	for y := 0; y < canvasH; y++ {
		grid[y] = make([]Cell3D, canvasW)
		for x := 0; x < canvasW; x++ {
			grid[y][x] = Cell3D{Ch: ' ', Style: lipgloss.NewStyle()}
		}
	}

	var fillRune rune
	switch fillStyle {
	case FillSolid:
		fillRune = '█'
	case FillMedium:
		fillRune = '▓'
	case FillLight:
		fillRune = '▒'
	case FillWireframe:
		fillRune = ' '
	}

	// PAINTER'S ALGORITHM: Render targets from Backmost (index len-1) down to Frontmost (index 0)
	for i := len(targets) - 1; i >= 0; i-- {
		target := targets[i]
		z := i

		offsetX := (len(targets) - 1 - z) * depthX
		offsetY := (len(targets) - 1 - z) * depthY
		baseY := canvasH - 3 - offsetY

		if baseY < 2 || baseY >= canvasH {
			continue
		}

		targetColor := target.ColorHex
		if target.Status == "DOWN" {
			targetColor = "#9d0006"
		}

		bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(targetColor))
		topStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(targetColor)).Bold(true)
		dimStyle := lipgloss.NewStyle().Foreground(th.CardBorder)

		sampleCount := len(target.Samples)
		maxSamplesWidth := canvasW - offsetX - 18
		if maxSamplesWidth > 120 {
			maxSamplesWidth = 120
		}
		if maxSamplesWidth < 10 {
			maxSamplesWidth = 10
		}

		startSampleIdx := 0
		if sampleCount > maxSamplesWidth {
			startSampleIdx = sampleCount - maxSamplesWidth
		}

		visibleLen := sampleCount - startSampleIdx
		if visibleLen <= 0 {
			for x := 0; x < 20; x++ {
				px := offsetX + 14 + x
				if px >= 0 && px < canvasW && baseY >= 0 && baseY < canvasH {
					grid[baseY][px] = Cell3D{Ch: '┄', Style: dimStyle}
				}
			}
		} else {
			for step := 0; step < visibleLen; step++ {
				s := target.Samples[startSampleIdx+step]
				px := offsetX + 14 + step
				if px < 0 || px >= canvasW {
					continue
				}

				var h int
				if s.Rtt < 0 {
					h = 0
				} else {
					ms := float64(s.Rtt.Milliseconds())
					h = int(math.Round(ms * scaleY))
					if h < 1 {
						h = 1
					}
				}

				if h > baseY-2 {
					h = baseY - 2
				}
				topY := baseY - h

				// Solid Occlusion Fill
				for py := topY; py <= baseY; py++ {
					if py < 0 || py >= canvasH {
						continue
					}

					var r rune
					var st lipgloss.Style

					if s.Rtt < 0 {
						r = '✕'
						st = lipgloss.NewStyle().Foreground(th.SparklineRed).Bold(true)
					} else if py == topY {
						r = '▀'
						st = topStyle
					} else if py == baseY {
						r = '▄'
						st = dimStyle
					} else {
						r = fillRune
						st = bodyStyle
					}

					grid[py][px] = Cell3D{Ch: r, Style: st}
				}
			}
		}

		// Host Name Tag attached to 3D Baseline
		tagText := fmt.Sprintf(" ▶ %-10s ", fmt.Sprintf("%s:%d", target.Host, target.Port))
		tagStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(targetColor)).
			Foreground(th.HostText).
			Bold(true)

		tagRunes := []rune(tagText)
		for tIdx, r := range tagRunes {
			px := offsetX + tIdx
			if px >= 0 && px < canvasW && baseY >= 0 && baseY < canvasH {
				grid[baseY][px] = Cell3D{Ch: r, Style: tagStyle}
			}
		}
	}

	var sb strings.Builder
	for y := 0; y < canvasH; y++ {
		for x := 0; x < canvasW; x++ {
			cell := grid[y][x]
			if cell.Ch == ' ' {
				sb.WriteRune(' ')
			} else {
				sb.WriteString(cell.Style.Render(string(cell.Ch)))
			}
		}
		if y < canvasH-1 {
			sb.WriteRune('\n')
		}
	}

	panelTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(th.DetailTitle).
		Render("🏔  3D Z-AXIS STACKED LATENCY WATERFALL")

	content := lipgloss.JoinVertical(lipgloss.Left,
		panelTitle,
		sb.String(),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(th.DetailBorder).
		Padding(0, 1).
		Width(width - 2).
		Height(height - 8).
		Render(content)
}
