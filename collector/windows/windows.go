//go:build windows

package windows

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"wkqcosoft.com/m/model"
)

// Collector는 Windows 시스템의 정보를 수집하는 구조체입니다
type Collector struct{}

// New는 새로운 Collector를 생성합니다
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

	go func() {
		defer wg.Done()
		cpu, errs[0] = c.CollectCPU()
	}()
	go func() {
		defer wg.Done()
		memory, errs[1] = c.CollectMemory()
	}()
	go func() {
		defer wg.Done()
		storage, errs[2] = c.CollectStorage()
	}()
	go func() {
		defer wg.Done()
		var err error
		gpu, err = c.CollectGPU()
		if err != nil {
			gpu = []model.GPUInfo{}
		}
	}()

	wg.Wait()

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

// CollectCPU는 CPU 정보를 수집합니다 (wmic → PowerShell CIM 폴백)
func (c *Collector) CollectCPU() (*model.CPUInfo, error) {
	raw := &cpuRaw{}

	// CPU 이름
	if v := c.wmicValue("cpu get Name /value", "Name"); v != "" {
		raw.Model = v
	}
	if raw.Model == "" {
		raw.Model = c.psValue("Win32_Processor", "Name")
	}

	// 물리 코어 수
	if v := c.wmicValue("cpu get NumberOfCores /value", "NumberOfCores"); v != "" {
		raw.Cores = parseInt(v)
	}
	if raw.Cores == 0 {
		raw.Cores = parseInt(c.psValue("Win32_Processor", "NumberOfCores"))
	}

	// 논리 프로세서 수 (스레드)
	if v := c.wmicValue("cpu get NumberOfLogicalProcessors /value", "NumberOfLogicalProcessors"); v != "" {
		raw.Threads = parseInt(v)
	}
	if raw.Threads == 0 {
		raw.Threads = parseInt(c.psValue("Win32_Processor", "NumberOfLogicalProcessors"))
	}

	// CPU 최대 클럭 속도 (MHz)
	if v := c.wmicValue("cpu get MaxClockSpeed /value", "MaxClockSpeed"); v != "" {
		raw.MaxFreq = parseInt(v)
	}
	if raw.MaxFreq == 0 {
		raw.MaxFreq = parseInt(c.psValue("Win32_Processor", "MaxClockSpeed"))
	}

	if !raw.valid() {
		return nil, fmt.Errorf("CPU 정보를 가져올 수 없습니다")
	}

	return &model.CPUInfo{
		Model:      raw.Model,
		Cores:      raw.Cores,
		Threads:    raw.Threads,
		MaxFreqMHz: raw.MaxFreq,
	}, nil
}

// wmicValue는 wmic 명령 출력에서 키 값을 추출합니다
func (c *Collector) wmicValue(args, key string) string {
	cmd := exec.Command("wmic", strings.Fields(args)...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return parseValueLine(string(output), key)
}

// psValue는 PowerShell CIM으로 값을 조회합니다 (wmic이 없는 최신 Windows 대응)
func (c *Collector) psValue(class, property string) string {
	script := "(Get-CimInstance -ClassName " + class + " | Select-Object -First 1)." + property
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// CollectMemory는 메모리 정보를 수집합니다 (wmic → PowerShell CIM 폴백)
func (c *Collector) CollectMemory() (*model.MemoryInfo, error) {
	memInfo := &model.MemoryInfo{}

	// 전체 물리 메모리
	output, err := exec.Command("wmic", "ComputerSystem", "get", "TotalPhysicalMemory", "/value").Output()
	if err != nil {
		output, err = c.psOutput("(Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory")
	}
	if err != nil {
		return nil, err
	}

	if v := parseValueLine(string(output), "TotalPhysicalMemory"); v != "" {
		if totalBytes, err := strconv.ParseUint(v, 10, 64); err == nil {
			memInfo.TotalGB = float64(totalBytes) / 1024 / 1024 / 1024
		}
	} else if v := strings.TrimSpace(string(output)); v != "" {
		if totalBytes, err := strconv.ParseUint(v, 10, 64); err == nil {
			memInfo.TotalGB = float64(totalBytes) / 1024 / 1024 / 1024
		}
	}

	// 사용 가능한 메모리
	output, err = exec.Command("wmic", "OS", "get", "FreePhysicalMemory", "/value").Output()
	if err != nil {
		output, err = c.psOutput("[math]::Round((Get-CimInstance Win32_OperatingSystem).FreePhysicalMemory)")
	}
	if err != nil {
		return nil, err
	}

	if v := parseValueLine(string(output), "FreePhysicalMemory"); v != "" {
		if freeKB, err := strconv.ParseUint(v, 10, 64); err == nil {
			memInfo.AvailableGB = float64(freeKB) / 1024 / 1024
		}
	} else if v := strings.TrimSpace(string(output)); v != "" {
		if freeKB, err := strconv.ParseUint(v, 10, 64); err == nil {
			memInfo.AvailableGB = float64(freeKB) / 1024 / 1024
		}
	}

	memInfo.UsedGB = memInfo.TotalGB - memInfo.AvailableGB
	if memInfo.TotalGB > 0 {
		memInfo.UsedPercent = (memInfo.UsedGB / memInfo.TotalGB) * 100
	}

	return memInfo, nil
}

// psOutput은 PowerShell 명령을 실행해 출력을 반환합니다
func (c *Collector) psOutput(script string) ([]byte, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	return cmd.Output()
}

// CollectStorage는 저장장치 정보를 수집합니다 (wmic → PowerShell CIM 폴백)
func (c *Collector) CollectStorage() ([]model.StorageInfo, error) {
	// 논리 디스크 정보 가져오기
	output, err := exec.Command("wmic", "logicaldisk", "get", "DeviceID,FileSystem,Size,FreeSpace", "/format:csv").Output()
	if err != nil {
		script := "Get-CimInstance Win32_LogicalDisk | Where-Object {$_.Size} | " +
			"ForEach-Object { \"{0},{1},{2},{3}\" -f $_.DeviceID, $_.FileSystem, $_.Size, $_.FreeSpace }"
		output, err = c.psOutput(script)
		if err != nil {
			return nil, err
		}
		// PowerShell 출력은 헤더 없는 CSV — 파서가 헤더를 건너뛰므로 더미 헤더 추가
		output = append([]byte("DeviceID,FileSystem,Size,FreeSpace\n"), output...)
	}

	var storages []model.StorageInfo
	lines := strings.Split(string(output), "\n")

	// CSV 헤더를 건너뛰고 파싱
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 더미 헤더 행 건너뛰기 (PowerShell 폴백 시 추가됨)
		if strings.HasPrefix(line, "DeviceID,") {
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
	output, err := exec.Command("wmic", "path", "win32_VideoController", "get", "Name,AdapterRAM,DriverVersion", "/format:csv").Output()
	if err != nil {
		script := "Get-CimInstance Win32_VideoController | " +
			"ForEach-Object { \"{0},{1},{2}\" -f $_.AdapterRAM, $_.DriverVersion, $_.Name }"
		output, err = c.psOutput(script)
		if err != nil {
			return nil, err
		}
		output = append([]byte("Node,AdapterRAM,DriverVersion,Name\n"), output...)
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
