package darwin

import "testing"

func TestParseVmStatLine(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want uint64
	}{
		{"pages free", "Pages free:                    123456.", 123456},
		{"no trailing dot", "Pages active: 42", 42},
		{"invalid", "Pages wired down:", 0},
		{"empty", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseVmStatLine(tt.in); got != tt.want {
				t.Errorf("parseVmStatLine(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
