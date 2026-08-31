package analyzer

import (
	"fmt"
	"os"
	"time"

	llm "github.com/wkqco33/LLM_client_go"
	"github.com/wkqco33/LLM_client_go/azure"
	"github.com/wkqco33/LLM_client_go/ollama"
	"github.com/wkqco33/LLM_client_go/openai"

	pkgconfig "wkqcosoft.com/m/config"
)

const (
	// EnvProvider는 LLM 프로바이더를 지정하는 환경변수입니다 (ollama/openai/azure)
	EnvProvider = "PCSC_AI_PROVIDER"
	// EnvModel은 사용할 모델명을 지정하는 환경변수입니다
	EnvModel = "PCSC_AI_MODEL"
)

const (
	defaultOllamaModel = "qwen3.4b"
	defaultOpenAIModel = "gpt-4o-mini"
	llmTimeout         = 120 * time.Second
)

// NewClientFromEnv는 환경변수 기반으로 llm.Client를 생성합니다 (기본: Ollama 로컬)
func NewClientFromEnv() (llm.Client, error) {
	switch os.Getenv(EnvProvider) {
	case "", "ollama":
		return ollama.New(ollama.Config{Timeout: llmTimeout}), nil
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY 환경변수가 설정되지 않았습니다")
		}
		return openai.New(openai.Config{APIKey: key, Timeout: llmTimeout}), nil
	case "azure":
		key := os.Getenv("AZURE_API_KEY")
		endpoint := os.Getenv("AZURE_ENDPOINT")
		if key == "" || endpoint == "" {
			return nil, fmt.Errorf("AZURE_API_KEY와 AZURE_ENDPOINT 환경변수가 필요합니다")
		}
		return azure.New(azure.Config{APIKey: key, Endpoint: endpoint, Timeout: llmTimeout}), nil
	default:
		return nil, fmt.Errorf("알 수 없는 프로바이더입니다: %s (ollama/openai/azure 지원)", os.Getenv(EnvProvider))
	}
}

// ModelFromEnv는 환경변수에서 모델명을 조회하며, 없으면 프로바이더 기본값을 반환합니다
func ModelFromEnv() string {
	if m := os.Getenv(EnvModel); m != "" {
		return m
	}
	switch os.Getenv(EnvProvider) {
	case "openai":
		return defaultOpenAIModel
	default:
		return defaultOllamaModel
	}
}

// ProviderFromConfig는 프로바이더를 결정합니다 (우선순위: env > 설정 파일 > 기본값)
func ProviderFromConfig() string {
	if p := os.Getenv(EnvProvider); p != "" {
		return p
	}
	cfg, err := pkgconfig.Load()
	if err != nil || cfg.AI.Provider != "" {
		if err == nil {
			return cfg.AI.Provider
		}
	}
	return "ollama"
}

// ModelWithConfig는 모델명을 결정합니다 (우선순위: env > 설정 파일 > 프로바이더 기본값)
func ModelWithConfig() string {
	if m := os.Getenv(EnvModel); m != "" {
		return m
	}
	cfg, err := pkgconfig.Load()
	if err == nil && cfg.AI.Model != "" {
		return cfg.AI.Model
	}
	// 설정 파일 모델이 없으면 프로바이더 기본값 (env 기준)
	return ModelFromEnv()
}

// ollamaURLFromConfig는 Ollama 엔드포인트를 결정합니다 (env > 설정 파일 > 라이브러리 기본값)
func ollamaURLFromConfig() string {
	if u := os.Getenv("OLLAMA_BASE_URL"); u != "" {
		return u
	}
	cfg, err := pkgconfig.Load()
	if err == nil && cfg.AI.OllamaBaseURL != "" {
		return cfg.AI.OllamaBaseURL
	}
	return ""
}

// NewClientFromConfig는 환경변수와 설정 파일을 반영해 llm.Client를 생성합니다
func NewClientFromConfig() (llm.Client, error) {
	switch ProviderFromConfig() {
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY 환경변수가 설정되지 않았습니다")
		}
		return openai.New(openai.Config{APIKey: key, Timeout: llmTimeout}), nil
	case "azure":
		key := os.Getenv("AZURE_API_KEY")
		endpoint := os.Getenv("AZURE_ENDPOINT")
		if key == "" || endpoint == "" {
			return nil, fmt.Errorf("AZURE_API_KEY와 AZURE_ENDPOINT 환경변수가 필요합니다")
		}
		return azure.New(azure.Config{APIKey: key, Endpoint: endpoint, Timeout: llmTimeout}), nil
	case "ollama", "":
		cfg := ollama.Config{Timeout: llmTimeout}
		if u := ollamaURLFromConfig(); u != "" {
			cfg.BaseURL = u
		}
		return ollama.New(cfg), nil
	default:
		return nil, fmt.Errorf("알 수 없는 프로바이더입니다: %s (ollama/openai/azure 지원)", ProviderFromConfig())
	}
}
