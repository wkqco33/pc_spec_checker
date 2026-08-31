package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"wkqcosoft.com/m/config"
)

func runConfigCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, errOut bytes.Buffer
	root := newRootCommand()
	root.OutWriter = &out
	root.ErrWriter = &errOut

	err := root.Execute(append([]string{"config"}, args...))
	return out.String(), err
}

func TestConfigCommand_Registered(t *testing.T) {
	out, err := runConfigCommand(t, "--help")
	if err != nil {
		t.Fatalf("config --help 실패: %v", err)
	}
	for _, want := range []string{"init", "show", "set"} {
		if !strings.Contains(out, want) {
			t.Errorf("config 도움말에 %q가 없습니다", want)
		}
	}
}

func TestConfigInit_CreatesFile(t *testing.T) {
	tempHome(t)

	out, err := runConfigCommand(t, "init")
	if err != nil {
		t.Fatalf("config init 실패: %v", err)
	}
	if !strings.Contains(out, config.DefaultPath()) {
		t.Errorf("출력에 설정 파일 경로가 없습니다: %q", out)
	}
	if _, err := os.Stat(config.DefaultPath()); err != nil {
		t.Errorf("설정 파일이 생성되지 않았습니다: %v", err)
	}
}

func TestConfigInit_ExistingFile_NoOverwrite(t *testing.T) {
	tempHome(t)

	if _, err := runConfigCommand(t, "init"); err != nil {
		t.Fatalf("config init 실패: %v", err)
	}

	// 파일 수정 후 다시 init
	cfg, _ := config.Load()
	cfg.AI.Model = "custom-model"
	if err := config.Save(cfg); err != nil {
		t.Fatalf("설정 저장 실패: %v", err)
	}

	out, err := runConfigCommand(t, "init")
	if err != nil {
		t.Fatalf("config init 재실행 실패: %v", err)
	}
	if !strings.Contains(out, "이미 존재") {
		t.Errorf("기존 파일 보호 메시지가 필요합니다: %q", out)
	}

	loaded, _ := config.Load()
	if loaded.AI.Model != "custom-model" {
		t.Errorf("기존 설정이 덮어써졌습니다: %q", loaded.AI.Model)
	}
}

func TestConfigShow_DisplaysValues(t *testing.T) {
	tempHome(t)

	out, err := runConfigCommand(t, "show")
	if err != nil {
		t.Fatalf("config show 실패: %v", err)
	}
	if !strings.Contains(out, "ollama") {
		t.Errorf("기본 프로바이더가 표시되어야 합니다: %q", out)
	}
	if !strings.Contains(out, "qwen3.4b") {
		t.Errorf("기본 모델이 표시되어야 합니다: %q", out)
	}
}

func TestConfigSet_Model(t *testing.T) {
	tempHome(t)

	if _, err := runConfigCommand(t, "set", "ai.model", "llama3.1"); err != nil {
		t.Fatalf("config set 실패: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("설정 로드 실패: %v", err)
	}
	if loaded.AI.Model != "llama3.1" {
		t.Errorf("모델이 저장되지 않았습니다: %q", loaded.AI.Model)
	}
}

func TestConfigSet_Provider_Validation(t *testing.T) {
	tempHome(t)

	if _, err := runConfigCommand(t, "set", "ai.provider", "gemini"); err == nil {
		t.Error("지원하지 않는 프로바이더에서 에러가 발생해야 합니다")
	}
}

func TestConfigSet_UnknownKey(t *testing.T) {
	tempHome(t)

	if _, err := runConfigCommand(t, "set", "nope.key", "x"); err == nil {
		t.Error("알 수 없는 키에서 에러가 발생해야 합니다")
	}
}

func TestConfigSet_MissingValue(t *testing.T) {
	tempHome(t)

	if _, err := runConfigCommand(t, "set", "ai.model"); err == nil {
		t.Error("값이 없으면 에러가 발생해야 합니다")
	}
}
