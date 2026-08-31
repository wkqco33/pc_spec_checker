package analyzer

import "testing"

func validResult() AnalysisResult {
	return AnalysisResult{
		Summary: "전반적으로 균형 잡힌 사양입니다.",
		Verdicts: []UsageVerdict{
			{Category: CategoryWeb, Grade: GradeExcellent, Score: 90, Comment: "웹서핑에 충분합니다."},
			{Category: CategoryGame, Grade: GradeGood, Score: 70, Comment: "인디 게임은 원활합니다."},
			{Category: CategoryDev, Grade: GradeFair, Score: 50, Comment: "대형 프로젝트 빌드는 느릴 수 있습니다."},
			{Category: CategoryVideo, Grade: GradeGood, Score: 75, Comment: "4K 재생이 가능합니다."},
			{Category: CategoryOffice, Grade: GradeExcellent, Score: 95, Comment: "사무용으로 충분합니다."},
		},
		Upgrade: "메모리 추가를 권장합니다.",
	}
}

func TestValidate_ValidResult(t *testing.T) {
	r := validResult()
	if err := r.Validate(); err != nil {
		t.Fatalf("유효한 결과에서 검증 실패: %v", err)
	}
}

func TestValidate_EmptySummary(t *testing.T) {
	r := validResult()
	r.Summary = ""
	if err := r.Validate(); err == nil {
		t.Error("빈 요약에서 에러가 반환되어야 합니다")
	}
}

func TestValidate_NoVerdicts(t *testing.T) {
	r := validResult()
	r.Verdicts = nil
	if err := r.Validate(); err == nil {
		t.Error("평가 항목이 없으면 에러가 반환되어야 합니다")
	}
}

func TestValidate_InvalidCategory(t *testing.T) {
	r := validResult()
	r.Verdicts[0].Category = "crypto-mining"
	if err := r.Validate(); err == nil {
		t.Error("알 수 없는 카테고리에서 에러가 반환되어야 합니다")
	}
}

func TestValidate_InvalidGrade(t *testing.T) {
	r := validResult()
	r.Verdicts[1].Grade = "legendary"
	if err := r.Validate(); err == nil {
		t.Error("알 수 없는 등급에서 에러가 반환되어야 합니다")
	}
}

func TestValidate_ScoreOutOfRange(t *testing.T) {
	r := validResult()
	r.Verdicts[2].Score = 150
	if err := r.Validate(); err == nil {
		t.Error("점수가 100 초과하면 에러가 반환되어야 합니다")
	}

	r.Verdicts[2].Score = -1
	if err := r.Validate(); err == nil {
		t.Error("점수가 음수면 에러가 반환되어야 합니다")
	}
}

func TestAllCategoriesComplete(t *testing.T) {
	r := validResult()
	if !r.AllCategoriesComplete() {
		t.Error("5개 카테고리가 모두 있으면 true여야 합니다")
	}

	r.Verdicts = r.Verdicts[:4]
	if r.AllCategoriesComplete() {
		t.Error("카테고리가 누락되면 false여야 합니다")
	}
}

func TestGradeLabel(t *testing.T) {
	cases := map[string]string{
		GradeExcellent: "매우 좋음",
		GradeGood:      "좋음",
		GradeFair:      "보통",
		GradePoor:      "부족",
		"unknown":      "unknown",
	}
	for grade, want := range cases {
		if got := GradeLabel(grade); got != want {
			t.Errorf("GradeLabel(%q) = %q, want %q", grade, got, want)
		}
	}
}
