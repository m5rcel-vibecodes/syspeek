//go:build darwin

package collector

import (
	"bufio"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectProcesses gathers processes on macOS.
func CollectProcesses(limit int) ([]model.ProcessInfo, error) {
	cmd := exec.Command("ps", "-axo", "pid,user,%cpu,%mem,rss,stat,comm")
	out, err := cmd.Output()
	if err != nil {
		return []model.ProcessInfo{}, nil
	}

	var procs []model.ProcessInfo
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	// Skip header line
	if scanner.Scan() {
		_ = scanner.Text()
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		pid, _ := strconv.Atoi(fields[0])
		user := fields[1]
		cpuPct, _ := strconv.ParseFloat(fields[2], 64)
		memPct, _ := strconv.ParseFloat(fields[3], 64)
		rssKB, _ := strconv.ParseUint(fields[4], 10, 64)
		stat := fields[5]
		comm := strings.Join(fields[6:], " ")

		procs = append(procs, model.ProcessInfo{
			PID:           pid,
			User:          user,
			CPUPercent:    cpuPct,
			MemoryBytes:   rssKB * 1024,
			MemoryPercent: memPct,
			Command:       comm,
			Status:        stat,
		})
	}

	// Sort by CPU% descending, then MemoryBytes
	sort.Slice(procs, func(i, j int) bool {
		if procs[i].CPUPercent != procs[j].CPUPercent {
			return procs[i].CPUPercent > procs[j].CPUPercent
		}
		return procs[i].MemoryBytes > procs[j].MemoryBytes
	})

	if limit > 0 && len(procs) > limit {
		procs = procs[:limit]
	}

	return procs, nil
}
