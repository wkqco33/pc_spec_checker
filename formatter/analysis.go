package formatter

import (
	"fmt"
	"strings"

	"wkqcosoft.com/m/analyzer"
)

// categoryLabels는 카테고리 코드의 표시 이름입니다
var categoryLabels = map[string]string{
	analyzer.CategoryGame:   "게임",
	analyzer.CategoryDev:    "개발",
	analyzer.CategoryVideo:  "영상시청",
	analyzer.CategoryWeb:    "웹서핑",
	analyzer.CategoryOffice: "사무",
}

// categoryOrder는 출력 순서가 고정된 카테고리 목록입니다
var categoryOrder = []string{
	analyzer.CategoryGame,
	analyzer.CategoryDev,
	analyzer.CategoryVideo,
	analyzer.CategoryWeb,
	analyzer.CategoryOffice,
}

// FormatAnalysis는 AI 분석 결과를 콘솔용 텍스트로 포맷팅합니다
func FormatAnalysis(result *analyzer.AnalysisResult) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("┌─ AI 사양 분석\n")
	fmt.Fprintf(&b, "│  %s\n", result.Summary)
	b.WriteString("│\n")

	byCategory := make(map[string]analyzer.UsageVerdict, len(result.Verdicts))
	for _, v := range result.Verdicts {
		byCategory[v.Category] = v
	}

	for _, cat := range categoryOrder {
		v, ok := byCategory[cat]
		if !ok {
			continue
		}
		fmt.Fprintf(&b, "│  %-8s [%s] %3d점 — %s\n",
			categoryLabels[cat], analyzer.GradeLabel(v.Grade), v.Score, v.Comment)
	}

	b.WriteString("│\n")
	fmt.Fprintf(&b, "│  업그레이드 제안: %s\n", result.Upgrade)
	b.WriteString("└─────────────────────────────────────────────────────────────────\n")

	return b.String()
}

// FormatAIAnswer는 AI 질문 답변을 콘솔용 텍스트로 포맷팅합니다
func FormatAIAnswer(question, answer string) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("┌─ AI 질문 분석\n")
	fmt.Fprintf(&b, "│  질문: %s\n", question)
	b.WriteString("│\n")

	for _, line := range strings.Split(answer, "\n") {
		fmt.Fprintf(&b, "│  %s\n", line)
	}

	b.WriteString("└─────────────────────────────────────────────────────────────────\n")
	return b.String()
}
