package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	return home
}

func TestDefaultConfig(t *testing.T) {
	c := Default()
	if c.AI.Provider != "ollama" {
		t.Errorf("기본 프로바이더는 ollama여야 합니다: %q", c.AI.Provider)
	}
	if c.AI.Model != "qwen3.4b" {
		t.Errorf("기본 모델은 qwen3.4b여야 합니다: %q", c.AI.Model)
	}
	if c.AI.OllamaBaseURL != "" {
		t.Errorf("Ollama URL 기본값은 비어 있어야 합니다(라이브러리 기본 사용): %q", c.AI.OllamaBaseURL)
	}
}

func TestConfigPath(t *testing.T) {
	home := tempHome(t)
	got := DefaultPath()
	want := filepath.Join(home, ".config", "pcsc", "config.json")
	if got != want {
		t.Errorf("ConfigPath = %q, want %q", got, want)
	}
}

func TestLoad_NoFile(t *testing.T) {
	tempHome(t)
	c, err := Load()
	if err != nil {
		t.Fatalf("파일이 없어도 에러 없이 기본값을 반환해야 합니다: %v", err)
	}
	if c.AI.Provider != "ollama" {
		t.Errorf("기본값이 반환되어야 합니다: %q", c.AI.Provider)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	tempHome(t)

	c := Default()
	c.AI.Provider = "openai"
	c.AI.Model = "gpt-4o-mini"
	c.AI.OllamaBaseURL = "http://localhost:1234/v1"

	if err := Save(c); err != nil {
		t.Fatalf("저장 실패: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("로드 실패: %v", err)
	}
	if loaded.AI.Provider != "openai" {
		t.Errorf("Provider 라운드트립 실패: %q", loaded.AI.Provider)
	}
	if loaded.AI.Model != "gpt-4o-mini" {
		t.Errorf("Model 라운드트립 실패: %q", loaded.AI.Model)
	}
	if loaded.AI.OllamaBaseURL != "http://localhost:1234/v1" {
		t.Errorf("OllamaBaseURL 라운드트립 실패: %q", loaded.AI.OllamaBaseURL)
	}
}

func TestSave_CreatesDir(t *testing.T) {
	home := tempHome(t)

	c := Default()
	if err := Save(c); err != nil {
		t.Fatalf("저장 실패: %v", err)
	}
	path := DefaultPath()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("설정 파일이 생성되어야 합니다: %v", err)
	}
	if !strings.Contains(path, ".config") {
		t.Errorf("경로가 XDG 설정 디렉터리여야 합니다: %q", path)
	}
	_ = home
}

func TestSaveLoad_InvalidJSON(t *testing.T) {
	home := tempHome(t)
	path := filepath.Join(home, ".config", "pcsc")
	os.MkdirAll(path, 0o755)
	os.WriteFile(filepath.Join(path, "config.json"), []byte("{invalid"), 0o644)

	if _, err := Load(); err == nil {
		t.Error("잘못된 JSON에서 에러가 반환되어야 합니다")
	}
}
