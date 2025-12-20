//go:build linux

package collector

import (
	"testing"

	"wkqcosoft.com/m/collector/linux"
	"wkqcosoft.com/m/model"
)

// TestLinuxCollector_CollectCPU 테스트는 CPU 정보 수집을 검증합니다
func TestLinuxCollector_CollectCPU(t *testing.T) {
	collector := linux.New()

	cpu, err := collector.CollectCPU()
	if err != nil {
		t.Fatalf("CPU 정보 수집 실패: %v", err)
	}

	// CPU 정보 유효성 검증
	if cpu.Model == "" {
		t.Error("CPU 모델명이 비어있습니다")
	}

	if cpu.Cores <= 0 {
		t.Errorf("CPU 코어 수가 유효하지 않습니다: %d", cpu.Cores)
	}

	if cpu.Threads <= 0 {
		t.Errorf("CPU 스레드 수가 유효하지 않습니다: %d", cpu.Threads)
	}

	// 논리 코어는 물리 코어보다 크거나 같아야 함
	if cpu.Threads < cpu.Cores {
		t.Errorf("스레드 수(%d)가 코어 수(%d)보다 작습니다", cpu.Threads, cpu.Cores)
	}

	t.Logf("CPU 정보 수집 성공: %s (%d코어/%d스레드, %d MHz)",
		cpu.Model, cpu.Cores, cpu.Threads, cpu.MaxFreqMHz)
}

// TestLinuxCollector_CollectMemory 테스트는 메모리 정보 수집을 검증합니다
func TestLinuxCollector_CollectMemory(t *testing.T) {
	collector := linux.New()

	memory, err := collector.CollectMemory()
	if err != nil {
		t.Fatalf("메모리 정보 수집 실패: %v", err)
	}

	// 메모리 정보 유효성 검증
	if memory.TotalGB <= 0 {
		t.Errorf("전체 메모리가 유효하지 않습니다: %.2f GB", memory.TotalGB)
	}

	if memory.UsedGB < 0 {
		t.Errorf("사용 중인 메모리가 음수입니다: %.2f GB", memory.UsedGB)
	}

	if memory.AvailableGB < 0 {
		t.Errorf("사용 가능한 메모리가 음수입니다: %.2f GB", memory.AvailableGB)
	}

	if memory.UsedPercent < 0 || memory.UsedPercent > 100 {
		t.Errorf("메모리 사용률이 범위를 벗어났습니다: %.2f%%", memory.UsedPercent)
	}

	// 사용 중인 메모리 + 사용 가능한 메모리 ≈ 전체 메모리
	totalCalc := memory.UsedGB + memory.AvailableGB
	if totalCalc < memory.TotalGB*0.9 || totalCalc > memory.TotalGB*1.1 {
		t.Errorf("메모리 계산이 일치하지 않습니다: 전체=%.2f, 사용+가용=%.2f",
			memory.TotalGB, totalCalc)
	}

	t.Logf("메모리 정보 수집 성공: %.2f GB / %.2f GB (%.1f%% 사용 중)",
		memory.UsedGB, memory.TotalGB, memory.UsedPercent)
}

// TestLinuxCollector_CollectStorage 테스트는 저장장치 정보 수집을 검증합니다
func TestLinuxCollector_CollectStorage(t *testing.T) {
	collector := linux.New()

	storages, err := collector.CollectStorage()
	if err != nil {
		t.Fatalf("저장장치 정보 수집 실패: %v", err)
	}

	// 최소한 하나의 저장장치는 있어야 함
	if len(storages) == 0 {
		t.Fatal("저장장치가 하나도 발견되지 않았습니다")
	}

	// 각 저장장치 정보 유효성 검증
	for i, storage := range storages {
		if storage.Device == "" {
			t.Errorf("저장장치 #%d: 장치명이 비어있습니다", i)
		}

		if storage.MountPoint == "" {
			t.Errorf("저장장치 #%d: 마운트 지점이 비어있습니다", i)
		}

		if storage.TotalGB <= 0 {
			t.Errorf("저장장치 #%d: 전체 용량이 유효하지 않습니다: %.2f GB", i, storage.TotalGB)
		}

		if storage.UsedGB < 0 {
			t.Errorf("저장장치 #%d: 사용 중인 용량이 음수입니다: %.2f GB", i, storage.UsedGB)
		}

		if storage.FreeGB < 0 {
			t.Errorf("저장장치 #%d: 남은 용량이 음수입니다: %.2f GB", i, storage.FreeGB)
		}

		if storage.UsedPercent < 0 || storage.UsedPercent > 100 {
			t.Errorf("저장장치 #%d: 사용률이 범위를 벗어났습니다: %.2f%%", i, storage.UsedPercent)
		}

		t.Logf("저장장치 #%d: %s [%s] %.2f GB / %.2f GB (%.1f%%)",
			i, storage.Device, storage.MountPoint, storage.UsedGB, storage.TotalGB, storage.UsedPercent)
	}
}

