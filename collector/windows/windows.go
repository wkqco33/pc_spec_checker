//go:build windows

package windows

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"wkqcosoft.com/m/model"
)

// Collector는 Windows 시스템의 정보를 수집하는 구조체입니다
type Collector struct{}

// New는 새로운 Collector를 생성합니다
func New() *Collector {
	return &Collector{}
}

// CollectAll은 모든 시스템 정보를 수집합니다
func (c *Collector) CollectAll() (*model.SystemInfo, error) {
	cpu, err := c.CollectCPU()
	if err != nil {
		return nil, fmt.Errorf("CPU 정보 수집 실패: %w", err)
	}

	memory, err := c.CollectMemory()
	if err != nil {
		return nil, fmt.Errorf("메모리 정보 수집 실패: %w", err)
	}

	storage, err := c.CollectStorage()
	if err != nil {
		return nil, fmt.Errorf("저장장치 정보 수집 실패: %w", err)
	}

	gpu, err := c.CollectGPU()
	if err != nil {
		// GPU 정보는 선택적이므로 에러가 나도 계속 진행
		gpu = []model.GPUInfo{}
	}

	return &model.SystemInfo{
		CPU:     *cpu,
		Memory:  *memory,
		Storage: storage,
		GPU:     gpu,
	}, nil
}

// CollectCPU는 CPU 정보를 수집합니다
func (c *Collector) CollectCPU() (*model.CPUInfo, error) {
	cpuInfo := &model.CPUInfo{}

	// CPU 이름
	cmd := exec.Command("wmic", "cpu", "get", "Name", "/value")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Name=") {
				cpuInfo.Model = strings.TrimSpace(strings.TrimPrefix(line, "Name="))
				break
			}
		}
	}

	// 물리 코어 수
	cmd = exec.Command("wmic", "cpu", "get", "NumberOfCores", "/value")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "NumberOfCores=") {
				coresStr := strings.TrimSpace(strings.TrimPrefix(line, "NumberOfCores="))
				if cores, err := strconv.Atoi(coresStr); err == nil {
					cpuInfo.Cores = cores
				}
				break
			}
		}
	}

	// 논리 프로세서 수 (스레드)
	cmd = exec.Command("wmic", "cpu", "get", "NumberOfLogicalProcessors", "/value")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "NumberOfLogicalProcessors=") {
				threadsStr := strings.TrimSpace(strings.TrimPrefix(line, "NumberOfLogicalProcessors="))
				if threads, err := strconv.Atoi(threadsStr); err == nil {
					cpuInfo.Threads = threads
				}
				break
			}
		}
	}

	// CPU 최대 클럭 속도 (MHz)
	cmd = exec.Command("wmic", "cpu", "get", "MaxClockSpeed", "/value")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MaxClockSpeed=") {
				freqStr := strings.TrimSpace(strings.TrimPrefix(line, "MaxClockSpeed="))
				if freq, err := strconv.Atoi(freqStr); err == nil {
					cpuInfo.MaxFreqMHz = freq
				}
				break
			}
		}
	}

	if cpuInfo.Model == "" {
		return nil, fmt.Errorf("CPU 정보를 가져올 수 없습니다")
	}

	return cpuInfo, nil
}

// CollectMemory는 메모리 정보를 수집합니다
func (c *Collector) CollectMemory() (*model.MemoryInfo, error) {
	memInfo := &model.MemoryInfo{}

	// 전체 물리 메모리
	cmd := exec.Command("wmic", "ComputerSystem", "get", "TotalPhysicalMemory", "/value")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "TotalPhysicalMemory=") {
			memStr := strings.TrimSpace(strings.TrimPrefix(line, "TotalPhysicalMemory="))
			if totalBytes, err := strconv.ParseUint(memStr, 10, 64); err == nil {
				memInfo.TotalGB = float64(totalBytes) / 1024 / 1024 / 1024
			}
			break
		}
	}

	// 사용 가능한 메모리
	cmd = exec.Command("wmic", "OS", "get", "FreePhysicalMemory", "/value")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	lines = strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "FreePhysicalMemory=") {
			memStr := strings.TrimSpace(strings.TrimPrefix(line, "FreePhysicalMemory="))
			if freeKB, err := strconv.ParseUint(memStr, 10, 64); err == nil {
				memInfo.AvailableGB = float64(freeKB) / 1024 / 1024
			}
			break
		}
	}

	memInfo.UsedGB = memInfo.TotalGB - memInfo.AvailableGB
	if memInfo.TotalGB > 0 {
		memInfo.UsedPercent = (memInfo.UsedGB / memInfo.TotalGB) * 100
	}

	return memInfo, nil
}

