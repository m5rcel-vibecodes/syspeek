package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/config"
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// RenderServicesView prints a detailed services report.
func RenderServicesView(w io.Writer, summary *model.ServiceSummary, cfg *config.Config) {
	th := NewTheme(cfg.Colors)

	mgrName := strings.ToUpper(summary.Manager)
	if mgrName == "" || mgrName == "UNKNOWN" || mgrName == "NONE" {
		mgrName = "SYSTEM SERVICES"
	} else {
		mgrName = fmt.Sprintf("SERVICES (%s)", mgrName)
	}

	fmt.Fprintln(w, th.Bold(th.Cyan(mgrName)))
	fmt.Fprintln(w, th.Gray("─────────────────────────────────────────────────────────────────────────────"))

	fmt.Fprintf(w, "Running: %s   Failed: %s   Total Detected: %d\n",
		th.Green(fmt.Sprintf("%d", summary.Running)),
		th.Red(fmt.Sprintf("%d", summary.Failed)),
		summary.Total,
	)
	fmt.Fprintln(w, th.Gray("─────────────────────────────────────────────────────────────────────────────"))

	if len(summary.Items) == 0 {
		fmt.Fprintln(w, "No services active or service manager not detected.")
		return
	}

	fmt.Fprintf(w, "%-6s %-32s %-12s %s\n", "STAT", "SERVICE", "STATE", "DESCRIPTION")
	fmt.Fprintln(w, th.Gray("─────────────────────────────────────────────────────────────────────────────"))

	for _, item := range summary.Items {
		var icon string
		if item.Status == "failed" {
			icon = th.Cross()
		} else if item.Active {
			icon = th.Check()
		} else {
			icon = th.Bullet()
		}

		name := item.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		desc := item.Description
		if len(desc) > 35 {
			desc = desc[:32] + "..."
		}

		fmt.Fprintf(w, " %-5s %-32s %-12s %s\n",
			icon,
			name,
			item.Status,
			desc,
		)
	}
}
