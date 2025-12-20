package main

import (
	"fmt"
	"os"
	"runtime"

	"wkqcosoft.com/m/collector"
	"wkqcosoft.com/m/formatter"
)

func main() {
	// 현재 OS에 맞는 시스템 정보 수집기 생성
	systemCollector, err := collector.NewCollector()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		os.Exit(1)
	}

	// OS 정보 표시
	osName := getOSName()
	fmt.Printf("시스템 정보를 수집 중... (%s)\n", osName)

	// 모든 시스템 정보 수집
	systemInfo, err := systemCollector.CollectAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: 시스템 정보를 수집할 수 없습니다: %v\n", err)
		os.Exit(1)
	}

	// 콘솔 포매터 생성
	consoleFormatter := formatter.NewConsoleFormatter()

	// 시스템 정보를 포맷팅하여 출력
	output := consoleFormatter.Format(systemInfo)
	fmt.Print(output)
}

// getOSName은 현재 운영체제의 이름을 반환합니다
func getOSName() string {
	switch runtime.GOOS {
	case "linux":
		return "Linux"
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}
