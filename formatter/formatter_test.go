package formatter

import (
	"strings"
	"testing"

	"wkqcosoft.com/m/model"
)

// TestConsoleFormatter_Format 테스트는 ConsoleFormatter의 Format 메서드를 검증합니다
func TestConsoleFormatter_Format(t *testing.T) {
	// 테스트용 시스템 정보 준비
	testInfo := &model.SystemInfo{
		CPU: model.CPUInfo{
			Model:      "Intel Core i7-9700K",
			Cores:      8,
			Threads:    8,
			MaxFreqMHz: 3600,
		},
		Memory: model.MemoryInfo{
			TotalGB:     16.0,
			UsedGB:      8.5,
			AvailableGB: 7.5,
			UsedPercent: 53.125,
		},
		Storage: []model.StorageInfo{
			{
				Device:      "/dev/sda1",
				MountPoint:  "/",
				Type:        "ext4",
				TotalGB:     500.0,
				UsedGB:      250.0,
				FreeGB:      250.0,
				UsedPercent: 50.0,
			},
		},
		GPU: []model.GPUInfo{
			{
				Name:     "NVIDIA GeForce RTX 3080",
				Vendor:   "NVIDIA",
				MemoryGB: 10.0,
				Driver:   "470.57.02",
			},
		},
	}

	formatter := NewConsoleFormatter()
	result := formatter.Format(testInfo)

	// 결과 검증: 주요 정보가 포함되어 있는지 확인
	requiredStrings := []string{
		"PC 사양 정보",
		"CPU 정보",
		"Intel Core i7-9700K",
		"메모리 (RAM) 정보",
		"16.00 GB",
		"저장장치 정보",
		"/dev/sda1",
		"GPU 정보",
		"NVIDIA GeForce RTX 3080",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(result, required) {
			t.Errorf("포맷된 출력에 '%s'가 포함되어 있지 않습니다", required)
		}
	}
}

// TestConsoleFormatter_FormatCPU 테스트는 CPU 정보 포맷팅을 검증합니다
func TestConsoleFormatter_FormatCPU(t *testing.T) {
	cpu := &model.CPUInfo{
		Model:      "AMD Ryzen 9 5900X",
		Cores:      12,
		Threads:    24,
		MaxFreqMHz: 4800,
	}

	formatter := NewConsoleFormatter()
	result := formatter.FormatCPU(cpu)

	expected := []string{"AMD Ryzen 9 5900X", "12코어", "24스레드", "4800 MHz"}
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("CPU 포맷 결과에 '%s'가 포함되어 있지 않습니다. 실제: %s", exp, result)
		}
	}
}

// TestConsoleFormatter_FormatMemory 테스트는 메모리 정보 포맷팅을 검증합니다
func TestConsoleFormatter_FormatMemory(t *testing.T) {
	memory := &model.MemoryInfo{
		TotalGB:     32.0,
		UsedGB:      16.0,
		AvailableGB: 16.0,
		UsedPercent: 50.0,
	}

	formatter := NewConsoleFormatter()
	result := formatter.FormatMemory(memory)

	expected := []string{"16.00 GB", "32.00 GB", "50.0%"}
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("메모리 포맷 결과에 '%s'가 포함되어 있지 않습니다. 실제: %s", exp, result)
		}
	}
}

// TestConsoleFormatter_FormatStorage 테스트는 저장장치 정보 포맷팅을 검증합니다
func TestConsoleFormatter_FormatStorage(t *testing.T) {
	storage := &model.StorageInfo{
		Device:      "/dev/nvme0n1p1",
		MountPoint:  "/home",
		Type:        "ext4",
		TotalGB:     1000.0,
		UsedGB:      600.0,
		FreeGB:      400.0,
		UsedPercent: 60.0,
	}

	formatter := NewConsoleFormatter()
	result := formatter.FormatStorage(storage)

	expected := []string{"/home", "600.00 GB", "1000.00 GB", "60.0%"}
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("저장장치 포맷 결과에 '%s'가 포함되어 있지 않습니다. 실제: %s", exp, result)
		}
	}
}

// TestConsoleFormatter_FormatGPU 테스트는 GPU 정보 포맷팅을 검증합니다
func TestConsoleFormatter_FormatGPU(t *testing.T) {
	tests := []struct {
		name     string
		gpu      *model.GPUInfo
		expected []string
	}{
		{
			name: "메모리 정보가 있는 GPU",
			gpu: &model.GPUInfo{
				Name:     "NVIDIA RTX 4090",
				Vendor:   "NVIDIA",
				MemoryGB: 24.0,
				Driver:   "525.60.11",
			},
			expected: []string{"NVIDIA RTX 4090", "NVIDIA", "24.00 GB"},
		},
		{
			name: "메모리 정보가 없는 GPU",
			gpu: &model.GPUInfo{
				Name:     "Intel UHD Graphics 630",
				Vendor:   "Intel",
				MemoryGB: 0,
				Driver:   "N/A",
			},
			expected: []string{"Intel UHD Graphics 630", "Intel"},
		},
	}

	formatter := NewConsoleFormatter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatGPU(tt.gpu)
			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("GPU 포맷 결과에 '%s'가 포함되어 있지 않습니다. 실제: %s", exp, result)
				}
			}
		})
	}
}

// TestConsoleFormatter_FormatEmptyGPU 테스트는 GPU가 없는 경우를 검증합니다
func TestConsoleFormatter_FormatEmptyGPU(t *testing.T) {
	testInfo := &model.SystemInfo{
		CPU: model.CPUInfo{
			Model:      "Test CPU",
			Cores:      4,
			Threads:    8,
			MaxFreqMHz: 3000,
		},
		Memory: model.MemoryInfo{
			TotalGB:     8.0,
			UsedGB:      4.0,
			AvailableGB: 4.0,
			UsedPercent: 50.0,
		},
		Storage: []model.StorageInfo{},
		GPU:     []model.GPUInfo{}, // 빈 GPU 슬라이스
	}

	formatter := NewConsoleFormatter()
	result := formatter.Format(testInfo)

	if !strings.Contains(result, "GPU 정보를 찾을 수 없습니다") {
		t.Error("GPU가 없을 때 적절한 메시지가 표시되지 않습니다")
	}
}
