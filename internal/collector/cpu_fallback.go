//go:build !linux && !darwin

package collector

import (
	"runtime"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectCPU is the fallback CPU collector for unsupported platforms.
func CollectCPU() (*model.CPUInfo, error) {
	cores := runtime.NumCPU()
	return &model.CPUInfo{
		Model:         "Generic CPU",
		Vendor:        "Unknown",
		PhysicalCores: cores,
		LogicalCores:  cores,
		FrequencyMHz:  0,
		UsagePercent:  0,
		LoadAverages:  [3]float64{0, 0, 0},
		TemperatureC:  0,
	}, nil
}
