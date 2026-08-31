package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"

	"wkqcosoft.com/m/analyzer"
	"wkqcosoft.com/m/collector"
	"wkqcosoft.com/m/formatter"
	"wkqcosoft.com/m/model"
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
		showAI        bool
		aiModel       string
		root          *wcli.Command
	)

	root = &wcli.Command{
		Use:   "pcsc",
		Short: "시스템 하드웨어 사양 정보를 수집하고 표시하는 CLI 도구입니다",
		Long: "pcsc는 Linux/macOS/Windows의 하드웨어 사양 정보를 수집해 콘솔에 표시합니다.\n\n" +
			"AI 분석:\n" +
			"  pcsc --ai                          사양 요약 분석\n" +
			"  pcsc --ai \"이 PC로 rust 개발 괜찮을까?\"  질문에 맞춘 사양 분석",
		Version: Version,
		Run: func(ctx *wcli.Context) error {
			question := extractAIQuestion(os.Args[1:])
			return runCollect(root, showBuildTime, showAI, aiModel, question)
		},
	}

	root.Flags().BoolVar(&showBuildTime, "build-time", "", false, "빌드 시간을 함께 표시합니다")
	root.Flags().BoolVar(&showAI, "ai", "", false, "AI 사양 분석을 함께 표시합니다 (기본: Ollama 로컬)")
	root.Flags().StringVar(&aiModel, "ai-model", "", "", "AI 분석에 사용할 모델명 (기본: 설정 파일 또는 PCSC_AI_MODEL)")

	root.AddCommand(newConfigCommand(root))

	return root
}

// runCollect는 시스템 정보를 수집하고 출력합니다
func runCollect(root *wcli.Command, showBuildTime, showAI bool, aiModel, question string) error {
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
	_, _ = fmt.Fprint(w, consoleFormatter.Format(systemInfo))

	if showAI {
		runAIAnalysis(w, systemInfo, aiModel, question)
	}

	return nil
}

// runAIAnalysis는 LLM으로 사양 분석을 수행하고 출력합니다. 실패해도 경고만 표시합니다.
// 질문이 있으면 질문 답변 모드, 없으면 구조화 요약 모드로 동작합니다
func runAIAnalysis(w io.Writer, systemInfo *model.SystemInfo, aiModel, question string) {
	client, err := analyzer.NewClientFromConfig()
	if err != nil {
		rich.Fprintln(w, "[yellow]AI 분석을 건너뜁니다: %v[/yellow]", err)
		return
	}

	if aiModel == "" {
		aiModel = analyzer.ModelWithConfig()
	}

	ai := analyzer.NewLLMAnalyzer(client, aiModel)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if question != "" {
		rich.Fprintln(w, "[cyan]AI 질문 분석 중... (LLM 응답 대기)[/cyan]")
		answer, err := ai.Ask(ctx, systemInfo, question)
		if err != nil {
			rich.Fprintln(w, "[yellow]AI 분석 실패 (기본 출력은 유지됩니다): %v[/yellow]", err)
			return
		}
		_, _ = fmt.Fprint(w, formatter.FormatAIAnswer(question, answer))
		return
	}

	rich.Fprintln(w, "[cyan]AI 사양 분석 중... (LLM 응답 대기)[/cyan]")
	result, err := ai.Analyze(ctx, systemInfo)
	if err != nil {
		rich.Fprintln(w, "[yellow]AI 분석 실패 (기본 출력은 유지됩니다): %v[/yellow]", err)
		return
	}

	_, _ = fmt.Fprint(w, formatter.FormatAnalysis(result))
}

// extractAIQuestion은 원본 인자에서 --ai 이후의 위치 인자(질문)를 추출합니다
func extractAIQuestion(args []string) string {
	var parts []string
	valueFlags := map[string]bool{"--ai-model": true}

	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			continue
		}
		if strings.HasPrefix(a, "-") {
			// 값을 별도 인자로 받는 플래그면 그 값도 건너뜀
			if valueFlags[a] && i+1 < len(args) {
				i++
			}
			continue
		}
		parts = append(parts, a)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

// hasAIQuestion은 --ai 사용 시 위치 인자(질문)가 있는지 확인합니다
func hasAIQuestion(args []string) bool {
	return extractAIQuestion(args) != ""
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
