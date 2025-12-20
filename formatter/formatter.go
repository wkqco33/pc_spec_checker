package formatter

import (
	"fmt"
	"strings"

	"wkqcosoft.com/m/model"
)

// Formatter는 시스템 정보를 포맷팅하는 인터페이스입니다
type Formatter interface {
	Format(*model.SystemInfo) string
}

// ConsoleFormatter는 콘솔 출력용 포매터입니다
type ConsoleFormatter struct{}

// NewConsoleFormatter는 새로운 ConsoleFormatter를 생성합니다
func NewConsoleFormatter() *ConsoleFormatter {
	return &ConsoleFormatter{}
}

// Format은 시스템 정보를 보기 좋은 형태로 포맷팅합니다
func (f *ConsoleFormatter) Format(info *model.SystemInfo) string {
	var builder strings.Builder

	// 제목
	builder.WriteString("\n")
	builder.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	builder.WriteString("║              PC 사양 정보 (System Specifications)              ║\n")
	builder.WriteString("╚════════════════════════════════════════════════════════════════╝\n")
	builder.WriteString("\n")

	// CPU 정보
	builder.WriteString("┌─ CPU 정보\n")
	builder.WriteString(fmt.Sprintf("│  모델명: %s\n", info.CPU.Model))
	builder.WriteString(fmt.Sprintf("│  물리 코어: %d개\n", info.CPU.Cores))
	builder.WriteString(fmt.Sprintf("│  논리 코어(스레드): %d개\n", info.CPU.Threads))
	builder.WriteString(fmt.Sprintf("│  최대 클럭: %d MHz\n", info.CPU.MaxFreqMHz))
	builder.WriteString("└─────────────────────────────────────────────────────────────────\n")
	builder.WriteString("\n")

	// 메모리 정보
	builder.WriteString("┌─ 메모리 (RAM) 정보\n")
	builder.WriteString(fmt.Sprintf("│  전체 용량: %.2f GB\n", info.Memory.TotalGB))
	builder.WriteString(fmt.Sprintf("│  사용 중: %.2f GB (%.1f%%)\n", info.Memory.UsedGB, info.Memory.UsedPercent))
	builder.WriteString(fmt.Sprintf("│  사용 가능: %.2f GB\n", info.Memory.AvailableGB))
	builder.WriteString("└─────────────────────────────────────────────────────────────────\n")
	builder.WriteString("\n")

	// 저장장치 정보
	builder.WriteString("┌─ 저장장치 정보\n")
	for i, storage := range info.Storage {
		if i > 0 {
			builder.WriteString("│  ────────────────────────────────────────────────────────────\n")
		}
		builder.WriteString(fmt.Sprintf("│  장치: %s\n", storage.Device))
		builder.WriteString(fmt.Sprintf("│  마운트 지점: %s\n", storage.MountPoint))
		builder.WriteString(fmt.Sprintf("│  파일시스템: %s\n", storage.Type))
		builder.WriteString(fmt.Sprintf("│  전체 용량: %.2f GB\n", storage.TotalGB))
		builder.WriteString(fmt.Sprintf("│  사용 중: %.2f GB (%.1f%%)\n", storage.UsedGB, storage.UsedPercent))
		builder.WriteString(fmt.Sprintf("│  남은 용량: %.2f GB\n", storage.FreeGB))
	}
	builder.WriteString("└─────────────────────────────────────────────────────────────────\n")
	builder.WriteString("\n")

	// GPU 정보
	builder.WriteString("┌─ GPU 정보\n")
	if len(info.GPU) == 0 {
		builder.WriteString("│  GPU 정보를 찾을 수 없습니다.\n")
	} else {
		for i, gpu := range info.GPU {
			if i > 0 {
				builder.WriteString("│  ────────────────────────────────────────────────────────────\n")
			}
			builder.WriteString(fmt.Sprintf("│  이름: %s\n", gpu.Name))
			builder.WriteString(fmt.Sprintf("│  제조사: %s\n", gpu.Vendor))
			if gpu.MemoryGB > 0 {
				builder.WriteString(fmt.Sprintf("│  메모리: %.2f GB\n", gpu.MemoryGB))
			}
			if gpu.Driver != "" && gpu.Driver != "N/A" {
				builder.WriteString(fmt.Sprintf("│  드라이버: %s\n", gpu.Driver))
			}
		}
	}
	builder.WriteString("└─────────────────────────────────────────────────────────────────\n")
	builder.WriteString("\n")

	return builder.String()
}

// FormatCPU는 CPU 정보만 포맷팅합니다
func (f *ConsoleFormatter) FormatCPU(cpu *model.CPUInfo) string {
	return fmt.Sprintf("CPU: %s (%d코어/%d스레드, %d MHz)",
		cpu.Model, cpu.Cores, cpu.Threads, cpu.MaxFreqMHz)
}

// FormatMemory는 메모리 정보만 포맷팅합니다
func (f *ConsoleFormatter) FormatMemory(mem *model.MemoryInfo) string {
	return fmt.Sprintf("메모리: %.2f GB / %.2f GB (%.1f%% 사용 중)",
		mem.UsedGB, mem.TotalGB, mem.UsedPercent)
}

// FormatStorage는 저장장치 정보만 포맷팅합니다
func (f *ConsoleFormatter) FormatStorage(storage *model.StorageInfo) string {
	return fmt.Sprintf("저장장치 [%s]: %.2f GB / %.2f GB (%.1f%% 사용 중)",
		storage.MountPoint, storage.UsedGB, storage.TotalGB, storage.UsedPercent)
}

// FormatGPU는 GPU 정보만 포맷팅합니다
func (f *ConsoleFormatter) FormatGPU(gpu *model.GPUInfo) string {
	if gpu.MemoryGB > 0 {
		return fmt.Sprintf("GPU: %s [%s] (%.2f GB)",
			gpu.Name, gpu.Vendor, gpu.MemoryGB)
	}
	return fmt.Sprintf("GPU: %s [%s]", gpu.Name, gpu.Vendor)
}
