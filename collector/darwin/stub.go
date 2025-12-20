//go:build !darwin

package darwin

import "wkqcosoft.com/m/model"

// Collector는 macOS용 collector 스텁입니다
type Collector struct{}

// New는 스텁 생성자입니다
func New() *Collector {
	return &Collector{}
}

// 인터페이스 구현 (실제로는 사용되지 않음)
func (c *Collector) CollectAll() (*model.SystemInfo, error) {
	return nil, nil
}

func (c *Collector) CollectCPU() (*model.CPUInfo, error) {
	return nil, nil
}

func (c *Collector) CollectMemory() (*model.MemoryInfo, error) {
	return nil, nil
}

func (c *Collector) CollectStorage() ([]model.StorageInfo, error) {
	return nil, nil
}

func (c *Collector) CollectGPU() ([]model.GPUInfo, error) {
	return nil, nil
}
