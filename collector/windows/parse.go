package windows

import (
	"strconv"
	"strings"
)

// cpuRaw는 CPU 수집 중간 결과입니다
type cpuRaw struct {
	Model   string
	Cores   int
	Threads int
	MaxFreq int
}

// valid는 필수 값(CPU 모델명)이 있는지 확인합니다
func (r *cpuRaw) valid() bool {
	return r.Model != ""
}

// parseValueLine은 wmic /value 또는 PowerShell CIM 목록 출력에서 키 값을 추출합니다
func parseValueLine(output, key string) string {
	sep := strings.ReplaceAll(output, "\r\n", "\n")
	for _, line := range strings.Split(sep, "\n") {
		line = strings.TrimSpace(line)
		// wmic 형식: Key=Value
		if v, ok := strings.CutPrefix(line, key+"="); ok {
			return strings.TrimSpace(v)
		}
		// PowerShell 형식: Key : Value
		if idx := strings.Index(line, " : "); idx > 0 && strings.TrimSpace(line[:idx]) == key {
			return strings.TrimSpace(line[idx+3:])
		}
	}
	return ""
}

// parseCSVRows는 CSV 출력에서 헤더를 제외한 행을 필드 배열로 반환합니다 (CRLF 처리)
func parseCSVRows(output string) [][]string {
	sep := strings.ReplaceAll(output, "\r\n", "\n")
	var rows [][]string
	for i, line := range strings.Split(sep, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if i == 0 {
			continue // 헤더
		}
		fields := strings.Split(line, ",")
		trimmed := make([]string, len(fields))
		for j, f := range fields {
			trimmed[j] = strings.TrimSpace(f)
		}
		rows = append(rows, trimmed)
	}
	return rows
}

// parseInt는 문자열을 정수로 변환하며 실패 시 0을 반환합니다
func parseInt(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}

// parseStorageCSV는 wmic CSV 출력(Node,DeviceID,FileSystem,FreeSpace,Size)을 StorageInfo 목록으로 변환합니다
func parseStorageCSV(output string) []storageItem {
	var items []storageItem
	for _, row := range parseCSVRows(output) {
		if len(row) < 5 {
			continue
		}
		// Node,DeviceID,FileSystem,FreeSpace,Size 순서
		deviceID := row[1]
		fileSystem := row[2]
		freeStr := row[3]
		sizeStr := row[4]

		if deviceID == "" || sizeStr == "" {
			continue
		}
		totalBytes, err1 := strconv.ParseUint(sizeStr, 10, 64)
		freeBytes, err2 := strconv.ParseUint(freeStr, 10, 64)
		if err1 != nil || err2 != nil || totalBytes == 0 {
			continue
		}
		items = append(items, storageItem{
			Device:     deviceID,
			Type:       fileSystem,
			TotalBytes: totalBytes,
			FreeBytes:  freeBytes,
		})
	}
	return items
}

// storageItem은 저장장치 수집 중간 결과입니다
type storageItem struct {
	Device     string
	Type       string
	TotalBytes uint64
	FreeBytes  uint64
}
