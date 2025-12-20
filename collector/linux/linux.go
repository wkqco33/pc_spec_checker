//go:build linux

package linux

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"wkqcosoft.com/m/model"
)

// Collector는 Linux 시스템의 정보를 수집하는 구조체입니다
type Collector struct{}

// New는 새로운 Linux Collector를 생성합니다
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
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cpuInfo := &model.CPUInfo{}
	scanner := bufio.NewScanner(file)

	physicalIDs := make(map[string]bool)
	threadCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")

		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "model name":
			if cpuInfo.Model == "" {
				cpuInfo.Model = value
			}
		case "physical id":
			physicalIDs[value] = true
		case "processor":
			threadCount++
		case "cpu MHz":
			if freq, err := strconv.ParseFloat(value, 64); err == nil {
				mhz := int(freq)
				if mhz > cpuInfo.MaxFreqMHz {
					cpuInfo.MaxFreqMHz = mhz
				}
			}
		}
	}

	// 물리 코어 수 계산 (physical id의 개수)
	cpuInfo.Cores = len(physicalIDs)
	if cpuInfo.Cores == 0 {
		cpuInfo.Cores = threadCount // fallback
	}
	cpuInfo.Threads = threadCount

	// 최대 주파수는 cpufreq에서도 확인
	if cpuInfo.MaxFreqMHz == 0 {
		cpuInfo.MaxFreqMHz = getMaxFreqFromSys()
	}

	return cpuInfo, nil
}

// getMaxFreqFromSys는 sysfs에서 최대 주파수를 읽어옵니다
func getMaxFreqFromSys() int {
	data, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq")
	if err != nil {
		return 0
	}

	freq, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}

	return freq / 1000 // kHz를 MHz로 변환
}

// CollectMemory는 메모리 정보를 수집합니다
func (c *Collector) CollectMemory() (*model.MemoryInfo, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	memInfo := &model.MemoryInfo{}
	scanner := bufio.NewScanner(file)

	var memTotal, memAvailable, memFree, buffers, cached uint64

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSuffix(parts[0], ":")
		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}

		switch key {
		case "MemTotal":
			memTotal = value
		case "MemAvailable":
			memAvailable = value
		case "MemFree":
			memFree = value
		case "Buffers":
			buffers = value
		case "Cached":
			cached = value
		}
	}

	// KB를 GB로 변환
	memInfo.TotalGB = float64(memTotal) / 1024 / 1024

	// MemAvailable이 있으면 사용, 없으면 계산
	if memAvailable > 0 {
		memInfo.AvailableGB = float64(memAvailable) / 1024 / 1024
	} else {
		memInfo.AvailableGB = float64(memFree+buffers+cached) / 1024 / 1024
	}

	memInfo.UsedGB = memInfo.TotalGB - memInfo.AvailableGB
	memInfo.UsedPercent = (memInfo.UsedGB / memInfo.TotalGB) * 100

	return memInfo, nil
}

// CollectStorage는 저장장치 정보를 수집합니다
func (c *Collector) CollectStorage() ([]model.StorageInfo, error) {
	// df 명령어를 사용하여 마운트된 파일시스템 정보를 가져옵니다
	cmd := exec.Command("df", "-BG", "-T")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var storages []model.StorageInfo
	lines := strings.Split(string(output), "\n")

	// 첫 줄(헤더)은 건너뜁니다
	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < 7 {
			continue
		}

		device := fields[0]
		fsType := fields[1]
		mountPoint := fields[6]

		// tmpfs, devtmpfs 등 가상 파일시스템은 제외
		if strings.HasPrefix(device, "/dev/loop") ||
			fsType == "tmpfs" ||
			fsType == "devtmpfs" ||
			fsType == "squashfs" {
			continue
		}

		// 용량 파싱 (G 제거)
		totalStr := strings.TrimSuffix(fields[2], "G")
		usedStr := strings.TrimSuffix(fields[3], "G")
		availStr := strings.TrimSuffix(fields[4], "G")

		total, _ := strconv.ParseFloat(totalStr, 64)
		used, _ := strconv.ParseFloat(usedStr, 64)
		avail, _ := strconv.ParseFloat(availStr, 64)

		usedPercent := 0.0
		if total > 0 {
			usedPercent = (used / total) * 100
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
	var gpus []model.GPUInfo

	// NVIDIA GPU 확인 (nvidia-smi 사용)
	nvidiaGPUs, err := collectNvidiaGPU()
	if err == nil {
		gpus = append(gpus, nvidiaGPUs...)
	}

	// AMD/Intel GPU는 lspci로 확인
	otherGPUs, err := collectGPUFromLspci()
	if err == nil {
		gpus = append(gpus, otherGPUs...)
	}

	if len(gpus) == 0 {
		return nil, fmt.Errorf("GPU 정보를 찾을 수 없습니다")
	}

	return gpus, nil
}

// collectNvidiaGPU는 NVIDIA GPU 정보를 수집합니다
func collectNvidiaGPU() ([]model.GPUInfo, error) {
	// nvidia-smi가 설치되어 있는지 확인
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
		memParts := strings.Fields(memStr)
		memGB := 0.0
		if len(memParts) > 0 {
			memMiB, _ := strconv.ParseFloat(memParts[0], 64)
			memGB = memMiB / 1024
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

// collectGPUFromLspci는 lspci를 사용하여 GPU 정보를 수집합니다
func collectGPUFromLspci() ([]model.GPUInfo, error) {
	cmd := exec.Command("lspci")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var gpus []model.GPUInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		// VGA 또는 3D controller 찾기
		if !strings.Contains(line, "VGA") && !strings.Contains(line, "3D controller") {
			continue
		}

		// GPU 이름 추출
		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}

		name := strings.TrimSpace(parts[2])
		vendor := "Unknown"

		// 제조사 파악
		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, "nvidia") {
			vendor = "NVIDIA"
		} else if strings.Contains(lowerName, "amd") || strings.Contains(lowerName, "ati") {
			vendor = "AMD"
		} else if strings.Contains(lowerName, "intel") {
			vendor = "Intel"
		}

		// lspci로는 메모리 정보를 정확히 얻기 어려우므로 sysfs 확인
		pciAddr := strings.TrimSpace(parts[0])
		memGB := getGPUMemoryFromSys(pciAddr)

		gpus = append(gpus, model.GPUInfo{
			Name:     name,
			Vendor:   vendor,
			MemoryGB: memGB,
			Driver:   "N/A",
		})
	}

	return gpus, nil
}

// getGPUMemoryFromSys는 sysfs에서 GPU 메모리를 읽어옵니다
func getGPUMemoryFromSys(pciAddr string) float64 {
	// /sys/bus/pci/devices/0000:XX:XX.X/resource 파일에서 메모리 정보 확인
	resourcePath := filepath.Join("/sys/bus/pci/devices", "0000:"+pciAddr, "resource")

	data, err := os.ReadFile(resourcePath)
	if err != nil {
		return 0
	}

	// resource 파일의 각 라인은 메모리 영역을 나타냅니다
	lines := strings.Split(string(data), "\n")
	totalMem := uint64(0)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// 시작 주소와 끝 주소 파싱
		start, err1 := strconv.ParseUint(fields[0], 0, 64)
		end, err2 := strconv.ParseUint(fields[1], 0, 64)

		if err1 == nil && err2 == nil && end > start {
			totalMem += (end - start)
		}
	}

	// 바이트를 GB로 변환
	return float64(totalMem) / 1024 / 1024 / 1024
}
