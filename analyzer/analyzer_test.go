package analyzer

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	llm "github.com/wkqco33/LLM_client_go"
)

// fakeClient는 llm.Client를 흉내 내는 테스트 더블입니다
type fakeClient struct {
	responses   []string // 호출 순서대로 반환할 응답
	err         error
	lastRequest llm.ChatRequest
	callCount   int
}

func (f *fakeClient) Complete(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	f.lastRequest = req
	f.callCount++
	if f.err != nil {
		return nil, f.err
	}
	idx := f.callCount - 1
	if idx >= len(f.responses) {
		idx = len(f.responses) - 1
	}
	return &llm.ChatResponse{
		Choices: []llm.Choice{
			{Message: llm.Message{Content: f.responses[idx]}},
		},
	}, nil
}

func (f *fakeClient) Stream(ctx context.Context, req llm.ChatRequest) (llm.Stream, error) {
	return nil, nil
}

func (f *fakeClient) CreateEmbeddings(ctx context.Context, req llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return nil, nil
}

func (f *fakeClient) TokenCounter(model string) any {
	return nil
}

func validJSONResponse() string {
	data, _ := json.Marshal(validResult())
	return string(data)
}

func TestLLMAnalyzer_Analyze_Success(t *testing.T) {
	fake := &fakeClient{responses: []string{validJSONResponse()}}
	a := NewLLMAnalyzer(fake, "test-model")

	result, err := a.Analyze(context.Background(), sampleInfo())
	if err != nil {
		t.Fatalf("분석 실패: %v", err)
	}

	if result.Summary != "전반적으로 균형 잡힌 사양입니다." {
		t.Errorf("요약이 올바르지 않습니다: %q", result.Summary)
	}
	if !result.AllCategoriesComplete() {
		t.Error("5개 카테고리 평가가 모두 있어야 합니다")
	}
}

func TestLLMAnalyzer_Analyze_SendsCorrectRequest(t *testing.T) {
	fake := &fakeClient{responses: []string{validJSONResponse()}}
	a := NewLLMAnalyzer(fake, "test-model")

	if _, err := a.Analyze(context.Background(), sampleInfo()); err != nil {
		t.Fatalf("분석 실패: %v", err)
	}

	req := fake.lastRequest
	if req.Model != "test-model" {
		t.Errorf("모델명이 전달되지 않았습니다: %q", req.Model)
	}
	if len(req.Messages) != 2 {
		t.Fatalf("메시지 개수가 2이어야 합니다: %d", len(req.Messages))
	}
	if req.Messages[0].Role != llm.RoleSystem {
		t.Errorf("첫 메시지는 system이어야 합니다: %q", req.Messages[0].Role)
	}
	if req.Messages[1].Role != llm.RoleUser {
		t.Errorf("두번째 메시지는 user여야 합니다: %q", req.Messages[1].Role)
	}
}

func TestLLMAnalyzer_Analyze_MarkdownWrappedJSON(t *testing.T) {
	inner := validJSONResponse()
	wrapped := "```json\n" + inner + "\n```"
	fake := &fakeClient{responses: []string{wrapped}}
	a := NewLLMAnalyzer(fake, "test-model")

	result, err := a.Analyze(context.Background(), sampleInfo())
	if err != nil {
		t.Fatalf("마크다운 래핑 JSON 파싱 실패: %v", err)
	}
	if result.Summary == "" {
		t.Error("요약이 비어 있습니다")
	}
}

func TestLLMAnalyzer_Analyze_InvalidJSON(t *testing.T) {
	fake := &fakeClient{responses: []string{"분석할 수 없습니다"}}
	a := NewLLMAnalyzer(fake, "test-model")

	if _, err := a.Analyze(context.Background(), sampleInfo()); err == nil {
		t.Error("유효하지 않은 JSON에서 에러가 반환되어야 합니다")
	}
}

func TestLLMAnalyzer_Analyze_ValidationError(t *testing.T) {
	data, _ := json.Marshal(validResult())
	var m map[string]any
	json.Unmarshal(data, &m)
	verdicts := m["verdicts"].([]any)
	verdicts[0].(map[string]any)["category"] = "invalid-category"
	corrupted, _ := json.Marshal(m)

	fake := &fakeClient{responses: []string{string(corrupted)}}
	a := NewLLMAnalyzer(fake, "test-model")

	if _, err := a.Analyze(context.Background(), sampleInfo()); err == nil {
		t.Error("스키마 위반 결과에서 에러가 반환되어야 합니다")
	}
}

func TestLLMAnalyzer_Analyze_ClientError(t *testing.T) {
	fake := &fakeClient{err: context.DeadlineExceeded}
	a := NewLLMAnalyzer(fake, "test-model")

	if _, err := a.Analyze(context.Background(), sampleInfo()); err == nil {
		t.Error("클라이언트 오류 시 에러가 반환되어야 합니다")
	}
}

func TestLLMAnalyzer_Analyze_EmptyContent(t *testing.T) {
	fake := &fakeClient{responses: []string{""}}
	a := NewLLMAnalyzer(fake, "test-model")

	if _, err := a.Analyze(context.Background(), sampleInfo()); err == nil {
		t.Error("빈 응답에서 에러가 반환되어야 합니다")
	}
}

