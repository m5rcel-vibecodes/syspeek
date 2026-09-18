package ui

import (
	"fmt"
	"io"

	"github.com/m5rcel-vibecodes/syspeek/internal/config"
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// RenderMemoryView prints a detailed memory diagnostic report.
func RenderMemoryView(w io.Writer, mem *model.MemoryInfo, cfg *config.Config) {
	th := NewTheme(cfg.Colors)

	fmt.Fprintln(w, th.Bold(th.Cyan("MEMORY INFORMATION")))
	fmt.Fprintln(w, th.Gray("──────────────────────────────────────────────────"))

	if mem.TotalBytes == 0 {
		fmt.Fprintln(w, "Memory statistics unavailable.")
		return
	}

	bar := UsageBar(mem.UsagePercent, 20, th)
	fmt.Fprintf(w, "%-16s %s %s\n", "RAM Usage:", bar, th.UsageColor(mem.UsagePercent, FormatPercent(mem.UsagePercent)))
	fmt.Fprintf(w, "%-16s %s / %s\n", "Used / Total:", FormatBytes(mem.UsedBytes), FormatBytes(mem.TotalBytes))
	fmt.Fprintf(w, "%-16s %s\n", "Available:", FormatBytes(mem.AvailableBytes))
	fmt.Fprintf(w, "%-16s %s\n", "Free:", FormatBytes(mem.FreeBytes))

	if mem.CachedBytes > 0 {
		fmt.Fprintf(w, "%-16s %s\n", "Cached:", FormatBytes(mem.CachedBytes))
	}
	if mem.BuffersBytes > 0 {
		fmt.Fprintf(w, "%-16s %s\n", "Buffers:", FormatBytes(mem.BuffersBytes))
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, th.Bold(th.Cyan("SWAP INFORMATION")))
	fmt.Fprintln(w, th.Gray("──────────────────────────────────────────────────"))

	if mem.SwapTotalBytes > 0 {
		swapBar := UsageBar(mem.SwapPercent, 20, th)
		fmt.Fprintf(w, "%-16s %s %s\n", "Swap Usage:", swapBar, th.UsageColor(mem.SwapPercent, FormatPercent(mem.SwapPercent)))
		fmt.Fprintf(w, "%-16s %s / %s\n", "Used / Total:", FormatBytes(mem.SwapUsedBytes), FormatBytes(mem.SwapTotalBytes))
		fmt.Fprintf(w, "%-16s %s\n", "Free:", FormatBytes(mem.SwapFreeBytes))
	} else {
		fmt.Fprintf(w, "%-16s %s\n", "Swap:", "None / Disabled")
	}
}
