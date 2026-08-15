package probe

import (
	"testing"
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
		h, p := ParseHostPort(tt.input, tt.defaultPort)
		if h != tt.expectedHost || p != tt.expectedPort {
			t.Errorf("ParseHostPort(%q, %d) = (%q, %d), expected (%q, %d)",
				tt.input, tt.defaultPort, h, p, tt.expectedHost, tt.expectedPort)
		}
	}
}