func TestLLMAnalyzer_Analyze_RetryAfterInvalidResponse(t *testing.T) {
	// 1차: verdicts 누락 → 2차: 정상 응답으로 재요청 성공
	fake := &fakeClient{responses: []string{
		`{"summary": "요약", "verdicts": [], "upgrade": "없음"}`,
		validJSONResponse(),
	}}
	a := NewLLMAnalyzer(fake, "test-model")

	result, err := a.Analyze(context.Background(), sampleInfo())
	if err != nil {
		t.Fatalf("재요청 후 분석 실패: %v", err)
	}
	if !result.AllCategoriesComplete() {
		t.Error("재요청 결과가 정상이어야 합니다")
	}
	if fake.callCount != 2 {
		t.Errorf("총 2회 호출되어야 합니다: %d", fake.callCount)
	}
}

func TestLLMAnalyzer_Analyze_RetrySendsRepairHint(t *testing.T) {
	fake := &fakeClient{responses: []string{
		`{"summary": "요약"}`,
		validJSONResponse(),
	}}
	a := NewLLMAnalyzer(fake, "test-model")

	if _, err := a.Analyze(context.Background(), sampleInfo()); err != nil {
		t.Fatalf("재요청 후 분석 실패: %v", err)
	}

	// 마지막 요청에는 이전 응답과 교정 지시가 포함되어야 합니다
	msgs := fake.lastRequest.Messages
	found := false
	for _, m := range msgs {
		if strings.Contains(m.Content, "형식에 맞지 않습니다") {
			found = true
		}
	}
	if !found {
		t.Error("재요청에 교정 지시가 포함되어야 합니다")
	}
}

func TestLLMAnalyzer_Analyze_RetryFailsTwice_ReturnsError(t *testing.T) {
	fake := &fakeClient{responses: []string{
		"무효 응답 1",
		"무효 응답 2",
	}}
	a := NewLLMAnalyzer(fake, "test-model")

	if _, err := a.Analyze(context.Background(), sampleInfo()); err == nil {
		t.Error("두 번 모두 실패하면 에러가 반환되어야 합니다")
	}
	if fake.callCount != 2 {
		t.Errorf("총 2회 호출 후 포기해야 합니다: %d", fake.callCount)
	}
}

func TestLLMAnalyzer_DebugDumpsRawResponse(t *testing.T) {
	t.Setenv("PCSC_AI_DEBUG", "1")

	fake := &fakeClient{responses: []string{"무효 응답", validJSONResponse()}}
	a := NewLLMAnalyzer(fake, "test-model")

	if _, err := a.Analyze(context.Background(), sampleInfo()); err != nil {
		t.Fatalf("재요청 후 분석 실패: %v", err)
	}
	if a.debugLog == "" {
		t.Error("PCSC_AI_DEBUG 설정 시 원본 응답이 기록되어야 합니다")
	}
	if !strings.Contains(a.debugLog, "무효 응답") {
		t.Errorf("무효 응답이 기록되어야 합니다: %q", a.debugLog)
	}
}

func TestRepairTruncatedJSON_ClosesBraces(t *testing.T) {
	// verdicts 배열 중간에 잘린 응답
	truncated := `{"summary": "요약", "verdicts": [{"category": "game", "grade": "good", "score": 70, "comment": "좋음"}, {"category": "dev"`
	repaired := repairTruncatedJSON(truncated)
	var r AnalysisResult
	if err := json.Unmarshal([]byte(repaired), &r); err != nil {
		t.Fatalf("복구된 JSON 파싱 실패: %v", err)
	}
	if r.Summary != "요약" {
		t.Errorf("복구 결과가 올바르지 않습니다: %q", r.Summary)
	}
}

func TestRepairTruncatedJSON_CompleteJSONUnchanged(t *testing.T) {
	complete := validJSONResponse()
	if got := repairTruncatedJSON(complete); got != complete {
		t.Error("완전한 JSON은 변경되지 않아야 합니다")
	}
}

func TestLLMAnalyzer_Analyze_TruncatedResponse_Recovered(t *testing.T) {
	// 1차: 잘린 응답 → 복구 후 바로 성공 (재요청 없이)
	fake := &fakeClient{responses: []string{
		`{"summary": "복구 요약", "verdicts": [{"category": "game", "grade": "good", "score": 70, "comment": "좋음"}`,
	}}
	a := NewLLMAnalyzer(fake, "test-model")

	result, err := a.Analyze(context.Background(), sampleInfo())
	if err != nil {
		t.Fatalf("잘린 응답 복구 실패: %v", err)
	}
	if result.Summary != "복구 요약" {
		t.Errorf("복구된 요약이 올바르지 않습니다: %q", result.Summary)
	}
}

func TestExtractJSON_Plain(t *testing.T) {
	inner := `{"a": 1}`
	got := extractJSON(inner)
	if !strings.HasPrefix(got, "{") {
		t.Errorf("JSON 추출 실패: %q", got)
	}
}

func TestExtractJSON_Markdown(t *testing.T) {
	got := extractJSON("설명\n```json\n{\"a\": 1}\n```\n끝")
	if got != `{"a": 1}` {
		t.Errorf("마크다운에서 JSON 추출 실패: %q", got)
	}
}
