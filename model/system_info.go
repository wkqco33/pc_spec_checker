package model

// SystemInfo는 PC의 전체 사양 정보를 담는 구조체입니다
type SystemInfo struct {
	CPU     CPUInfo       `json:"cpu"`
	Memory  MemoryInfo    `json:"memory"`
	Storage []StorageInfo `json:"storage"`
	GPU     []GPUInfo     `json:"gpu"`
}

// CPUInfo는 CPU 정보를 담는 구조체입니다
type CPUInfo struct {
	Model      string `json:"model"`        // CPU 모델명
	Cores      int    `json:"cores"`        // 물리 코어 수
	Threads    int    `json:"threads"`      // 논리 코어(스레드) 수
	MaxFreqMHz int    `json:"max_freq_mhz"` // 최대 클럭 속도 (MHz)
}

// MemoryInfo는 메모리(RAM) 정보를 담는 구조체입니다
type MemoryInfo struct {
	TotalGB     float64 `json:"total_gb"`     // 전체 메모리 용량 (GB)
	AvailableGB float64 `json:"available_gb"` // 사용 가능한 메모리 (GB)
	UsedGB      float64 `json:"used_gb"`      // 사용 중인 메모리 (GB)
	UsedPercent float64 `json:"used_percent"` // 메모리 사용률 (%)
}

// StorageInfo는 저장장치 정보를 담는 구조체입니다
type StorageInfo struct {
	Device      string  `json:"device"`       // 장치 경로 (/dev/sda 등)
	MountPoint  string  `json:"mount_point"`  // 마운트 지점
	Type        string  `json:"type"`         // 파일시스템 타입
	TotalGB     float64 `json:"total_gb"`     // 전체 용량 (GB)
	UsedGB      float64 `json:"used_gb"`      // 사용 중인 용량 (GB)
	FreeGB      float64 `json:"free_gb"`      // 남은 용량 (GB)
	UsedPercent float64 `json:"used_percent"` // 사용률 (%)
}

// GPUInfo는 GPU 정보를 담는 구조체입니다
type GPUInfo struct {
	Name     string  `json:"name"`      // GPU 이름
	Vendor   string  `json:"vendor"`    // 제조사
	MemoryGB float64 `json:"memory_gb"` // GPU 메모리 (GB)
	Driver   string  `json:"driver"`    // 드라이버 버전
}