// CollectStorage는 저장장치 정보를 수집합니다
func (c *Collector) CollectStorage() ([]model.StorageInfo, error) {
	// 논리 디스크 정보 가져오기
	cmd := exec.Command("wmic", "logicaldisk", "get", "DeviceID,FileSystem,Size,FreeSpace", "/format:csv")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var storages []model.StorageInfo
	lines := strings.Split(string(output), "\n")

	// CSV 헤더를 건너뛰고 파싱
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			continue
		}

		// Node,DeviceID,FileSystem,FreeSpace,Size 순서
		deviceID := strings.TrimSpace(fields[1])
		fileSystem := strings.TrimSpace(fields[2])
		freeSpaceStr := strings.TrimSpace(fields[3])
		sizeStr := strings.TrimSpace(fields[4])

		// 빈 값 건너뛰기
		if deviceID == "" || sizeStr == "" {
			continue
		}

		// 바이트를 GB로 변환
		totalBytes, err1 := strconv.ParseUint(sizeStr, 10, 64)
		freeBytes, err2 := strconv.ParseUint(freeSpaceStr, 10, 64)

		if err1 != nil || err2 != nil || totalBytes == 0 {
			continue
		}

		totalGB := float64(totalBytes) / 1024 / 1024 / 1024
		freeGB := float64(freeBytes) / 1024 / 1024 / 1024
		usedGB := totalGB - freeGB
		usedPercent := (usedGB / totalGB) * 100

		storages = append(storages, model.StorageInfo{
			Device:      deviceID,
			MountPoint:  deviceID, // Windows에서는 드라이브 레터가 마운트 포인트
			Type:        fileSystem,
			TotalGB:     totalGB,
			UsedGB:      usedGB,
			FreeGB:      freeGB,
			UsedPercent: usedPercent,
		})
	}

	if len(storages) == 0 {
		return nil, fmt.Errorf("저장장치 정보를 찾을 수 없습니다")
	}

	return storages, nil
}

// CollectGPU는 GPU 정보를 수집합니다
func (c *Collector) CollectGPU() ([]model.GPUInfo, error) {
	// NVIDIA GPU 정보 먼저 시도 (nvidia-smi가 있는 경우)
	nvidiaGPUs, err := c.collectNvidiaGPU()
	if err == nil && len(nvidiaGPUs) > 0 {
		return nvidiaGPUs, nil
	}

	// wmic로 GPU 정보 가져오기
	cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "Name,AdapterRAM,DriverVersion", "/format:csv")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var gpus []model.GPUInfo
	lines := strings.Split(string(output), "\n")

	// CSV 파싱 (헤더 건너뛰기)
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 4 {
			continue
		}

		// Node,AdapterRAM,DriverVersion,Name 순서
		ramStr := strings.TrimSpace(fields[1])
		driver := strings.TrimSpace(fields[2])
		name := strings.TrimSpace(fields[3])

		if name == "" {
			continue
		}

		// RAM을 GB로 변환
		memGB := 0.0
		if ramBytes, err := strconv.ParseUint(ramStr, 10, 64); err == nil && ramBytes > 0 {
			memGB = float64(ramBytes) / 1024 / 1024 / 1024
		}

		gpus = append(gpus, model.GPUInfo{
			Name:     name,
			Vendor:   detectVendorFromName(name),
			MemoryGB: memGB,
			Driver:   driver,
		})
	}

	if len(gpus) == 0 {
		return nil, fmt.Errorf("GPU 정보를 찾을 수 없습니다")
	}

	return gpus, nil
}

// collectNvidiaGPU는 nvidia-smi를 사용하여 NVIDIA GPU 정보를 수집합니다
func (c *Collector) collectNvidiaGPU() ([]model.GPUInfo, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,memory.total,driver_version", "--format=csv,noheader")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var gpus []model.GPUInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		memStr := strings.TrimSpace(parts[1])
		driver := strings.TrimSpace(parts[2])

		// 메모리 파싱 (예: "8192 MiB")
		memGB := 0.0
		memParts := strings.Fields(memStr)
		if len(memParts) > 0 {
			if memMiB, err := strconv.ParseFloat(memParts[0], 64); err == nil {
				memGB = memMiB / 1024
			}
		}

		gpus = append(gpus, model.GPUInfo{
			Name:     name,
			Vendor:   "NVIDIA",
			MemoryGB: memGB,
			Driver:   driver,
		})
	}

	return gpus, nil
}

// detectVendorFromName은 GPU 이름에서 제조사를 추측합니다
func detectVendorFromName(name string) string {
	lowerName := strings.ToLower(name)
	if strings.Contains(lowerName, "nvidia") || strings.Contains(lowerName, "geforce") || strings.Contains(lowerName, "quadro") {
		return "NVIDIA"
	} else if strings.Contains(lowerName, "amd") || strings.Contains(lowerName, "radeon") {
		return "AMD"
	} else if strings.Contains(lowerName, "intel") {
		return "Intel"
	}
	return "Unknown"
}
