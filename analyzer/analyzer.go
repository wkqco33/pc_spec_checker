package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	llm "github.com/wkqco33/LLM_client_go"

	"wkqcosoft.com/m/model"
)

// maxAttempts는 LLM 요청 최대 시도 횟수입니다 (최초 1회 + 교정 재요청 1회)
const maxAttempts = 2

// repairHint는 응답이 스키마를 벗어났을 때 재요청에 포함되는 교정 지시입니다
const repairHint = "이전 응답은 요청한 JSON 형식에 맞지 않습니다. " +
	"별도 설명 없이 아래 스키마와 동일한 구조의 유효한 JSON만 다시 출력해주세요."

// Analyzer는 시스템 정보를 분석하는 인터페이스입니다
type Analyzer interface {
	Analyze(ctx context.Context, info *model.SystemInfo) (*AnalysisResult, error)
}

// LLMAnalyzer는 llm.Client를 통해 사양 분석을 수행합니다
type LLMAnalyzer struct {
	client   llm.Client
	model    string
	debugLog string // PCSC_AI_DEBUG=1일 때 원본 응답 기록
}

// NewLLMAnalyzer는 새로운 LLMAnalyzer를 생성합니다
func NewLLMAnalyzer(client llm.Client, model string) *LLMAnalyzer {
	return &LLMAnalyzer{client: client, model: model}
}

// Analyze는 시스템 정보를 LLM에 전달해 분석 결과를 반환합니다.
// 응답이 스키마에 맞지 않으면 교정 지시와 함께 1회 재요청합니다.
func (a *LLMAnalyzer) Analyze(ctx context.Context, info *model.SystemInfo) (*AnalysisResult, error) {
	baseMessages := []llm.Message{
		{Role: llm.RoleSystem, Content: BuildSystemPrompt()},
		{Role: llm.RoleUser, Content: BuildUserPrompt(info)},
	}

	messages := baseMessages
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := a.client.Complete(ctx, llm.ChatRequest{
			Model:    a.model,
			Messages: messages,
		})
		if err != nil {
			return nil, fmt.Errorf("LLM 요청 실패: %w", err)
		}

		if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
			lastErr = fmt.Errorf("LLM 응답이 비어 있습니다")
			continue
		}

		content := resp.Choices[0].Message.Content
		if a.debugEnabled() {
			a.debugLog += fmt.Sprintf("=== 시도 %d 원본 응답 ===\n%s\n", attempt, content)
			fmt.Fprintf(os.Stderr, "[pcsc-ai-debug] %s\n", a.debugLog)
		}

		result, err := a.parseResult(content)
		if err == nil {
			return result, nil
		}
		lastErr = err

		// 마지막 시도가 아니면 교정 지시를 붙여 재요청
		if attempt < maxAttempts {
			messages = append(messages,
				llm.Message{Role: llm.RoleAssistant, Content: content},
				llm.Message{Role: llm.RoleUser, Content: repairHint + "\n\n" + schemaHint()},
			)
		}
	}

	return nil, fmt.Errorf("분석 결과를 얻지 못했습니다 (%d회 시도): %w", maxAttempts, lastErr)
}

// parseResult는 응답 내용을 파싱하고 검증합니다
func (a *LLMAnalyzer) parseResult(content string) (*AnalysisResult, error) {
	extracted := extractJSON(content)
	if !json.Valid([]byte(extracted)) {
		// 잘린 JSON(max_tokens 초과 등)을 닫는 문자로 복구 시도
		extracted = repairTruncatedJSON(extracted)
	}

	var result AnalysisResult
	if err := json.Unmarshal([]byte(extracted), &result); err != nil {
		return nil, fmt.Errorf("분석 결과 파싱 실패: %w", err)
	}
	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("분석 결과가 유효하지 않습니다: %w", err)
	}
	return &result, nil
}

// repairTruncatedJSON은 잘린 JSON의 열린 괄호를 닫아 복구를 시도합니다
func repairTruncatedJSON(s string) string {
	s = strings.TrimRight(s, " \t\n\r,") // 마지막 쉼표/공백 제거
	if s == "" {
		return s
	}

	// 마지막 유효 지점 탐색: 잘린 값이면 그 앞의 쉼표나 중괄호까지 되돌림
	if lastBrace := strings.LastIndex(s, "}"); lastBrace >= 0 {
		s = s[:lastBrace+1]
	}

	var stack []byte
	inString := false
	escape := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if escape {
			escape = false
			continue
		}
		switch c {
		case '\\':
			if inString {
				escape = true
			}
		case '"':
			inString = !inString
		case '{', '[':
			if !inString {
				stack = append(stack, c)
			}
		case '}', ']':
			if !inString && len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}

	// 문자열이 열려 있으면 닫고, 열린 괄호를 역순으로 닫음
	if inString {
		s += `"`
	}
	for i := len(stack) - 1; i >= 0; i-- {
		switch stack[i] {
		case '{':
			s += "}"
		case '[':
			s += "]"
		}
	}
	// 배열/객체 끝 잘림 뒤 남은 쉼표 정리
	s = strings.TrimRight(s, ",")
	return s
}

// debugEnabled는 디버그 모드 여부를 반환합니다
func (a *LLMAnalyzer) debugEnabled() bool {
	return os.Getenv("PCSC_AI_DEBUG") == "1"
}

// extractJSON은 응답에서 JSON 본문만 추출합니다 (마크다운 코드블록 등 제거)
func extractJSON(content string) string {
	content = strings.TrimSpace(content)

	if idx := strings.Index(content, "```json"); idx >= 0 {
		content = content[idx+len("```json"):]
	} else if idx := strings.Index(content, "```"); idx >= 0 {
		content = content[idx+len("```"):]
	}

	if idx := strings.Index(content, "```"); idx >= 0 {
		content = content[:idx]
	}

	content = strings.TrimSpace(content)

	// 코드블록 없이 앞뒤 설명 텍스트가 붙는 경우 첫 { 또는 [부터 잘라냄
	if start := strings.IndexAny(content, "{["); start > 0 {
		content = content[start:]
	}

	return strings.TrimSpace(content)
}
