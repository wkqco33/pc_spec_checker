package formatter

import (
	"strings"
	"testing"

	"wkqcosoft.com/m/analyzer"
)

func sampleResult() *analyzer.AnalysisResult {
	return &analyzer.AnalysisResult{
		Summary: "전반적으로 균형 잡힌 사양입니다.",
		Verdicts: []analyzer.UsageVerdict{
			{Category: analyzer.CategoryGame, Grade: analyzer.GradeGood, Score: 70, Comment: "인디 게임은 원활합니다."},
			{Category: analyzer.CategoryDev, Grade: analyzer.GradeFair, Score: 50, Comment: "대형 빌드는 느립니다."},
			{Category: analyzer.CategoryVideo, Grade: analyzer.GradeGood, Score: 75, Comment: "4K 재생 가능."},
			{Category: analyzer.CategoryWeb, Grade: analyzer.GradeExcellent, Score: 90, Comment: "충분합니다."},
			{Category: analyzer.CategoryOffice, Grade: analyzer.GradeExcellent, Score: 95, Comment: "과합니다."},
		},
		Upgrade: "메모리 추가를 권장합니다.",
	}
}

func TestFormatAnalysis_ContainsSections(t *testing.T) {
	out := FormatAnalysis(sampleResult())

	checks := []string{
		"AI 사양 분석",
		"전반적으로 균형 잡힌 사양입니다.",
		"게임", "개발", "영상시청", "웹서핑", "사무",
		"70", "50", "90",
		"메모리 추가를 권장합니다.",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("출력에 %q가 없습니다", want)
		}
	}
}

func TestFormatAnalysis_CategoryOrder(t *testing.T) {
	out := FormatAnalysis(sampleResult())

	gameIdx := strings.Index(out, "게임")
	devIdx := strings.Index(out, "개발")
	videoIdx := strings.Index(out, "영상시청")
	webIdx := strings.Index(out, "웹서핑")
	officeIdx := strings.Index(out, "사무")

	if gameIdx > devIdx || devIdx > videoIdx || videoIdx > webIdx || webIdx > officeIdx {
		t.Errorf("카테고리 출력 순서가 잘못되었습니다: %d %d %d %d %d",
			gameIdx, devIdx, videoIdx, webIdx, officeIdx)
	}
}

func TestFormatAIAnswer_ContainsQuestionAndAnswer(t *testing.T) {
	out := FormatAIAnswer("이 PC에서 rust 개발을 할건데 적합할까?", "네, 적합합니다. 6코어로 충분합니다.")

	checks := []string{
		"AI 질문 분석",
		"이 PC에서 rust 개발을 할건데 적합할까?",
		"네, 적합합니다. 6코어로 충분합니다.",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("출력에 %q가 없습니다", want)
		}
	}
}

func TestFormatAIAnswer_MultiLineAnswer(t *testing.T) {
	out := FormatAIAnswer("질문", "첫 줄\n둘째 줄\n셋째 줄")
	for _, line := range []string{"첫 줄", "둘째 줄", "셋째 줄"} {
		if !strings.Contains(out, line) {
			t.Errorf("여러 줄 답변이 누락되었습니다: %q", line)
		}
	}
}
