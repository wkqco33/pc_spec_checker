package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"wkqcosoft.com/m/collector"
	"wkqcosoft.com/m/formatter"
)

var (
	// Version은 빌드 시 LDFLAGS를 통해 주입됩니다.
	Version = "dev"
	// BuildTime은 빌드 시 LDFLAGS를 통해 주입됩니다.
	BuildTime = "unknown"
)

func main() {
	versionFlag := flag.Bool("v", false, "버전 정보를 표시합니다")
	longVersionFlag := flag.Bool("version", false, "버전 정보를 표시합니다")
	helpFlag := flag.Bool("h", false, "도움말을 표시합니다")
	longHelpFlag := flag.Bool("help", false, "도움말을 표시합니다")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "사용법: pc_spec_checker [옵션]\n\n")
		fmt.Fprintf(os.Stderr, "시스템 하드웨어 사양 정보를 수집하고 표시하는 CLI 도구입니다.\n\n")
		fmt.Fprintf(os.Stderr, "옵션:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *versionFlag || *longVersionFlag {
		fmt.Printf("pc_spec_checker 버전: %s\n", Version)
		fmt.Printf("빌드 시간: %s\n", BuildTime)
		fmt.Printf("Go 버전: %s\n", runtime.Version())
		return
	}

	if *helpFlag || *longHelpFlag {
		flag.Usage()
		return
	}

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
