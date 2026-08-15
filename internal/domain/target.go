package domain

import (
	"time"

	"github.com/NimbleMarkets/ntcharts/linechart/streamlinechart"
	"github.com/NimbleMarkets/ntcharts/sparkline"
)

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
	ColorHex  string // Theme color for 3D Z-stack rendering

	// ntcharts models
	Chart     streamlinechart.Model
	Sparkline sparkline.Model
}

var TargetPalette = []string{
	"#8BE9FD", // Cyan
	"#50FA7B", // Green
	"#BD93F9", // Purple
	"#FF79C6", // Pink
	"#F1FA8C", // Yellow
	"#FFB86C", // Orange
}

// NewTarget constructs a Target instance initialized with ntcharts models.
func NewTarget(host string, port int, chartWidth, chartHeight int, colorIdx int) Target {
	stChart := streamlinechart.New(chartWidth, chartHeight)
	spChart := sparkline.New(16, 1)
	color := TargetPalette[colorIdx%len(TargetPalette)]

	return Target{
		Host:      host,
		Port:      port,
		Samples:   make([]Sample, 0, 180),
		Status:    "PENDING",
		ColorHex:  color,
		Chart:     stChart,
		Sparkline: spChart,
	}
}

// LossPercentage calculates the packet loss percentage.
func (t Target) LossPercentage() float64 {
	if t.Sent == 0 {
		return 0.0
	}
	lost := t.Sent - t.Received
	return (float64(lost) / float64(t.Sent)) * 100.0
}

// AddResult records a latency measurement and updates Target stats.
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
		t.Chart.Push(0)
		t.Sparkline.Push(0)
	} else {
		t.Received++
		t.LastRtt = rtt
		sample.Rtt = rtt
		t.LastError = ""

		ms := float64(rtt.Milliseconds())
		t.Chart.Push(ms)
		t.Sparkline.Push(ms)

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

	t.Chart.Draw()
	t.Sparkline.Draw()

	t.Samples = append(t.Samples, sample)
	if len(t.Samples) > maxHistory {
		t.Samples = t.Samples[1:]
	}
}

// Reset clears all recorded stats and history for a target.
func (t *Target) Reset(chartWidth, chartHeight int) {
	t.Sent = 0
	t.Received = 0
	t.Samples = nil
	t.MinRtt = 0
	t.MaxRtt = 0
	t.AvgRtt = 0
	t.Status = "PENDING"
	t.LastError = ""
	t.Chart = streamlinechart.New(chartWidth, chartHeight)
	t.Sparkline = sparkline.New(16, 1)
}
