//go:build darwin

package darwin

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"wkqcosoft.com/m/model"
)

// Collector는 macOS 시스템의 정보를 수집하는 구조체입니다
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

	// CPU 모델명
	cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	if output, err := cmd.Output(); err == nil {
		cpuInfo.Model = strings.TrimSpace(string(output))
	}

	// 물리 코어 수
	cmd = exec.Command("sysctl", "-n", "hw.physicalcpu")
	if output, err := cmd.Output(); err == nil {
		if cores, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			cpuInfo.Cores = cores
		}
	}

	// 논리 코어 수 (스레드)
	cmd = exec.Command("sysctl", "-n", "hw.logicalcpu")
	if output, err := cmd.Output(); err == nil {
		if threads, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			cpuInfo.Threads = threads
		}
	}

	// CPU 주파수 (Hz 단위로 나오므로 MHz로 변환)
	cmd = exec.Command("sysctl", "-n", "hw.cpufrequency")
	if output, err := cmd.Output(); err == nil {
		if freq, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64); err == nil {
			cpuInfo.MaxFreqMHz = int(freq / 1000000)
		}
	}

	// Apple Silicon의 경우 다른 방법 시도
	if cpuInfo.MaxFreqMHz == 0 {
		cmd = exec.Command("sysctl", "-n", "hw.cpufrequency_max")
		if output, err := cmd.Output(); err == nil {
			if freq, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64); err == nil {
				cpuInfo.MaxFreqMHz = int(freq / 1000000)
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

	// 전체 메모리 (바이트 단위)
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	totalBytes, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return nil, err
	}
	memInfo.TotalGB = float64(totalBytes) / 1024 / 1024 / 1024

	// vm_stat으로 메모리 사용 정보 가져오기
	cmd = exec.Command("vm_stat")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	// vm_stat 출력 파싱
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var pagesFree, pagesActive, pagesInactive, pagesWired, pagesSpeculative uint64
	var pageSize uint64 = 4096 // 기본 페이지 크기

	for scanner.Scan() {
		line := scanner.Text()

		// 페이지 크기 확인
		if strings.Contains(line, "page size of") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "of" && i+1 < len(parts) {
					if size, err := strconv.ParseUint(parts[i+1], 10, 64); err == nil {
						pageSize = size
					}
				}
			}
		}

		// 각 메모리 상태별 페이지 수 파싱
		if strings.HasPrefix(line, "Pages free:") {
			pagesFree = parseVmStatLine(line)
		} else if strings.HasPrefix(line, "Pages active:") {
			pagesActive = parseVmStatLine(line)
		} else if strings.HasPrefix(line, "Pages inactive:") {
			pagesInactive = parseVmStatLine(line)
		} else if strings.HasPrefix(line, "Pages wired down:") {
			pagesWired = parseVmStatLine(line)
		} else if strings.HasPrefix(line, "Pages speculative:") {
			pagesSpeculative = parseVmStatLine(line)
		}
	}

	// 메모리 계산
	usedPages := pagesActive + pagesWired
	freePages := pagesFree + pagesInactive + pagesSpeculative

	memInfo.UsedGB = float64(usedPages*pageSize) / 1024 / 1024 / 1024
	memInfo.AvailableGB = float64(freePages*pageSize) / 1024 / 1024 / 1024
	memInfo.UsedPercent = (memInfo.UsedGB / memInfo.TotalGB) * 100

	return memInfo, nil
}

// parseVmStatLine은 vm_stat 출력에서 숫자를 추출합니다
func parseVmStatLine(line string) uint64 {
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		// 마지막 필드에서 마침표 제거
		numStr := strings.TrimSuffix(parts[len(parts)-1], ".")
		if num, err := strconv.ParseUint(numStr, 10, 64); err == nil {
			return num
		}
	}
	return 0
}

