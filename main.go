package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"

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
	root := newRootCommand()
	if err := root.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

// newRootCommand는 pcsc 루트 커맨드를 구성합니다
func newRootCommand() *wcli.Command {
	var (
		showBuildTime bool
		root          *wcli.Command
	)

	root = &wcli.Command{
		Use:     "pcsc",
		Short:   "시스템 하드웨어 사양 정보를 수집하고 표시하는 CLI 도구입니다",
		Long:    "pcsc는 Linux/macOS/Windows의 하드웨어 사양 정보를 수집해 콘솔에 표시합니다.",
		Version: Version,
		Run: func(ctx *wcli.Context) error {
			return runCollect(root, showBuildTime)
		},
	}

	root.Flags().BoolVar(&showBuildTime, "build-time", "", false, "빌드 시간을 함께 표시합니다")

	return root
}

// runCollect는 시스템 정보를 수집하고 출력합니다
func runCollect(root *wcli.Command, showBuildTime bool) error {
	// 테스트 등에서 OutWriter 미설정 시 기본값 사용
	w := root.OutWriter
	if w == nil {
		w = os.Stdout
	}

	if showBuildTime {
		rich.Fprintln(w, "[dim]빌드 시간: %s[/dim]", BuildTime)
	}

	osName := getOSName()
	rich.Fprintln(w, "[cyan]시스템 정보를 수집 중... (%s)[/cyan]", osName)

	systemCollector, err := collector.NewCollector()
	if err != nil {
		return fmt.Errorf("수집기를 생성할 수 없습니다: %w", err)
	}

	systemInfo, err := systemCollector.CollectAll()
	if err != nil {
		return fmt.Errorf("시스템 정보를 수집할 수 없습니다: %w", err)
	}

	consoleFormatter := formatter.NewConsoleFormatter()
	fmt.Fprint(w, consoleFormatter.Format(systemInfo))

	return nil
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
