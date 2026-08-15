package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines terminal colors for all dashboard elements.
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

var Themes = []Theme{
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

// ParseThemeName resolves string names to theme palette indices.
func ParseThemeName(name string) int {
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

// ColorToFGANSI formats a foreground-only ANSI escape code.
func ColorToFGANSI(c lipgloss.TerminalColor) string {
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

// ColorToBGANSI formats a background-only ANSI escape code.
func ColorToBGANSI(c lipgloss.TerminalColor) string {
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
