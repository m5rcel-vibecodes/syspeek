//go:build !linux && !darwin

package collector

import (
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectDisks is the fallback disk collector.
func CollectDisks() ([]model.DiskPartition, error) {
	return []model.DiskPartition{}, nil
}
