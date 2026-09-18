package ui

import (
	"fmt"
	"io"

	"github.com/m5rcel-vibecodes/syspeek/internal/config"
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// RenderGPUView prints a detailed GPU report.
func RenderGPUView(w io.Writer, gpus []model.GPUInfo, cfg *config.Config) {
	th := NewTheme(cfg.Colors)

	fmt.Fprintln(w, th.Bold(th.Cyan("GRAPHICS ACCELERATORS")))
	fmt.Fprintln(w, th.Gray("──────────────────────────────────────────────────"))

	if len(gpus) == 0 {
		fmt.Fprintln(w, "No GPU or graphics accelerator detected.")
		return
	}

	for i, gpu := range gpus {
		fmt.Fprintf(w, "%-16s %s\n", "Name:", gpu.Name)
		if gpu.Vendor != "" {
			fmt.Fprintf(w, "%-16s %s\n", "Vendor:", gpu.Vendor)
		}
		if gpu.Driver != "" {
			fmt.Fprintf(w, "%-16s %s\n", "Driver:", gpu.Driver)
		}
		if gpu.VRAMTotalBytes > 0 {
			vPct := float64(gpu.VRAMUsedBytes) / float64(gpu.VRAMTotalBytes) * 100.0
			bar := UsageBar(vPct, 15, th)
			fmt.Fprintf(w, "%-16s %s / %s %s\n", "VRAM:", FormatBytes(gpu.VRAMUsedBytes), FormatBytes(gpu.VRAMTotalBytes), bar)
		}
		if gpu.UtilizationPercent > 0 {
			bar := UsageBar(gpu.UtilizationPercent, 15, th)
			fmt.Fprintf(w, "%-16s %s %s\n", "Utilization:", bar, FormatPercent(gpu.UtilizationPercent))
		}
		if gpu.TemperatureC > 0 {
			fmt.Fprintf(w, "%-16s %s\n", "Temperature:", FormatTemp(gpu.TemperatureC, cfg.TemperatureUnit))
		}

		if i < len(gpus)-1 {
			fmt.Fprintln(w, th.Gray("──────────────────────────────────────────────────"))
		}
	}
}
