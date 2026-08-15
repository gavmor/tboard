package domain

import (
	"errors"
	"testing"
	"time"
)

func TestTargetLossPercentage(t *testing.T) {
	target := NewTarget("1.1.1.1", 53, 50, 10, 0)
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
	target := NewTarget("google.com", 443, 50, 10, 0)

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

func TestTargetReset(t *testing.T) {
	target := NewTarget("github.com", 443, 50, 10, 0)
	target.AddResult(20*time.Millisecond, nil)
	target.Reset(50, 10)

	if target.Sent != 0 || target.Received != 0 || len(target.Samples) != 0 {
		t.Errorf("expected clean stats after reset, got sent=%d, recv=%d, samples=%d",
			target.Sent, target.Received, len(target.Samples))
	}
}
