package windows

import "testing"

func TestDetectVendorFromName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"NVIDIA GeForce RTX 3080", "NVIDIA"},
		{"NVIDIA Quadro RTX 4000", "NVIDIA"},
		{"AMD Radeon RX 6800", "AMD"},
		{"Intel UHD Graphics", "Intel"},
		{"Unknown Adapter", "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectVendorFromName(tt.name); got != tt.want {
				t.Errorf("detectVendorFromName(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
