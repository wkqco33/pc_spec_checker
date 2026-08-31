package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestAIFlagsRegistered(t *testing.T) {
	root := newRootCommand()
	var out bytes.Buffer
	root.OutWriter = &out

	if err := root.Execute([]string{"--help"}); err != nil {
		t.Fatalf("Execute 실패: %v", err)
	}
	outStr := out.String()
	for _, flag := range []string{"--ai", "--ai-model"} {
		if !strings.Contains(outStr, flag) {
			t.Errorf("도움말에 %s 플래그가 없습니다", flag)
		}
	}
}

func TestExecuteWithAIFlag_NoLLMServer(t *testing.T) {
	var out, errOut bytes.Buffer
	root := newRootCommand()
	root.OutWriter = &out
	root.ErrWriter = &errOut

	// LLM 서버가 없어도 명령 자체는 성공해야 합니다 (graceful degradation)
	if err := root.Execute([]string{"--ai"}); err != nil {
		t.Fatalf("--ai 실행 실패: %v", err)
	}

	outStr := out.String()
	if !strings.Contains(outStr, "시스템 정보를 수집 중") {
		t.Errorf("수집 시작 메시지가 없습니다: %q", outStr)
	}
	if strings.Contains(errOut.String(), "panic") {
		t.Errorf("panic이 발생했습니다: %q", errOut.String())
	}
}
