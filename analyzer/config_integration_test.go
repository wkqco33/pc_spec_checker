package analyzer

import (
	"path/filepath"
	"testing"

	pkgconfig "wkqcosoft.com/m/config"
)

func writeConfigFile(t *testing.T, c *pkgconfig.Config) {
	t.Helper()
	// 기본 경로를 테스트 홈 아래로 우회
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	if err := pkgconfig.Save(c); err != nil {
		t.Fatalf("설정 파일 저장 실패: %v", err)
	}
}

func TestEffectiveConfig_FromConfigFile(t *testing.T) {
	writeConfigFile(t, &pkgconfig.Config{
		AI: pkgconfig.AIConfig{
			Provider: "openai",
			Model:    "gpt-4o",
		},
	})
	setEnv(t, "PCSC_AI_PROVIDER", "")
	setEnv(t, "PCSC_AI_MODEL", "")

	provider := ProviderFromConfig()
	if provider != "openai" {
		t.Errorf("설정 파일 프로바이더 미반영: %q", provider)
	}
	if got := ModelWithConfig(); got != "gpt-4o" {
		t.Errorf("설정 파일 모델 미반영: %q", got)
	}
}

func TestEffectiveConfig_EnvOverridesFile(t *testing.T) {
	writeConfigFile(t, &pkgconfig.Config{
		AI: pkgconfig.AIConfig{Provider: "openai", Model: "gpt-4o"},
	})
	setEnv(t, "PCSC_AI_PROVIDER", "ollama")
	setEnv(t, "PCSC_AI_MODEL", "llama3.1")

	if got := ProviderFromConfig(); got != "ollama" {
		t.Errorf("환경변수가 설정 파일보다 우선해야 합니다: %q", got)
	}
	if got := ModelWithConfig(); got != "llama3.1" {
		t.Errorf("환경변수 모델이 우선해야 합니다: %q", got)
	}
}

func TestEffectiveConfig_NoFileNoEnv_Defaults(t *testing.T) {
	// 빈 홈 디렉터리 — 설정 파일 없음
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	setEnv(t, "PCSC_AI_PROVIDER", "")
	setEnv(t, "PCSC_AI_MODEL", "")

	if got := ProviderFromConfig(); got != "ollama" {
		t.Errorf("기본 프로바이더는 ollama여야 합니다: %q", got)
	}
	if got := ModelWithConfig(); got != "qwen3.4b" {
		t.Errorf("기본 모델은 qwen3.4b여야 합니다: %q", got)
	}
}

func TestNewClientFromConfigFile_OllamaBaseURL(t *testing.T) {
	writeConfigFile(t, &pkgconfig.Config{
		AI: pkgconfig.AIConfig{
			Provider:      "ollama",
			Model:         "qwen3.4b",
			OllamaBaseURL: "http://192.168.1.5:11434/v1",
		},
	})
	setEnv(t, "PCSC_AI_PROVIDER", "")
	setEnv(t, "OLLAMA_BASE_URL", "")

	client, err := NewClientFromConfig()
	if err != nil {
		t.Fatalf("설정 파일 기반 클라이언트 생성 실패: %v", err)
	}
	if client == nil {
		t.Fatal("클라이언트가 nil입니다")
	}
}

func TestOllamaURLOverride_EnvWins(t *testing.T) {
	writeConfigFile(t, &pkgconfig.Config{
		AI: pkgconfig.AIConfig{OllamaBaseURL: "http://192.168.1.5:11434/v1"},
	})
	setEnv(t, "PCSC_AI_PROVIDER", "")
	setEnv(t, "OLLAMA_BASE_URL", "http://env-host:11434/v1")

	url := ollamaURLFromConfig()
	if url != "http://env-host:11434/v1" {
		t.Errorf("OLLAMA_BASE_URL 환경변수가 우선해야 합니다: %q", url)
	}
}
