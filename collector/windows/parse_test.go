package windows

import "testing"

func TestParseCPUValueOutput(t *testing.T) {
	output := "\r\nName=Intel(R) Core(TM) i7-8850H CPU @ 2.60GHz\r\n\r\nNumberOfCores=6\r\n"
	if got := parseValueLine(output, "Name"); got != "Intel(R) Core(TM) i7-8850H CPU @ 2.60GHz" {
		t.Errorf("Name 파싱 실패: %q", got)
	}
	if got := parseValueLine(output, "NumberOfCores"); got != "6" {
		t.Errorf("NumberOfCores 파싱 실패: %q", got)
	}
	if got := parseValueLine(output, "NonExistent"); got != "" {
		t.Errorf("없는 키는 빈 값을 반환해야 합니다: %q", got)
	}
}

func TestParseCPUValueOutput_PowerShellListFormat(t *testing.T) {
	// PowerShell CIM 출력: Key : Value 형태
	output := "\r\nName      : AMD Ryzen 5 5600X\r\nNumberOfCores : 6\r\n"
	if got := parseValueLine(output, "Name"); got != "AMD Ryzen 5 5600X" {
		t.Errorf("PowerShell 형식 Name 파싱 실패: %q", got)
	}
	if got := parseValueLine(output, "NumberOfCores"); got != "6" {
		t.Errorf("PowerShell 형식 파싱 실패: %q", got)
	}
}

func TestParseCSVRows_HandlesCRLF(t *testing.T) {
	csv := "Node,DeviceID,FileSystem,FreeSpace,Size\r\nDESKTOP,C:,NTFS,100,500\r\n"
	rows := parseCSVRows(csv)
	if len(rows) != 1 {
		t.Fatalf("행 개수 오류: %d", len(rows))
	}
	if rows[0][1] != "C:" {
		t.Errorf("DeviceID 파싱 실패: %v", rows[0])
	}
}

func TestParseInt(t *testing.T) {
	if got := parseInt("42"); got != 42 {
		t.Errorf("parseInt 실패: %d", got)
	}
	if got := parseInt(""); got != 0 {
		t.Errorf("빈 값은 0이어야 합니다: %d", got)
	}
	if got := parseInt("abc"); got != 0 {
		t.Errorf("숫자가 아니면 0이어야 합니다: %d", got)
	}
}

func TestMergeCPUInfo(t *testing.T) {
	base := &cpuRaw{Model: "Test CPU", Threads: 12, Cores: 6}
	if base.Model != "Test CPU" || base.Cores != 6 || base.Threads != 12 {
		t.Errorf("cpuRaw 필드 검증 실패: %+v", base)
	}
}

func TestCPUInfoValid(t *testing.T) {
	if (&cpuRaw{Model: "X"}).valid() != true {
		t.Error("모델이 있으면 유효해야 합니다")
	}
	if (&cpuRaw{Model: ""}).valid() != false {
		t.Error("모델이 없으면 무효해야 합니다")
	}
}
