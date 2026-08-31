// Package config는 pcsc의 설정 파일(~/.config/pcsc/config.json) 관리를 담당합니다
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config는 pcsc 설정 파일의 최상위 구조입니다
type Config struct {
	AI AIConfig `json:"ai"`
}

// AIConfig는 AI 분석 관련 설정입니다
type AIConfig struct {
	Provider      string `json:"provider"`        // ollama / openai / azure
	Model         string `json:"model"`           // 모델명
	OllamaBaseURL string `json:"ollama_base_url"` // Ollama 엔드포인트 (비어있으면 라이브러리 기본값)
}

// Default는 기본 설정을 반환합니다
func Default() *Config {
	return &Config{
		AI: AIConfig{
			Provider: "ollama",
			Model:    "qwen3.4b",
		},
	}
}

// DefaultPath는 설정 파일 경로를 반환합니다 ($XDG_CONFIG_HOME/pcsc/config.json)
func DefaultPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "pcsc", "config.json")
}

// Load는 설정 파일을 읽습니다. 파일이 없으면 기본값을 반환합니다
func Load() (*Config, error) {
	c := Default()
	path := DefaultPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, fmt.Errorf("설정 파일을 읽을 수 없습니다: %w", err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("설정 파일 파싱 실패 (%s): %w", path, err)
	}
	return c, nil
}

// Save는 설정 파일을 저장합니다. 디렉터리가 없으면 생성합니다
func Save(c *Config) error {
	path := DefaultPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("설정 디렉터리 생성 실패: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("설정 직렬화 실패: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("설정 파일 저장 실패: %w", err)
	}
	return nil
}
