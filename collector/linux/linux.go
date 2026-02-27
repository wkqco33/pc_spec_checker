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
	"sync"
	"syscall"

	"wkqcosoft.com/m/model"
)

// Collector는 Linux 시스템의 정보를 수집하는 구조체입니다
type Collector struct{}

// New는 새로운 Linux Collector를 생성합니다
func New() *Collector {
	return &Collector{}
}

// CollectAll은 모든 시스템 정보를 병렬로 수집합니다
func (c *Collector) CollectAll() (*model.SystemInfo, error) {
	var (
		cpu     *model.CPUInfo
		memory  *model.MemoryInfo
		storage []model.StorageInfo
		gpu     []model.GPUInfo
		errs    = make([]error, 4)
		wg      sync.WaitGroup
	)

	wg.Add(4)

	// CPU 정보 수집
	go func() {
		defer wg.Done()
		cpu, errs[0] = c.CollectCPU()
	}()

	// 메모리 정보 수집
	go func() {
		defer wg.Done()
		memory, errs[1] = c.CollectMemory()
	}()

	// 저장장치 정보 수집
	go func() {
		defer wg.Done()
		storage, errs[2] = c.CollectStorage()
	}()

	// GPU 정보 수집 (에러 시 빈 리스트로 처리)
	go func() {
		defer wg.Done()
		var err error
		gpu, err = c.CollectGPU()
		if err != nil {
			gpu = []model.GPUInfo{}
		}
	}()

	wg.Wait()

	// 주요 정보(CPU, Memory, Storage) 수집 실패 시 에러 반환
	for i := 0; i < 3; i++ {
		if errs[i] != nil {
			return nil, errs[i]
		}
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

	// 물리 코어를 식별하기 위해 "physical id"와 "core id"의 조합을 저장
	// 예: "0:1" (0번 소켓의 1번 코어)
	coreMap := make(map[string]bool)
	threadCount := 0
	var currentPhysicalID, currentCoreID string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			// 빈 줄은 프로세서 정보의 끝을 의미할 수 있으므로 저장된 ID 조합을 맵에 추가
			if currentPhysicalID != "" && currentCoreID != "" {
				coreMap[currentPhysicalID+":"+currentCoreID] = true
			}
			continue
		}

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
			currentPhysicalID = value
		case "core id":
			currentCoreID = value
		case "processor":
			threadCount++
		}
	}
	// 마지막 프로세서 정보 처리
	if currentPhysicalID != "" && currentCoreID != "" {
		coreMap[currentPhysicalID+":"+currentCoreID] = true
	}

	cpuInfo.Cores = len(coreMap)
	// 만약 core id 정보가 없는 환경이라면 fallback으로 cpu cores 필드 탐색 시도
	if cpuInfo.Cores == 0 {
		cpuInfo.Cores = getCoresFallback()
	}
	// 여전히 0이라면 최소 1개로 설정
	if cpuInfo.Cores == 0 {
		cpuInfo.Cores = 1
	}
	
	cpuInfo.Threads = threadCount
	cpuInfo.MaxFreqMHz = getMaxFreqFromSysAll()

	return cpuInfo, nil
}

// getCoresFallback은 /proc/cpuinfo에서 "cpu cores" 필드를 직접 찾아 반환합니다.
func getCoresFallback() int {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "cpu cores") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				val, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
				return val // 이는 소켓당 코어 수이므로 소켓 수를 곱해야 정확하지만 일반적인 환경에선 유효함
			}
		}
	}
	return 0
}

// getMaxFreqFromSysAll은 모든 코어를 스캔하여 시스템 전체의 최대 주파수를 찾습니다
func getMaxFreqFromSysAll() int {
	maxFreq := 0
	files, err := filepath.Glob("/sys/devices/system/cpu/cpu*/cpufreq/cpuinfo_max_freq")
	if err != nil || len(files) == 0 {
		return 0
	}

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		freq, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		mhz := freq / 1000
		if mhz > maxFreq {
			maxFreq = mhz
		}
	}
	return maxFreq
}

// CollectMemory는 메모리 정보를 수집합니다
func (c *Collector) CollectMemory() (*model.MemoryInfo, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var total, available, free, buffers, cached uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		val, _ := strconv.ParseUint(fields[1], 10, 64)

		switch key {
		case "MemTotal":
			total = val
		case "MemAvailable":
			available = val
		case "MemFree":
			free = val
		case "Buffers":
			buffers = val
		case "Cached":
			cached = val
		}
	}

	memInfo := &model.MemoryInfo{
		TotalGB: float64(total) / 1024 / 1024,
	}
	if available > 0 {
		memInfo.AvailableGB = float64(available) / 1024 / 1024
	} else {
		memInfo.AvailableGB = float64(free+buffers+cached) / 1024 / 1024
	}
	memInfo.UsedGB = memInfo.TotalGB - memInfo.AvailableGB
	if memInfo.TotalGB > 0 {
		memInfo.UsedPercent = (memInfo.UsedGB / memInfo.TotalGB) * 100
	}

	return memInfo, nil
}

