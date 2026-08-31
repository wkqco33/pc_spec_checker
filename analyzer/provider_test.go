package analyzer

import (
	"os"
	"testing"
)

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	old, had := os.LookupEnv(key)
	_ = os.Setenv(key, value)
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func TestNewClientFromEnv_DefaultOllama(t *testing.T) {
	setEnv(t, "PCSC_AI_PROVIDER", "")
	setEnv(t, "OLLAMA_BASE_URL", "")

	client, err := NewClientFromEnv()
	if err != nil {
		t.Fatalf("기본 프로바이더 생성 실패: %v", err)
	}
	if client == nil {
		t.Fatal("클라이언트가 nil입니다")
	}
}

func TestNewClientFromEnv_OpenAI(t *testing.T) {
	setEnv(t, "PCSC_AI_PROVIDER", "openai")
	setEnv(t, "OPENAI_API_KEY", "test-key")

	client, err := NewClientFromEnv()
	if err != nil {
		t.Fatalf("OpenAI 클라이언트 생성 실패: %v", err)
	}
	if client == nil {
		t.Fatal("클라이언트가 nil입니다")
	}
}

func TestNewClientFromEnv_OpenAI_MissingKey(t *testing.T) {
	setEnv(t, "PCSC_AI_PROVIDER", "openai")
	_ = os.Unsetenv("OPENAI_API_KEY")

	if _, err := NewClientFromEnv(); err == nil {
		t.Error("API 키가 없으면 에러가 반환되어야 합니다")
	}
}

func TestNewClientFromEnv_Azure_MissingConfig(t *testing.T) {
	setEnv(t, "PCSC_AI_PROVIDER", "azure")
	_ = os.Unsetenv("AZURE_API_KEY")
	_ = os.Unsetenv("AZURE_ENDPOINT")

	if _, err := NewClientFromEnv(); err == nil {
		t.Error("Azure 설정이 없으면 에러가 반환되어야 합니다")
	}
}

func TestNewClientFromEnv_UnknownProvider(t *testing.T) {
	setEnv(t, "PCSC_AI_PROVIDER", "gemini")

	if _, err := NewClientFromEnv(); err == nil {
		t.Error("알 수 없는 프로바이더에서 에러가 반환되어야 합니다")
	}
}

func TestModelFromEnv_Default(t *testing.T) {
	setEnv(t, "PCSC_AI_MODEL", "")
	setEnv(t, "PCSC_AI_PROVIDER", "")

	if got := ModelFromEnv(); got != "qwen3.4b" {
		t.Errorf("기본 모델은 qwen3.4b여야 합니다: %q", got)
	}
}

func TestModelFromEnv_Override(t *testing.T) {
	setEnv(t, "PCSC_AI_MODEL", "gpt-4o-mini")

	if got := ModelFromEnv(); got != "gpt-4o-mini" {
		t.Errorf("모델 오버라이드 실패: %q", got)
	}
}
