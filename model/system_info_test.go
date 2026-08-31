package model

import (
	"encoding/json"
	"testing"
)

func TestSystemInfoJSONRoundTrip(t *testing.T) {
	in := SystemInfo{
		CPU:     CPUInfo{Model: "AMD Ryzen 5", Cores: 6, Threads: 12, MaxFreqMHz: 4000},
		Memory:  MemoryInfo{TotalGB: 16.0, AvailableGB: 8.0, UsedGB: 8.0, UsedPercent: 50},
		Storage: []StorageInfo{{Device: "/dev/sda", MountPoint: "/", Type: "ext4", TotalGB: 500, UsedGB: 250, FreeGB: 250, UsedPercent: 50}},
		GPU:     []GPUInfo{{Name: "RTX 3080", Vendor: "NVIDIA", MemoryGB: 10, Driver: "525"}},
	}

	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var out SystemInfo
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if out.CPU.Model != in.CPU.Model || out.Memory.TotalGB != in.Memory.TotalGB {
		t.Errorf("round-trip mismatch: %+v", out)
	}
	if len(out.Storage) != 1 || out.Storage[0].Device != "/dev/sda" {
		t.Errorf("storage round-trip mismatch: %+v", out.Storage)
	}
	if len(out.GPU) != 1 || out.GPU[0].Name != "RTX 3080" {
		t.Errorf("gpu round-trip mismatch: %+v", out.GPU)
	}
}

func TestSystemInfoEmptySlices(t *testing.T) {
	in := SystemInfo{CPU: CPUInfo{Model: "Test"}, Storage: []StorageInfo{}, GPU: []GPUInfo{}}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var out SystemInfo
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if out.Storage == nil || out.GPU == nil {
		t.Errorf("empty slices should survive round-trip: %+v", out)
	}
}
