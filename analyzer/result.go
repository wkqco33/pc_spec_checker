package analyzer

import "fmt"

// 용도 카테고리 정의
const (
	CategoryGame   = "game"
	CategoryDev    = "dev"
	CategoryVideo  = "video"
	CategoryWeb    = "web"
	CategoryOffice = "office"
)

// 성능 등급 정의
const (
	GradeExcellent = "excellent"
	GradeGood      = "good"
	GradeFair      = "fair"
	GradePoor      = "poor"
)

// AllCategories는 분석 대상 카테고리 전체 목록입니다
var AllCategories = []string{CategoryGame, CategoryDev, CategoryVideo, CategoryWeb, CategoryOffice}

// UsageVerdict는 특정 용도에 대한 성능 평가입니다
type UsageVerdict struct {
	Category string `json:"category"` // 용도 (game/dev/video/web/office)
	Grade    string `json:"grade"`    // 등급 (excellent/good/fair/poor)
	Score    int    `json:"score"`    // 점수 (0~100)
	Comment  string `json:"comment"`  // 평가 코멘트
}

// AnalysisResult는 LLM이 생성한 사양 분석 결과입니다
type AnalysisResult struct {
	Summary  string         `json:"summary"`  // 종합 요약
	Verdicts []UsageVerdict `json:"verdicts"` // 용도별 평가
	Upgrade  string         `json:"upgrade"`  // 업그레이드 제안
}

// Validate는 결과 값의 유효성을 검사합니다
func (r *AnalysisResult) Validate() error {
	if r.Summary == "" {
		return fmt.Errorf("요약이 비어 있습니다")
	}
	if len(r.Verdicts) == 0 {
		return fmt.Errorf("용도별 평가가 없습니다")
	}
	for _, v := range r.Verdicts {
		if !isValidCategory(v.Category) {
			return fmt.Errorf("알 수 없는 용도입니다: %s", v.Category)
		}
		if !isValidGrade(v.Grade) {
			return fmt.Errorf("알 수 없는 등급입니다: %s", v.Grade)
		}
		if v.Score < 0 || v.Score > 100 {
			return fmt.Errorf("점수는 0~100 범위여야 합니다: %d (%s)", v.Score, v.Category)
		}
	}
	return nil
}

// AllCategoriesComplete는 5개 카테고리 평가가 모두 포함되었는지 확인합니다
func (r *AnalysisResult) AllCategoriesComplete() bool {
	seen := make(map[string]bool, len(r.Verdicts))
	for _, v := range r.Verdicts {
		seen[v.Category] = true
	}
	for _, c := range AllCategories {
		if !seen[c] {
			return false
		}
	}
	return true
}

// GradeLabel은 등급을 사람이 읽을 수 있는 라벨로 변환합니다
func GradeLabel(grade string) string {
	switch grade {
	case GradeExcellent:
		return "매우 좋음"
	case GradeGood:
		return "좋음"
	case GradeFair:
		return "보통"
	case GradePoor:
		return "부족"
	default:
		return grade
	}
}

func isValidCategory(c string) bool {
	for _, valid := range AllCategories {
		if c == valid {
			return true
		}
	}
	return false
}

func isValidGrade(g string) bool {
	switch g {
	case GradeExcellent, GradeGood, GradeFair, GradePoor:
		return true
	}
	return false
}