// CollectStorage는 저장장치 정보를 수집합니다
func (c *Collector) CollectStorage() ([]model.StorageInfo, error) {
	// df 명령어 사용 (macOS에서도 동일)
	cmd := exec.Command("df", "-Hg")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var storages []model.StorageInfo
	lines := strings.Split(string(output), "\n")

	// 첫 줄(헤더)은 건너뜁니다
	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < 6 {
			continue
		}

		device := fields[0]
		mountPoint := fields[len(fields)-1]

		// 가상 파일시스템 제외
		if strings.HasPrefix(device, "map") ||
			strings.HasPrefix(device, "devfs") ||
			strings.HasPrefix(mountPoint, "/dev") ||
			strings.HasPrefix(mountPoint, "/System/Volumes") ||
			strings.HasPrefix(mountPoint, "/private/var") {
			continue
		}

		// 용량 파싱 (G 제거)
		totalStr := strings.TrimSuffix(fields[1], "G")
		usedStr := strings.TrimSuffix(fields[2], "G")
		availStr := strings.TrimSuffix(fields[3], "G")

		total, _ := strconv.ParseFloat(totalStr, 64)
		used, _ := strconv.ParseFloat(usedStr, 64)
		avail, _ := strconv.ParseFloat(availStr, 64)

		usedPercent := 0.0
		if total > 0 {
			usedPercent = (used / total) * 100
		}

		// 파일시스템 타입 가져오기
		fsType := "unknown"
		if len(fields) >= 9 {
			fsType = fields[7]
		} else {
			fsType = "apfs" // macOS 기본
		}

		storages = append(storages, model.StorageInfo{
			Device:      device,
			MountPoint:  mountPoint,
			Type:        fsType,
			TotalGB:     total,
			UsedGB:      used,
			FreeGB:      avail,
			UsedPercent: usedPercent,
		})
	}

	return storages, nil
}

// CollectGPU는 GPU 정보를 수집합니다
func (c *Collector) CollectGPU() ([]model.GPUInfo, error) {
	// system_profiler로 GPU 정보 가져오기
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var gpus []model.GPUInfo
	lines := strings.Split(string(output), "\n")

	var currentGPU *model.GPUInfo

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// GPU 이름 찾기
		if strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, "Displays") {
			if currentGPU != nil {
				gpus = append(gpus, *currentGPU)
			}
			gpuName := strings.TrimSuffix(trimmed, ":")
			currentGPU = &model.GPUInfo{
				Name:   gpuName,
				Vendor: detectVendor(gpuName),
			}
		}

		if currentGPU != nil {
			// VRAM 정보 찾기
			if strings.Contains(trimmed, "VRAM") || strings.Contains(trimmed, "vram") {
				parts := strings.Split(trimmed, ":")
				if len(parts) >= 2 {
					memStr := strings.TrimSpace(parts[1])
					// "8 GB" 형식에서 숫자 추출
					fields := strings.Fields(memStr)
					if len(fields) >= 1 {
						if mem, err := strconv.ParseFloat(fields[0], 64); err == nil {
							currentGPU.MemoryGB = mem
						}
					}
				}
			}

			// Vendor 정보
			if strings.Contains(trimmed, "Vendor:") {
				parts := strings.Split(trimmed, ":")
				if len(parts) >= 2 {
					vendor := strings.TrimSpace(parts[1])
					if vendor != "" {
						currentGPU.Vendor = vendor
					}
				}
			}
		}
	}

	// 마지막 GPU 추가
	if currentGPU != nil {
		gpus = append(gpus, *currentGPU)
	}

	if len(gpus) == 0 {
		return nil, fmt.Errorf("GPU 정보를 찾을 수 없습니다")
	}

	return gpus, nil
}

// detectVendor는 GPU 이름에서 제조사를 추측합니다
func detectVendor(name string) string {
	lowerName := strings.ToLower(name)
	if strings.Contains(lowerName, "nvidia") {
		return "NVIDIA"
	} else if strings.Contains(lowerName, "amd") || strings.Contains(lowerName, "radeon") {
		return "AMD"
	} else if strings.Contains(lowerName, "intel") {
		return "Intel"
	} else if strings.Contains(lowerName, "apple") {
		return "Apple"
	}
	return "Unknown"
}
