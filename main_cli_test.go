package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewRootCommand(t *testing.T) {
	root := newRootCommand()
	if root == nil {
		t.Fatal("newRootCommand()가 nil을 반환했습니다")
	}
	if !strings.HasPrefix(root.Use, "pcsc") {
		t.Errorf("Use = %q, pcsc 접두사가 필요합니다", root.Use)
	}
	if root.Short == "" {
		t.Error("Short 설명이 비어 있습니다")
	}
	if root.Version == "" {
		t.Error("Version이 설정되지 않았습니다")
	}
	if root.Run == nil {
		t.Error("Run이 nil입니다")
	}
	if root.Flags() == nil {
		t.Error("Flags가 초기화되지 않았습니다")
	}
}

func TestBuildTimeFlag(t *testing.T) {
	root := newRootCommand()
	if err := root.Execute([]string{"--build-time"}); err != nil {
		t.Fatalf("build-time 플래그 실행 실패: %v", err)
	}
}

func TestExecuteVersion(t *testing.T) {
	var out, errOut bytes.Buffer
	root := newRootCommand()
	root.Version = "9.9.9-test"
	root.OutWriter = &out
	root.ErrWriter = &errOut

	if err := root.Execute([]string{"--version"}); err != nil {
		t.Fatalf("Execute 실패: %v", err)
	}
	if !strings.Contains(out.String(), "9.9.9-test") {
		t.Errorf("버전 출력에 9.9.9-test가 없습니다: %q", out.String())
	}
	if errOut.Len() > 0 {
		t.Errorf("에러 출력이 예상치 않게 발생했습니다: %q", errOut.String())
	}
}

func TestExecuteHelp(t *testing.T) {
	var out, errOut bytes.Buffer
	root := newRootCommand()
	root.OutWriter = &out
	root.ErrWriter = &errOut

	if err := root.Execute([]string{"--help"}); err != nil {
		t.Fatalf("Execute 실패: %v", err)
	}
	outStr := out.String()
	if !strings.Contains(outStr, "pcsc") {
		t.Errorf("도움말에 pcsc가 없습니다: %q", outStr)
	}
	if !strings.Contains(outStr, "build-time") {
		t.Errorf("도움말에 build-time 플래그가 없습니다: %q", outStr)
	}
}

func TestExecuteUnknownFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	root := newRootCommand()
	root.OutWriter = &out
	root.ErrWriter = &errOut

	if err := root.Execute([]string{"--no-such-flag"}); err == nil {
		t.Error("알 수 없는 플래그에서 에러가 반환되어야 합니다")
	}
}

func TestCollectFlow(t *testing.T) {
	var out, errOut bytes.Buffer
	root := newRootCommand()
	root.OutWriter = &out
	root.ErrWriter = &errOut

	if err := root.Execute([]string{}); err != nil {
		t.Fatalf("Execute 실패: %v", err)
	}
	outStr := out.String()
	if !strings.Contains(outStr, "시스템 정보를 수집 중") {
		t.Errorf("수집 시작 메시지가 없습니다: %q", outStr)
	}
}
