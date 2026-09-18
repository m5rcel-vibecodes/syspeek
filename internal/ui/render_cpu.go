package ui

import (
	"fmt"
	"io"

	"github.com/m5rcel-vibecodes/syspeek/internal/config"
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// RenderCPUView prints a detailed CPU diagnostic report.
func RenderCPUView(w io.Writer, cpu *model.CPUInfo, cfg *config.Config) {
	th := NewTheme(cfg.Colors)

	fmt.Fprintln(w, th.Bold(th.Cyan("CPU INFORMATION")))
	fmt.Fprintln(w, th.Gray("──────────────────────────────────────────────────"))

	fmt.Fprintf(w, "%-16s %s\n", "Model:", cpu.Model)
	if cpu.Vendor != "" {
		fmt.Fprintf(w, "%-16s %s\n", "Vendor:", cpu.Vendor)
	}

	if cpu.PhysicalCores > 0 {
		fmt.Fprintf(w, "%-16s %d physical, %d logical\n", "Cores:", cpu.PhysicalCores, cpu.LogicalCores)
	} else {
		fmt.Fprintf(w, "%-16s %d logical\n", "Cores:", cpu.LogicalCores)
	}

	if cpu.FrequencyMHz > 0 {
		if cpu.FrequencyMHz >= 1000 {
			fmt.Fprintf(w, "%-16s %.2f GHz\n", "Frequency:", cpu.FrequencyMHz/1000.0)
		} else {
			fmt.Fprintf(w, "%-16s %.0f MHz\n", "Frequency:", cpu.FrequencyMHz)
		}
	} else {
		fmt.Fprintf(w, "%-16s %s\n", "Frequency:", "N/A")
	}

	bar := UsageBar(cpu.UsagePercent, 20, th)
	fmt.Fprintf(w, "%-16s %s %s\n", "CPU Usage:", bar, th.UsageColor(cpu.UsagePercent, FormatPercent(cpu.UsagePercent)))

	fmt.Fprintf(w, "%-16s %.2f, %.2f, %.2f\n", "Load Averages:", cpu.LoadAverages[0], cpu.LoadAverages[1], cpu.LoadAverages[2])

	if cpu.TemperatureC > 0 {
		fmt.Fprintf(w, "%-16s %s\n", "Temperature:", FormatTemp(cpu.TemperatureC, cfg.TemperatureUnit))
	} else {
		fmt.Fprintf(w, "%-16s %s\n", "Temperature:", "N/A (sensor unavailable)")
	}
}
