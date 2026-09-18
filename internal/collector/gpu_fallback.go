//go:build !linux && !darwin

package collector

import (
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectGPUs is the fallback GPU collector.
func CollectGPUs() ([]model.GPUInfo, error) {
	return []model.GPUInfo{}, nil
}
