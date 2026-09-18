//go:build darwin

package collector

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectServices inspects services managed by launchd on macOS.
func CollectServices() (*model.ServiceSummary, error) {
	cmd := exec.Command("launchctl", "list")
	out, err := cmd.Output()
	if err != nil {
		return &model.ServiceSummary{
			Manager: "launchd",
			Running: 0,
			Failed:  0,
			Total:   0,
			Items:   []model.ServiceItem{},
		}, nil
	}

	summary := &model.ServiceSummary{
		Manager: "launchd",
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	// Skip header line (PID Status Label)
	if scanner.Scan() {
		_ = scanner.Text()
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		pidStr := fields[0]
		statusStr := fields[1]
		label := fields[2]

		isRunning := pidStr != "-"
		lastStatus, _ := strconv.Atoi(statusStr)
		isFailed := !isRunning && lastStatus != 0

		status := "inactive"
		if isRunning {
			status = "running"
			summary.Running++
		} else if isFailed {
			status = "failed"
			summary.Failed++
		}

		summary.Total++
		summary.Items = append(summary.Items, model.ServiceItem{
			Name:   label,
			Status: status,
			Active: isRunning,
		})
	}

	return summary, nil
}
