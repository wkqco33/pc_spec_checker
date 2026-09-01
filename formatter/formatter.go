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

// Format은 시스템 정보의 핵심 항목을 요약해 포맷팅합니다
func (f *ConsoleFormatter) Format(info *model.SystemInfo) string {
	var builder strings.Builder

	builder.WriteString("\nPC 사양 요약\n")
	builder.WriteString("────────────────────────────────────────\n")
	fmt.Fprintf(&builder, "CPU: %s (%d코어/%d스레드)\n", info.CPU.Model, info.CPU.Cores, info.CPU.Threads)
	fmt.Fprintf(&builder, "메모리: %.2f GB (%.1f%% 사용 중)\n", info.Memory.TotalGB, info.Memory.UsedPercent)

	if len(info.Storage) == 0 {
		builder.WriteString("저장장치: 정보를 찾을 수 없습니다.\n")
	} else {
		for _, storage := range info.Storage {
			fmt.Fprintf(&builder, "저장장치: %.2f GB (%.1f%% 사용 중)\n", storage.TotalGB, storage.UsedPercent)
		}
	}

	if len(info.GPU) == 0 {
		builder.WriteString("GPU: 정보를 찾을 수 없습니다.\n")
	} else {
		for _, gpu := range info.GPU {
			if gpu.MemoryGB > 0 {
				fmt.Fprintf(&builder, "GPU: %s (%.2f GB)\n", gpu.Name, gpu.MemoryGB)
			} else {
				fmt.Fprintf(&builder, "GPU: %s\n", gpu.Name)
			}
		}
	}
	builder.WriteString("\n")

	return builder.String()
}

// FormatVerbose는 시스템 정보의 전체 항목을 상세하게 포맷팅합니다
func (f *ConsoleFormatter) FormatVerbose(info *model.SystemInfo) string {
	var builder strings.Builder

	// 제목
	builder.WriteString("\n")
	builder.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	builder.WriteString("║              PC 사양 정보 (System Specifications)              ║\n")
	builder.WriteString("╚════════════════════════════════════════════════════════════════╝\n")
	builder.WriteString("\n")

	// CPU 정보
	builder.WriteString("┌─ CPU 정보\n")
	fmt.Fprintf(&builder, "│  모델명: %s\n", info.CPU.Model)
	fmt.Fprintf(&builder, "│  물리 코어: %d개\n", info.CPU.Cores)
	fmt.Fprintf(&builder, "│  논리 코어(스레드): %d개\n", info.CPU.Threads)
	fmt.Fprintf(&builder, "│  최대 클럭: %d MHz\n", info.CPU.MaxFreqMHz)
	builder.WriteString("└─────────────────────────────────────────────────────────────────\n")
	builder.WriteString("\n")

	// 메모리 정보
	builder.WriteString("┌─ 메모리 (RAM) 정보\n")
	fmt.Fprintf(&builder, "│  전체 용량: %.2f GB\n", info.Memory.TotalGB)
	fmt.Fprintf(&builder, "│  사용 중: %.2f GB (%.1f%%)\n", info.Memory.UsedGB, info.Memory.UsedPercent)
	fmt.Fprintf(&builder, "│  사용 가능: %.2f GB\n", info.Memory.AvailableGB)
	builder.WriteString("└─────────────────────────────────────────────────────────────────\n")
	builder.WriteString("\n")

	// 저장장치 정보
	builder.WriteString("┌─ 저장장치 정보\n")
	for i, storage := range info.Storage {
		if i > 0 {
			builder.WriteString("│  ────────────────────────────────────────────────────────────\n")
		}
		fmt.Fprintf(&builder, "│  장치: %s\n", storage.Device)
		fmt.Fprintf(&builder, "│  마운트 지점: %s\n", storage.MountPoint)
		fmt.Fprintf(&builder, "│  파일시스템: %s\n", storage.Type)
		fmt.Fprintf(&builder, "│  전체 용량: %.2f GB\n", storage.TotalGB)
		fmt.Fprintf(&builder, "│  사용 중: %.2f GB (%.1f%%)\n", storage.UsedGB, storage.UsedPercent)
		fmt.Fprintf(&builder, "│  남은 용량: %.2f GB\n", storage.FreeGB)
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
			fmt.Fprintf(&builder, "│  이름: %s\n", gpu.Name)
			fmt.Fprintf(&builder, "│  제조사: %s\n", gpu.Vendor)
			if gpu.MemoryGB > 0 {
				fmt.Fprintf(&builder, "│  메모리: %.2f GB\n", gpu.MemoryGB)
			}
			if gpu.Driver != "" && gpu.Driver != "N/A" {
				fmt.Fprintf(&builder, "│  드라이버: %s\n", gpu.Driver)
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
