package analyzer

import (
	"fmt"
	"strconv"
	"strings"

	"wkqcosoft.com/m/model"
)

// BuildSystemPrompt는 분석기의 시스템 프롬프트를 반환합니다
func BuildSystemPrompt() string {
	return "당신은 PC 하드웨어 전문가입니다. 사용자가 제공한 PC 사양을 분석하여 JSON 형식으로만 답변합니다. " +
		"게임, 개발, 영상시청, 웹서핑, 사무 용도 각각에 대해 등급(excellent/good/fair/poor), " +
		"점수(0~100), 코멘트를 한국어로 작성하고, 종합 요약과 업그레이드 제안을 포함합니다."
}

// BuildUserPrompt는 시스템 정보를 분석 요청 프롬프트로 변환합니다
func BuildUserPrompt(info *model.SystemInfo) string {
	var s strings.Builder
	s.WriteString("다음 PC 사양을 분석해주세요.\n\n")

	s.WriteString("CPU: ")
	s.WriteString(info.CPU.Model)
	fmt.Fprintf(&s, " (%d코어/%d스레드, 최대 %d MHz)", info.CPU.Cores, info.CPU.Threads, info.CPU.MaxFreqMHz)
	s.WriteString("\n")

	fmt.Fprintf(&s, "메모리: %sGB (사용 가능 %sGB)\n",
		formatGB(info.Memory.TotalGB), formatGB(info.Memory.AvailableGB))

	for _, st := range info.Storage {
		fmt.Fprintf(&s, "저장장치 [%s]: %sGB (%s)\n", st.MountPoint, formatGB(st.TotalGB), st.Type)
	}

	for _, g := range info.GPU {
		s.WriteString("GPU: ")
		s.WriteString(g.Name)
		s.WriteString(" [")
		s.WriteString(g.Vendor)
		s.WriteString("]")
		if g.MemoryGB > 0 {
			s.WriteString(" " + formatGB(g.MemoryGB) + "GB")
		}
		s.WriteString("\n")
	}

	s.WriteString("\n각 용도(game, dev, video, web, office)에 대한 평가를 JSON으로 작성해주세요.\n")
	s.WriteString(schemaHint())

	return s.String()
}

// schemaHint는 응답 형식을 모델에 명확히 전달하기 위한 JSON 예시입니다
func schemaHint() string {
	return `응답은 반드시 아래 구조의 JSON만 출력해야 합니다 (코드블록·설명 금지):
{"summary": "종합 요약", "verdicts": [{"category": "game|dev|video|web|office 중 하나", "grade": "excellent|good|fair|poor 중 하나", "score": 0에서 100 사이 정수, "comment": "평가 코멘트"}], "upgrade": "업그레이드 제안"}
verdicts 배열에는 5개 용도(game, dev, video, web, office)가 모두 포함되어야 합니다.`
}

// ResponseSchema는 구조화 출력에 사용할 JSON 스키마를 반환합니다
func ResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary": map[string]any{
				"type":        "string",
				"description": "PC 사양에 대한 종합 요약 (한국어)",
			},
			"verdicts": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"category": map[string]any{
							"type": "string",
							"enum": AllCategories,
						},
						"grade": map[string]any{
							"type": "string",
							"enum": []string{GradeExcellent, GradeGood, GradeFair, GradePoor},
						},
						"score": map[string]any{
							"type":        "integer",
							"minimum":     0,
							"maximum":     100,
							"description": "성능 점수 (0~100)",
						},
						"comment": map[string]any{
							"type":        "string",
							"description": "평가 코멘트 (한국어)",
						},
					},
					"required": []string{"category", "grade", "score", "comment"},
				},
			},
			"upgrade": map[string]any{
				"type":        "string",
				"description": "업그레이드 제안 (한국어)",
			},
		},
		"required": []string{"summary", "verdicts", "upgrade"},
	}
}

func formatGB(gb float64) string {
	if gb == float64(int64(gb)) {
		return strconv.FormatInt(int64(gb), 10)
	}
	return strconv.FormatFloat(gb, 'f', -1, 64)
}
