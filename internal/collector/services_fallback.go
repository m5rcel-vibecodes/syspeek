//go:build !linux && !darwin

package collector

import (
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectServices is the fallback services collector.
func CollectServices() (*model.ServiceSummary, error) {
	return &model.ServiceSummary{
		Manager: "unknown",
		Running: 0,
		Failed:  0,
		Total:   0,
		Items:   []model.ServiceItem{},
	}, nil
}
