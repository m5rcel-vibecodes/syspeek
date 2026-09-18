package ui

import (
	"fmt"
	"io"

	"github.com/m5rcel-vibecodes/syspeek/internal/config"
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// RenderProcessesView prints a process table.
func RenderProcessesView(w io.Writer, procs []model.ProcessInfo, cfg *config.Config) {
	th := NewTheme(cfg.Colors)

	fmt.Fprintln(w, th.Bold(th.Cyan("PROCESS OVERVIEW")))
	fmt.Fprintln(w, th.Gray("─────────────────────────────────────────────────────────────────────────────"))

	if len(procs) == 0 {
		fmt.Fprintln(w, "No process metrics available.")
		return
	}

	fmt.Fprintf(w, "%-7s %-12s %-8s %-8s %-10s %s\n",
		"PID", "USER", "%CPU", "%MEM", "RES", "COMMAND",
	)
	fmt.Fprintln(w, th.Gray("─────────────────────────────────────────────────────────────────────────────"))

	for _, p := range procs {
		user := p.User
		if len(user) > 11 {
			user = user[:10] + "…"
		}

		cmd := p.Command
		if len(cmd) > 45 {
			cmd = cmd[:42] + "..."
		}

		cpuStr := fmt.Sprintf("%.1f%%", p.CPUPercent)
		memStr := fmt.Sprintf("%.1f%%", p.MemoryPercent)

		fmt.Fprintf(w, "%-7d %-12s %-8s %-8s %-10s %s\n",
			p.PID,
			user,
			th.UsageColor(p.CPUPercent, cpuStr),
			th.UsageColor(p.MemoryPercent, memStr),
			FormatBytes(p.MemoryBytes),
			cmd,
		)
	}
}
