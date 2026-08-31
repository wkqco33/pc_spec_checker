package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestQuestionDetection_IsQuestionMode(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{[]string{"--ai", "이 PC에서 rust 개발을 할건데 적합할까?"}, true},
		{[]string{"--ai"}, false},
		{[]string{"--ai", "--ai-model", "gpt-4o", "게임용으로 괜찮나요?"}, true},
		{[]string{"--build-time"}, false},
		{[]string{}, false},
	}
	for _, c := range cases {
		if got := hasAIQuestion(c.args); got != c.want {
			t.Errorf("hasAIQuestion(%v) = %v, want %v", c.args, got, c.want)
		}
	}
}

func TestExtractAIQuestion(t *testing.T) {
	question := extractAIQuestion([]string{"--ai", "--ai-model", "m1", "이 PC로 게임 가능?"})
	if question != "이 PC로 게임 가능?" {
		t.Errorf("질문 추출 실패: %q", question)
	}
}

func TestExecuteWithAIQuestion_NoLLMServer(t *testing.T) {
	var out, errOut bytes.Buffer
	root := newRootCommand()
	root.OutWriter = &out
	root.ErrWriter = &errOut

	// LLM 서버가 없어도 명령 자체는 성공해야 합니다 (graceful degradation)
	if err := root.Execute([]string{"--ai", "이 PC에서 개발해도 될까요?"}); err != nil {
		t.Fatalf("--ai 질문 모드 실행 실패: %v", err)
	}
	if !strings.Contains(out.String(), "시스템 정보를 수집 중") {
		t.Errorf("수집 시작 메시지가 없습니다: %q", out.String())
	}
}
