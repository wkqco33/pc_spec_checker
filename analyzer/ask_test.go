package analyzer

import (
	"context"
	"strings"
	"testing"
)

func TestBuildAnswerPrompt_ContainsQuestionAndSpecs(t *testing.T) {
	prompt := BuildAnswerPrompt(sampleInfo(), "이 PC에서 rust 개발을 할건데 적합할까?")

	checks := []string{
		"이 PC에서 rust 개발을 할건데 적합할까?",
		"AMD Ryzen 5 5600X",
		"16",
		"RTX 3060",
	}
	for _, want := range checks {
		if !strings.Contains(prompt, want) {
			t.Errorf("답변 프롬프트에 %q가 없습니다", want)
		}
	}
}

func TestBuildAnswerSystemPrompt_MentionsKoreanAndSpecs(t *testing.T) {
	prompt := BuildAnswerSystemPrompt()
	if !strings.Contains(prompt, "한국어") {
		t.Errorf("시스템 프롬프트에 한국어 지시가 필요합니다: %q", prompt)
	}
	if !strings.Contains(prompt, "사양") {
		t.Errorf("시스템 프롬프트에 사양 언급이 필요합니다: %q", prompt)
	}
	if !strings.Contains(prompt, "JSON") {
		t.Error("답변 모드는 JSON 출력을 금지하는 지시가 필요합니다")
	}
}

func TestLLMAnalyzer_Ask_Success(t *testing.T) {
	fake := &fakeClient{responses: []string{
		"네, rust 개발에 적합합니다. 6코어 12스레드로 충분합니다.",
	}}
	a := NewLLMAnalyzer(fake, "test-model")

	answer, err := a.Ask(context.Background(), sampleInfo(), "이 PC에서 rust 개발을 할건데 적합할까?")
	if err != nil {
		t.Fatalf("질문 답변 실패: %v", err)
	}
	if !strings.Contains(answer, "적합합니다") {
		t.Errorf("답변이 올바르지 않습니다: %q", answer)
	}

	// 질문이 user 메시지로 전달되었는지 확인
	req := fake.lastRequest
	lastMsg := req.Messages[len(req.Messages)-1]
	if !strings.Contains(lastMsg.Content, "rust 개발") {
		t.Errorf("질문이 프롬프트에 포함되어야 합니다: %q", lastMsg.Content)
	}
}

func TestLLMAnalyzer_Ask_EmptyAnswer_Error(t *testing.T) {
	fake := &fakeClient{responses: []string{""}}
	a := NewLLMAnalyzer(fake, "test-model")

	if _, err := a.Ask(context.Background(), sampleInfo(), "질문"); err == nil {
		t.Error("빈 답변에서 에러가 반환되어야 합니다")
	}
}

func TestLLMAnalyzer_Ask_MarkdownAnswerPassthrough(t *testing.T) {
	answer := "## 결론\n- **적합합니다**\n- 16GB 메모리 충분"
	fake := &fakeClient{responses: []string{answer}}
	a := NewLLMAnalyzer(fake, "test-model")

	got, err := a.Ask(context.Background(), sampleInfo(), "질문")
	if err != nil {
		t.Fatalf("답변 실패: %v", err)
	}
	if got != answer {
		t.Errorf("답변이 그대로 반환되어야 합니다: %q", got)
	}
}

func TestLLMAnalyzer_Ask_StripsUnwantedJSONBlock(t *testing.T) {
	// 모델이 요약 JSON을 덤으로 출력한 경우 — 코드블록 제거하고 자연어 답변만 남김
	fake := &fakeClient{responses: []string{
		"```json\n{\"summary\": \"요약\", \"verdicts\": []}\n```\n\n네, 적합합니다. 6코어로 충분합니다.",
	}}
	a := NewLLMAnalyzer(fake, "test-model")

	got, err := a.Ask(context.Background(), sampleInfo(), "질문")
	if err != nil {
		t.Fatalf("답변 실패: %v", err)
	}
	if strings.Contains(got, "```") {
		t.Errorf("JSON 코드블록이 제거되어야 합니다: %q", got)
	}
	if !strings.Contains(got, "적합합니다") {
		t.Errorf("자연어 답변이 유지되어야 합니다: %q", got)
	}
}

func TestLLMAnalyzer_Ask_KeepsMarkdownFormatting(t *testing.T) {
	// 자연어 답변의 마크다운(굵게, 목록)은 유지되어야 합니다
	answer := "**네, 적합합니다.**\n1. **CPU:** 충분합니다\n2. **메모리:** 여유가 있습니다"
	fake := &fakeClient{responses: []string{answer}}
	a := NewLLMAnalyzer(fake, "test-model")

	got, err := a.Ask(context.Background(), sampleInfo(), "질문")
	if err != nil {
		t.Fatalf("답변 실패: %v", err)
	}
	if !strings.Contains(got, "**CPU:**") && !strings.Contains(got, "**") {
		t.Errorf("마크다운 강조가 유지되어야 합니다: %q", got)
	}
}

func TestLLMAnalyzer_Ask_ClientError(t *testing.T) {
	fake := &fakeClient{err: context.DeadlineExceeded}
	a := NewLLMAnalyzer(fake, "test-model")

	if _, err := a.Ask(context.Background(), sampleInfo(), "질문"); err == nil {
		t.Error("클라이언트 오류 시 에러가 반환되어야 합니다")
	}
}
