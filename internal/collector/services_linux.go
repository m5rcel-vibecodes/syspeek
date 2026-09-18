//go:build linux

package collector

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectServices inspects service manager status on Linux.
func CollectServices() (*model.ServiceSummary, error) {
	// 1. Check systemd
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		if _, err := exec.LookPath("systemctl"); err == nil {
			return collectSystemdServices()
		}
	}

	// 2. Check OpenRC
	if _, err := os.Stat("/run/openrc"); err == nil {
		if _, err := exec.LookPath("rc-status"); err == nil {
			return collectOpenRCServices()
		}
	}

	// 3. Check sysvinit
	if _, err := os.Stat("/etc/init.d"); err == nil {
		if _, err := exec.LookPath("service"); err == nil {
			return collectSysvinitServices()
		}
	}

	return &model.ServiceSummary{
		Manager: "none",
		Running: 0,
		Failed:  0,
		Total:   0,
		Items:   []model.ServiceItem{},
	}, nil
}

func collectSystemdServices() (*model.ServiceSummary, error) {
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--plain")
	out, err := cmd.Output()
	if err != nil {
		// Try without --all
		cmd = exec.Command("systemctl", "list-units", "--type=service", "--no-pager", "--plain")
		out, err = cmd.Output()
		if err != nil {
			return nil, err
		}
	}

	summary := &model.ServiceSummary{
		Manager: "systemd",
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "UNIT") || strings.HasPrefix(line, "LOAD") {
			continue
		}
		if strings.HasPrefix(line, "To show all installed") || strings.HasPrefix(line, "Legend:") {
			break
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		unit := fields[0]
		if !strings.HasSuffix(unit, ".service") {
			continue
		}
		name := strings.TrimSuffix(unit, ".service")

		load := fields[1]
		active := fields[2]
		sub := fields[3]

		desc := ""
		if len(fields) >= 5 {
			desc = strings.Join(fields[4:], " ")
		}

		status := sub
		isActive := (active == "active")
		isFailed := (active == "failed" || sub == "failed")

		if isActive && (sub == "running" || sub == "exited") {
			summary.Running++
		}
		if isFailed {
			summary.Failed++
			status = "failed"
		}

		summary.Total++
		summary.Items = append(summary.Items, model.ServiceItem{
			Name:        name,
			Status:      status,
			Active:      isActive,
			Enabled:     load == "loaded",
			Description: desc,
		})
	}

	return summary, nil
}

func collectOpenRCServices() (*model.ServiceSummary, error) {
	cmd := exec.Command("rc-status", "-a")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	summary := &model.ServiceSummary{
		Manager: "openrc",
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Runlevel:") || strings.HasPrefix(line, "Dynamic:") {
			continue
		}

		// Example OpenRC line: " sshd                               [  started  ]"
		bracketOpen := strings.LastIndex(line, "[")
		bracketClose := strings.LastIndex(line, "]")
		if bracketOpen == -1 || bracketClose == -1 || bracketClose <= bracketOpen {
			continue
		}

		name := strings.TrimSpace(line[:bracketOpen])
		status := strings.ToLower(strings.TrimSpace(line[bracketOpen+1 : bracketClose]))

		isActive := (status == "started")
		isFailed := (status == "crashed")

		if isActive {
			summary.Running++
		}
		if isFailed {
			summary.Failed++
		}

		summary.Total++
		summary.Items = append(summary.Items, model.ServiceItem{
			Name:   name,
			Status: status,
			Active: isActive,
		})
	}

	return summary, nil
}

func collectSysvinitServices() (*model.ServiceSummary, error) {
	cmd := exec.Command("service", "--status-all")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	summary := &model.ServiceSummary{
		Manager: "sysvinit",
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Example: " [ + ]  apache2", " [ - ]  cron"
		if len(line) < 7 || line[0] != '[' {
			continue
		}

		marker := line[3]
		name := strings.TrimSpace(line[5:])
		status := "stopped"
		isActive := false

		if marker == '+' {
			status = "running"
			isActive = true
			summary.Running++
		} else if marker == '-' {
			status = "stopped"
		} else if marker == '?' {
			status = "unknown"
		}

		summary.Total++
		summary.Items = append(summary.Items, model.ServiceItem{
			Name:   name,
			Status: status,
			Active: isActive,
		})
	}

	return summary, nil
}