// CollectStorage는 syscall을 사용하여 정확한 저장장치 정보를 수집합니다
func (c *Collector) CollectStorage() ([]model.StorageInfo, error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var storages []model.StorageInfo
	scanner := bufio.NewScanner(file)
	seenMounts := make(map[string]bool)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]

		// 물리 디스크가 아닌 장치 제외
		if !strings.HasPrefix(device, "/dev/") || strings.HasPrefix(device, "/dev/loop") {
			continue
		}
		if seenMounts[mountPoint] {
			continue
		}

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mountPoint, &stat); err != nil {
			continue
		}

		total := float64(stat.Blocks) * float64(stat.Bsize) / 1024 / 1024 / 1024
		free := float64(stat.Bavail) * float64(stat.Bsize) / 1024 / 1024 / 1024
		used := total - free

		if total == 0 {
			continue
		}

		storages = append(storages, model.StorageInfo{
			Device:      device,
			MountPoint:  mountPoint,
			Type:        fsType,
			TotalGB:     total,
			UsedGB:      used,
			FreeGB:      free,
			UsedPercent: (used / total) * 100,
		})
		seenMounts[mountPoint] = true
	}

	return storages, nil
}

// CollectGPU는 GPU 정보를 수집합니다
func (c *Collector) CollectGPU() ([]model.GPUInfo, error) {
	var gpus []model.GPUInfo

	// NVIDIA 우선 (가장 정확한 정보 제공)
	if nvidiaGPUs, err := collectNvidiaGPU(); err == nil {
		gpus = append(gpus, nvidiaGPUs...)
	}

	// 기타 GPU (lspci 기반)
	if otherGPUs, err := collectGPUFromLspci(); err == nil {
		for _, og := range otherGPUs {
			// 중복 방지 (lspci에는 NVIDIA도 나오므로 이름으로 체크)
			isDuplicate := false
			for _, ng := range gpus {
				if strings.Contains(og.Name, ng.Name) || strings.Contains(ng.Name, og.Name) {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				gpus = append(gpus, og)
			}
		}
	}

	if len(gpus) == 0 {
		return nil, fmt.Errorf("GPU 정보를 찾을 수 없습니다")
	}

	return gpus, nil
}

// collectNvidiaGPU는 nvidia-smi를 사용하여 정확한 메모리 정보를 가져옵니다
func collectNvidiaGPU() ([]model.GPUInfo, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,memory.total,driver_version", "--format=csv,noheader")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var gpus []model.GPUInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		memParts := strings.Fields(parts[1])
		memMiB, _ := strconv.ParseFloat(memParts[0], 64)

		gpus = append(gpus, model.GPUInfo{
			Name:     strings.TrimSpace(parts[0]),
			Vendor:   "NVIDIA",
			MemoryGB: memMiB / 1024,
			Driver:   strings.TrimSpace(parts[2]),
		})
	}
	return gpus, nil
}

// collectGPUFromLspci는 lspci를 사용하여 범용 GPU 정보를 수집합니다
func collectGPUFromLspci() ([]model.GPUInfo, error) {
	cmd := exec.Command("lspci", "-D") // 도메인 포함 PCI 주소 확보
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var gpus []model.GPUInfo
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if !strings.Contains(line, "VGA") && !strings.Contains(line, "3D controller") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		pciAddr := parts[0]
		
		// 상세 이름 추출
		nameParts := strings.Split(parts[1], ": ")
		name := nameParts[len(nameParts)-1]
		vendor := "Unknown"
		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, "nvidia") {
			vendor = "NVIDIA"
		} else if strings.Contains(lowerName, "amd") || strings.Contains(lowerName, "ati") {
			vendor = "AMD"
		} else if strings.Contains(lowerName, "intel") {
			vendor = "Intel"
		}

		gpus = append(gpus, model.GPUInfo{
			Name:     name,
			Vendor:   vendor,
			MemoryGB: getGPUMemoryFromSys(pciAddr),
			Driver:   "N/A",
		})
	}
	return gpus, nil
}

// getGPUMemoryFromSys는 sysfs에서 BAR 중 가장 큰 prefetchable 영역을 VRAM으로 추정하여 반환합니다
func getGPUMemoryFromSys(pciAddr string) float64 {
	// PCI 도메인이 포함된 경로 확인
	path := filepath.Join("/sys/bus/pci/devices", pciAddr, "resource")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	var maxRegion uint64
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		start, _ := strconv.ParseUint(fields[0], 0, 64)
		end, _ := strconv.ParseUint(fields[1], 0, 64)
		flags, _ := strconv.ParseUint(fields[2], 0, 64)

		// IORESOURCE_MEM (0x200) 및 IORESOURCE_PREFETCH (0x2000) 플래그 확인
		// VRAM은 보통 prefetchable 메모리 영역 중 가장 큰 부분입니다.
		if (flags & 0x200) != 0 && (flags & 0x2000) != 0 {
			size := end - start + 1
			if size > maxRegion {
				maxRegion = size
			}
		}
	}

	return float64(maxRegion) / 1024 / 1024 / 1024
}

