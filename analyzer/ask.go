package analyzer

import (
	"context"
	"fmt"
	"strings"

	llm "github.com/wkqco33/LLM_client_go"

	"wkqcosoft.com/m/model"
)

// BuildAnswerSystemPrompt는 질문 답변 모드의 시스템 프롬프트를 반환합니다
func BuildAnswerSystemPrompt() string {
	return "당신은 PC 하드웨어 전문가입니다. 사용자가 제공한 PC 사양을 바탕으로 " +
		"사용자의 질문에 한국어로 답변합니다. 적합 여부, 이유, 필요시 업그레이드 제안을 포함하고, " +
		"간결하고 균형 잡힌 답변을 제공합니다. " +
		"주의: 이 모드는 대화형 답변 모드이므로 사양 요약 JSON을 출력하지 마세요. " +
		"마크다운 코드블록(```)도 사용하지 말고 자연어로만 답변해주세요."
}

// BuildAnswerPrompt는 사양과 사용자 질문을 답변 요청 프롬프트로 변환합니다
func BuildAnswerPrompt(info *model.SystemInfo, question string) string {
	var s strings.Builder
	s.WriteString("다음 PC 사양을 참고해주세요.\n\n")
	s.WriteString(BuildUserPrompt(info))
	s.WriteString("\n\n위 PC 사양에 대해 다음 질문에 답변해주세요:\n")
	s.WriteString(question)
	return s.String()
}

// Ask는 사양과 사용자 질문을 LLM에 전달해 자연어 답변을 반환합니다
func (a *LLMAnalyzer) Ask(ctx context.Context, info *model.SystemInfo, question string) (string, error) {
	resp, err := a.client.Complete(ctx, llm.ChatRequest{
		Model: a.model,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: BuildAnswerSystemPrompt()},
			{Role: llm.RoleUser, Content: BuildAnswerPrompt(info, question)},
		},
	})
	if err != nil {
		return "", fmt.Errorf("LLM 요청 실패: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("LLM 응답이 비어 있습니다")
	}

	answer := cleanAnswer(resp.Choices[0].Message.Content)
	if answer == "" {
		return "", fmt.Errorf("LLM 응답이 비어 있습니다")
	}
	return answer, nil
}

// cleanAnswer는 답변에서 불필요한 요소를 제거합니다.
// 모델이 사양 요약을 JSON 코드블록으로 덤으로 출력하는 경우 이를 걷어냅니다
func cleanAnswer(content string) string {
	s := strings.TrimSpace(content)

	// JSON 코드블록(```json ... ```) 제거 — verdicts/summary 등 요약 전용 필드 포함 시에만
	for {
		start := strings.Index(s, "```json")
		if start < 0 {
			break
		}
		rest := s[start:]
		end := strings.Index(rest[len("```json"):], "```")
		if end < 0 {
			break
		}
		block := rest[len("```json") : len("```json")+end]
		if isSummaryJSON(block) {
			s = s[:start] + rest[len("```json")+end+3:]
			continue
		}
		break
	}

	// 앞뒤 빈 줄/공백 정리
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "\n")
	return strings.TrimSpace(s)
}

// isSummaryJSON은 불록이 사양 요약 JSON인지 판별합니다
func isSummaryJSON(block string) bool {
	return strings.Contains(block, "verdicts") ||
		(strings.Contains(block, "summary") && strings.Contains(block, "upgrade"))
}
