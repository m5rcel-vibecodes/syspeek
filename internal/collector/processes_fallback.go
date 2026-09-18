//go:build !linux && !darwin

package collector

import (
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectProcesses is the fallback process collector.
func CollectProcesses(limit int) ([]model.ProcessInfo, error) {
	return []model.ProcessInfo{}, nil
}
