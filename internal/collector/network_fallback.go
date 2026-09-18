//go:build !linux && !darwin

package collector

import (
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectNetwork is the fallback network collector.
func CollectNetwork() ([]model.NetworkInterface, error) {
	return []model.NetworkInterface{}, nil
}
