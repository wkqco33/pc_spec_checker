package analyzer

import (
	"strings"
	"testing"

	"wkqcosoft.com/m/model"
)

func sampleInfo() *model.SystemInfo {
	return &model.SystemInfo{
		CPU: model.CPUInfo{
			Model:      "AMD Ryzen 5 5600X",
			Cores:      6,
			Threads:    12,
			MaxFreqMHz: 4600,
		},
		Memory: model.MemoryInfo{
			TotalGB:     16,
			AvailableGB: 10,
			UsedGB:      6,
			UsedPercent: 37.5,
		},
		Storage: []model.StorageInfo{
			{Device: "/dev/nvme0n1", MountPoint: "/", Type: "ext4", TotalGB: 500, UsedGB: 200, FreeGB: 300, UsedPercent: 40},
		},
		GPU: []model.GPUInfo{
			{Name: "NVIDIA GeForce RTX 3060", Vendor: "NVIDIA", MemoryGB: 12, Driver: "550.54"},
		},
	}
}

func TestBuildUserPrompt_ContainsSpecs(t *testing.T) {
	prompt := BuildUserPrompt(sampleInfo())

	checks := []string{
		"AMD Ryzen 5 5600X", "6", "12", "4600",
		"16", "500", "RTX 3060", "NVIDIA",
	}
	for _, want := range checks {
		if !strings.Contains(prompt, want) {
			t.Errorf("프롬프트에 %q가 없습니다", want)
		}
	}
}

func TestBuildUserPrompt_ContainsCategories(t *testing.T) {
	prompt := BuildUserPrompt(sampleInfo())
	for _, c := range AllCategories {
		if !strings.Contains(prompt, c) {
			t.Errorf("프롬프트에 카테고리 %q가 없습니다", c)
		}
	}
}

func TestBuildSystemPrompt_MentionsJSON(t *testing.T) {
	prompt := BuildSystemPrompt()
	if !strings.Contains(prompt, "JSON") {
		t.Errorf("시스템 프롬프트에 JSON 언급이 필요합니다: %q", prompt)
	}
}

func TestResponseSchema_ValidJSON(t *testing.T) {
	schema := ResponseSchema()
	if schema == nil {
		t.Fatal("스키마가 nil입니다")
	}
	if schema["type"] != "object" {
		t.Errorf("스키마 타입이 object가 아닙니다: %v", schema["type"])
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties가 map[string]any가 아닙니다")
	}
	for _, key := range []string{"summary", "verdicts", "upgrade"} {
		if _, ok := props[key]; !ok {
			t.Errorf("스키마에 %q 속성이 없습니다", key)
		}
	}
}
