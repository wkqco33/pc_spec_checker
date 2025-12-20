package collector

import (
	"fmt"
	"runtime"

	"wkqcosoft.com/m/collector/darwin"
	"wkqcosoft.com/m/collector/linux"
	"wkqcosoft.com/m/collector/windows"
	"wkqcosoft.com/m/model"
)

// SystemCollector는 시스템 정보 수집을 담당하는 인터페이스입니다
type SystemCollector interface {
	CollectAll() (*model.SystemInfo, error)
	CollectCPU() (*model.CPUInfo, error)
	CollectMemory() (*model.MemoryInfo, error)
	CollectStorage() ([]model.StorageInfo, error)
	CollectGPU() ([]model.GPUInfo, error)
}

// NewCollector는 현재 OS에 맞는 SystemCollector를 생성합니다
func NewCollector() (SystemCollector, error) {
	switch runtime.GOOS {
	case "linux":
		return linux.New(), nil
	case "darwin":
		return darwin.New(), nil
	case "windows":
		return windows.New(), nil
	default:
		return nil, fmt.Errorf("지원하지 않는 운영체제입니다: %s", runtime.GOOS)
	}
}