// TestLinuxCollector_CollectGPU 테스트는 GPU 정보 수집을 검증합니다
func TestLinuxCollector_CollectGPU(t *testing.T) {
	collector := linux.New()

	gpus, err := collector.CollectGPU()

	// GPU는 없을 수도 있으므로 에러가 발생해도 테스트를 계속 진행
	if err != nil {
		t.Logf("GPU 정보를 찾을 수 없습니다 (예상 가능한 상황): %v", err)
		return
	}

	// GPU가 발견된 경우 유효성 검증
	if len(gpus) == 0 {
		t.Log("GPU가 발견되지 않았습니다")
		return
	}

	for i, gpu := range gpus {
		if gpu.Name == "" {
			t.Errorf("GPU #%d: 이름이 비어있습니다", i)
		}

		if gpu.Vendor == "" {
			t.Errorf("GPU #%d: 제조사가 비어있습니다", i)
		}

		// 메모리는 0일 수도 있음 (정보를 가져오지 못한 경우)
		if gpu.MemoryGB < 0 {
			t.Errorf("GPU #%d: 메모리가 음수입니다: %.2f GB", i, gpu.MemoryGB)
		}

		t.Logf("GPU #%d: %s [%s] %.2f GB", i, gpu.Name, gpu.Vendor, gpu.MemoryGB)
	}
}

// TestLinuxCollector_CollectAll 테스트는 전체 시스템 정보 수집을 검증합니다
func TestLinuxCollector_CollectAll(t *testing.T) {
	collector := linux.New()

	sysInfo, err := collector.CollectAll()
	if err != nil {
		t.Fatalf("시스템 정보 수집 실패: %v", err)
	}

	// 모든 주요 컴포넌트가 수집되었는지 확인
	if sysInfo.CPU.Model == "" {
		t.Error("CPU 정보가 수집되지 않았습니다")
	}

	if sysInfo.Memory.TotalGB <= 0 {
		t.Error("메모리 정보가 수집되지 않았습니다")
	}

	if len(sysInfo.Storage) == 0 {
		t.Error("저장장치 정보가 수집되지 않았습니다")
	}

	// GPU는 선택적이므로 경고만 표시
	if len(sysInfo.GPU) == 0 {
		t.Log("경고: GPU 정보가 수집되지 않았습니다 (GPU가 없거나 정보를 가져올 수 없음)")
	}

	t.Log("전체 시스템 정보 수집 성공")
}

// TestMemoryInfoConsistency 테스트는 메모리 정보의 일관성을 검증합니다
func TestMemoryInfoConsistency(t *testing.T) {
	memory := &model.MemoryInfo{
		TotalGB:     16.0,
		UsedGB:      8.0,
		AvailableGB: 8.0,
		UsedPercent: 50.0,
	}

	// 사용률 계산 검증
	expectedPercent := (memory.UsedGB / memory.TotalGB) * 100
	if memory.UsedPercent != expectedPercent {
		t.Errorf("메모리 사용률이 일치하지 않습니다: 기대값=%.2f%%, 실제값=%.2f%%",
			expectedPercent, memory.UsedPercent)
	}

	// 전체 용량 검증
	total := memory.UsedGB + memory.AvailableGB
	if total != memory.TotalGB {
		t.Errorf("메모리 전체 용량이 일치하지 않습니다: 기대값=%.2f GB, 실제값=%.2f GB",
			memory.TotalGB, total)
	}
}

// TestStorageInfoConsistency 테스트는 저장장치 정보의 일관성을 검증합니다
func TestStorageInfoConsistency(t *testing.T) {
	storage := &model.StorageInfo{
		Device:      "/dev/sda1",
		MountPoint:  "/",
		Type:        "ext4",
		TotalGB:     500.0,
		UsedGB:      300.0,
		FreeGB:      200.0,
		UsedPercent: 60.0,
	}

	// 사용률 계산 검증
	expectedPercent := (storage.UsedGB / storage.TotalGB) * 100
	if storage.UsedPercent != expectedPercent {
		t.Errorf("저장장치 사용률이 일치하지 않습니다: 기대값=%.2f%%, 실제값=%.2f%%",
			expectedPercent, storage.UsedPercent)
	}

	// 전체 용량 검증 (약간의 오차 허용)
	total := storage.UsedGB + storage.FreeGB
	diff := total - storage.TotalGB
	if diff < -1.0 || diff > 1.0 {
		t.Errorf("저장장치 전체 용량이 일치하지 않습니다: 기대값=%.2f GB, 실제값=%.2f GB (차이: %.2f GB)",
			storage.TotalGB, total, diff)
	}
}
