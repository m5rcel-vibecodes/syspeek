package collector

import (
	"fmt"
	"time"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectSnapshot gathers a complete diagnostic snapshot across all subsystems.
func CollectSnapshot() (*model.Snapshot, error) {
	snap := &model.Snapshot{
		Timestamp: time.Now().Unix(),
	}

	// Collect each subsystem independently; none should ever fail the whole snapshot.
	if sys, err := CollectSystem(); err == nil && sys != nil {
		snap.System = *sys
	} else {
		snap.System = model.SystemInfo{OS: "Unknown", Hostname: "Unknown"}
	}

	if cpu, err := CollectCPU(); err == nil && cpu != nil {
		snap.CPU = *cpu
	}

	if mem, err := CollectMemory(); err == nil && mem != nil {
		snap.Memory = *mem
	}

	if disks, err := CollectDisks(); err == nil && disks != nil {
		snap.Disks = disks
	}

	if gpus, err := CollectGPUs(); err == nil && gpus != nil {
		snap.GPUs = gpus
	}

	if netIfaces, err := CollectNetwork(); err == nil && netIfaces != nil {
		snap.Network = netIfaces
	}

	if services, err := CollectServices(); err == nil && services != nil {
		snap.Services = *services
	}

	if procs, err := CollectProcesses(15); err == nil && procs != nil {
		snap.Processes = procs
	}

	return snap, nil
}

// FormatUptime formats uptime in seconds to human-readable string like "3d 14h 21m".
func FormatUptime(seconds uint64) string {
	if seconds == 0 {
		return "0m"
	}
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
