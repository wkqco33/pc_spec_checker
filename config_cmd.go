package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wkqco33/wcli"

	"wkqcosoft.com/m/config"
)

// newConfigCommand는 pcsc config 서브커맨드 그룹을 반환합니다
func newConfigCommand(root *wcli.Command) *wcli.Command {
	initCmd := &wcli.Command{
		Use:   "init",
		Short: "기본 설정 파일을 생성합니다",
		Run: func(ctx *wcli.Context) error {
			return runConfigInit(writerOf(root))
		},
	}

	showCmd := &wcli.Command{
		Use:   "show",
		Short: "현재 설정을 표시합니다",
		Run: func(ctx *wcli.Context) error {
			return runConfigShow(writerOf(root))
		},
	}

	setCmd := &wcli.Command{
		Use:   "set <key> <value>",
		Short: "설정 값을 변경합니다 (예: pcsc config set ai.model llama3.1)",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) != 2 {
				return fmt.Errorf("키와 값을 모두 지정해야 합니다 (예: pcsc config set ai.model llama3.1)")
			}
			return runConfigSet(writerOf(root), ctx.Args[0], ctx.Args[1])
		},
	}

	configCmd := &wcli.Command{
		Use:   "config",
		Short: "설정 파일을 관리합니다 (init/show/set)",
	}

	configCmd.AddCommand(initCmd, showCmd, setCmd)
	return configCmd
}

// writerOf는 루트 커맨드의 출력 대상을 반환합니다
func writerOf(root *wcli.Command) io.Writer {
	if root == nil || root.OutWriter == nil {
		return os.Stdout
	}
	return root.OutWriter
}

// setSupportedKeys는 config set으로 변경 가능한 키와 적용 함수입니다
var setSupportedKeys = map[string]func(*config.Config, string) error{
	"ai.provider": func(c *config.Config, v string) error {
		switch v {
		case "ollama", "openai", "azure":
			c.AI.Provider = v
			return nil
		default:
			return fmt.Errorf("지원하지 않는 프로바이더입니다: %s (ollama/openai/azure)", v)
		}
	},
	"ai.model": func(c *config.Config, v string) error {
		c.AI.Model = v
		return nil
	},
	"ai.ollama_base_url": func(c *config.Config, v string) error {
		c.AI.OllamaBaseURL = v
		return nil
	},
}

// runConfigInit은 기본 설정 파일을 생성합니다
func runConfigInit(w io.Writer) error {
	path := config.DefaultPath()

	if _, err := os.Stat(path); err == nil {
		_, _ = fmt.Fprintln(w, "설정 파일이 이미 존재합니다: "+path)
		return nil
	}

	if err := config.Save(config.Default()); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(w, "설정 파일이 생성되었습니다: "+path)
	return nil
}

// runConfigShow는 현재 설정을 출력합니다
func runConfigShow(w io.Writer) error {
	c, err := config.Load()
	if err != nil {
		return err
	}
	_, _ = fmt.Fprint(w, formatConfig(c))
	return nil
}

// runConfigSet은 설정 키의 값을 저장합니다
func runConfigSet(w io.Writer, key, value string) error {
	setter, ok := setSupportedKeys[key]
	if !ok {
		keys := make([]string, 0, len(setSupportedKeys))
		for k := range setSupportedKeys {
			keys = append(keys, k)
		}
		return fmt.Errorf("알 수 없는 설정 키입니다: %s (가능한 키: %s)", key, strings.Join(keys, ", "))
	}

	c, err := config.Load()
	if err != nil {
		return err
	}

	if err := setter(c, value); err != nil {
		return err
	}

	if err := config.Save(c); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(w, "설정이 저장되었습니다: %s = %s\n", key, value)
	return nil
}

// formatConfig는 설정을 사람이 읽을 수 있는 형태로 포맷팅합니다
func formatConfig(c *config.Config) string {
	var b strings.Builder
	b.WriteString("┌─ pcsc 설정 (" + config.DefaultPath() + ")\n")
	fmt.Fprintf(&b, "│  AI 프로바이더 : %s\n", c.AI.Provider)
	fmt.Fprintf(&b, "│  AI 모델       : %s\n", c.AI.Model)
	if c.AI.OllamaBaseURL != "" {
		fmt.Fprintf(&b, "│  Ollama URL    : %s\n", c.AI.OllamaBaseURL)
	}
	b.WriteString("└─────────────────────────────────────────────────────────────────\n")
	return b.String()
}
