package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/config"
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// RenderNetworkView prints a detailed network interfaces report.
func RenderNetworkView(w io.Writer, ifaces []model.NetworkInterface, cfg *config.Config) {
	th := NewTheme(cfg.Colors)

	fmt.Fprintln(w, th.Bold(th.Cyan("NETWORK INTERFACES")))
	fmt.Fprintln(w, th.Gray("─────────────────────────────────────────────────────────────────────────────"))

	if len(ifaces) == 0 {
		fmt.Fprintln(w, "No network interfaces found.")
		return
	}

	for i, iface := range ifaces {
		stateStr := strings.ToUpper(iface.State)
		if iface.State == "up" {
			stateStr = th.Green("UP")
		} else {
			stateStr = th.Gray("DOWN")
		}

		fmt.Fprintf(w, "%s %s  [%s]\n", th.Bullet(), th.Bold(iface.Name), stateStr)

		if iface.MAC != "" {
			fmt.Fprintf(w, "   %-12s %s\n", "MAC:", iface.MAC)
		}
		if len(iface.IPv4) > 0 {
			fmt.Fprintf(w, "   %-12s %s\n", "IPv4:", strings.Join(iface.IPv4, ", "))
		}
		if len(iface.IPv6) > 0 {
			fmt.Fprintf(w, "   %-12s %s\n", "IPv6:", strings.Join(iface.IPv6, ", "))
		}
		if iface.SpeedMbps > 0 {
			fmt.Fprintf(w, "   %-12s %s\n", "Speed:", FormatSpeed(iface.SpeedMbps))
		}

		if iface.RXBytes > 0 || iface.TXBytes > 0 {
			fmt.Fprintf(w, "   %-12s %s (%d pkts", "RX:", FormatBytes(iface.RXBytes), iface.RXPackets)
			if iface.RXErrors > 0 {
				fmt.Fprintf(w, ", %s errs", th.Red(fmt.Sprintf("%d", iface.RXErrors)))
			}
			fmt.Fprintln(w, ")")

			fmt.Fprintf(w, "   %-12s %s (%d pkts", "TX:", FormatBytes(iface.TXBytes), iface.TXPackets)
			if iface.TXErrors > 0 {
				fmt.Fprintf(w, ", %s errs", th.Red(fmt.Sprintf("%d", iface.TXErrors)))
			}
			fmt.Fprintln(w, ")")
		}

		if i < len(ifaces)-1 {
			fmt.Fprintln(w)
		}
	}
}
