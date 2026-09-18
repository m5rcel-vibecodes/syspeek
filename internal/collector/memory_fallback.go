//go:build !linux && !darwin

package collector

import (
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectMemory is the fallback memory collector.
func CollectMemory() (*model.MemoryInfo, error) {
	return &model.MemoryInfo{
		TotalBytes:     0,
		UsedBytes:      0,
		FreeBytes:      0,
		AvailableBytes: 0,
		UsagePercent:   0,
	}, nil
}
