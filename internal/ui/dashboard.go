package ui

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/m5rcel-vibecodes/syspeek/internal/config"
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// RenderDashboard prints the default unified terminal dashboard to w.
func RenderDashboard(w io.Writer, snap *model.Snapshot, cfg *config.Config) {
	th := NewTheme(cfg.Colors)
	termWidth := GetTerminalWidth()
	if termWidth < 40 {
		termWidth = 40
	}

	blankLine := func() {
		if !cfg.Compact {
			fmt.Fprintln(w)
		}
	}

	// 0. ASCII LOGO (Optional)
	if cfg.ShowASCIILogo {
		renderASCIILogo(w, th)
		blankLine()
	}

	// 1. SYSTEM BOX
	if cfg.SectionEnabled("system") {
		renderSystemBox(w, snap.System, th, termWidth)
		blankLine()
	}

	// 2. CPU SECTION
	if cfg.SectionEnabled("cpu") {
		fmt.Fprintln(w, th.Bold(th.Cyan("HARDWARE")))
		blankLine()
		renderCPUSection(w, snap.CPU, th, cfg)
		blankLine()
	}

	// 3. MEMORY SECTION
	if cfg.SectionEnabled("memory") {
		fmt.Fprintln(w, th.Bold(th.Cyan("MEMORY")))
		blankLine()
		renderMemorySection(w, snap.Memory, th)
		blankLine()
	}

	// 4. STORAGE SECTION
	if cfg.SectionEnabled("storage") || cfg.SectionEnabled("disk") {
		fmt.Fprintln(w, th.Bold(th.Cyan("STORAGE")))
		blankLine()
		renderStorageSection(w, snap.Disks, th, cfg, termWidth)
		blankLine()
	}

	// 5. GPU SECTION (if present and enabled)
	if (cfg.SectionEnabled("gpu") || len(snap.GPUs) > 0) && len(snap.GPUs) > 0 {
		fmt.Fprintln(w, th.Bold(th.Cyan("GPU")))
		blankLine()
		renderGPUSection(w, snap.GPUs, th, cfg)
		blankLine()
	}

	// 6. NETWORK SECTION
	if cfg.SectionEnabled("network") {
		fmt.Fprintln(w, th.Bold(th.Cyan("NETWORK")))
		blankLine()
		renderNetworkSection(w, snap.Network, th, cfg)
		blankLine()
	}

	// 7. SERVICES SECTION
	if cfg.SectionEnabled("services") {
		fmt.Fprintln(w, th.Bold(th.Cyan("SERVICES")))
		blankLine()
		renderServicesSection(w, snap.Services, th, cfg)
	}
}

func renderSystemBox(w io.Writer, sys model.SystemInfo, th *Theme, termWidth int) {
	boxWidth := 45
	if termWidth < boxWidth {
		boxWidth = termWidth
	}
	if boxWidth > 60 {
		boxWidth = 50
	}

	innerWidth := boxWidth - 2
	if innerWidth < 30 {
		innerWidth = 30
		boxWidth = 32
	}

	topFill := innerWidth - 10
	if topFill < 2 {
		topFill = 2
	}
	topLine := fmt.Sprintf("┌─ SYSTEM %s┐", strings.Repeat("─", topFill))
	botLine := fmt.Sprintf("└%s┘", strings.Repeat("─", innerWidth))

	fmt.Fprintln(w, th.Gray(topLine))

	type kv struct {
		k, v string
	}

	osDisplay := sys.OS
	if sys.DistroVersion != "" && !strings.Contains(sys.OS, sys.DistroVersion) {
		osDisplay = fmt.Sprintf("%s %s", sys.OS, sys.DistroVersion)
	}

	items := []kv{
		{"Host", sys.Hostname},
		{"OS", osDisplay},
		{"Kernel", sys.Kernel},
		{"Arch", sys.Arch},
		{"Uptime", sys.Uptime},
		{"Shell", sys.Shell},
	}

	for _, it := range items {
		if it.v == "" {
			continue
		}
		// Calculate padding
		label := fmt.Sprintf("%-10s", it.k)
		val := it.v
		// Truncate val if too long to prevent overflowing box
		maxValLen := innerWidth - 13
		if utf8.RuneCountInString(val) > maxValLen && maxValLen > 4 {
			runes := []rune(val)
			val = string(runes[:maxValLen-1]) + "…"
		}
		rowContent := fmt.Sprintf(" %s %s", label, val)
		fillLen := innerWidth - utf8.RuneCountInString(rowContent)
		if fillLen < 0 {
			fillLen = 0
		}
		fmt.Fprintln(w, th.Gray("│")+" "+th.Bold(label)+" "+val+strings.Repeat(" ", fillLen)+th.Gray("│"))
	}

	fmt.Fprintln(w, th.Gray(botLine))
}

func renderCPUSection(w io.Writer, cpu model.CPUInfo, th *Theme, cfg *config.Config) {
	fmt.Fprintf(w, "%-11s%s\n", "CPU", cpu.Model)
	if cpu.PhysicalCores > 0 && cpu.LogicalCores > 0 {
		fmt.Fprintf(w, "%-11s%d / %d\n", "Cores", cpu.PhysicalCores, cpu.LogicalCores)
	} else if cpu.LogicalCores > 0 {
		fmt.Fprintf(w, "%-11s%d\n", "Cores", cpu.LogicalCores)
	}

	usageStr := FormatPercent(cpu.UsagePercent)
	fmt.Fprintf(w, "%-11s%s\n", "Usage", th.UsageColor(cpu.UsagePercent, usageStr))

	if cpu.FrequencyMHz > 0 {
		if cpu.FrequencyMHz >= 1000 {
			fmt.Fprintf(w, "%-11s%.1f GHz\n", "Frequency", cpu.FrequencyMHz/1000.0)
		} else {
			fmt.Fprintf(w, "%-11s%.0f MHz\n", "Frequency", cpu.FrequencyMHz)
		}
	}

	if cpu.TemperatureC > 0 {
		fmt.Fprintf(w, "%-11s%s\n", "Temp", FormatTemp(cpu.TemperatureC, cfg.TemperatureUnit))
	}
}

func renderMemorySection(w io.Writer, mem model.MemoryInfo, th *Theme) {
	if mem.TotalBytes == 0 {
		fmt.Fprintf(w, "%-11s%s\n", "RAM", "N/A")
		return
	}

	usedStr := FormatBytes(mem.UsedBytes)
	totStr := FormatBytes(mem.TotalBytes)
	fmt.Fprintf(w, "%-11s%s / %s\n", "RAM", usedStr, totStr)

	usageStr := FormatPercent(mem.UsagePercent)
	fmt.Fprintf(w, "%-11s%s\n", "Usage", th.UsageColor(mem.UsagePercent, usageStr))

	if mem.SwapTotalBytes > 0 {
		swapUsed := FormatBytes(mem.SwapUsedBytes)
		swapTot := FormatBytes(mem.SwapTotalBytes)
		fmt.Fprintf(w, "%-11s%s / %s (%s)\n", "Swap", swapUsed, swapTot, FormatPercent(mem.SwapPercent))
	}
}

func renderStorageSection(w io.Writer, disks []model.DiskPartition, th *Theme, cfg *config.Config, termWidth int) {
	if len(disks) == 0 {
		fmt.Fprintf(w, "%-11s%s\n", "Storage", "No partitions found")
		return
	}

	// Filter disks if config specified
	var targetDisks []model.DiskPartition
	if len(cfg.DiskMounts) > 0 {
		mountMap := make(map[string]bool)
		for _, m := range cfg.DiskMounts {
			mountMap[m] = true
		}
		for _, d := range disks {
			if mountMap[d.MountPoint] {
				targetDisks = append(targetDisks, d)
			}
		}
	} else {
		// Filter out auxiliary APFS internal subvolumes for a clean dashboard
		for _, d := range disks {
			if strings.HasPrefix(d.MountPoint, "/System/Volumes/") && d.MountPoint != "/System/Volumes/Data" {
				continue
			}
			targetDisks = append(targetDisks, d)
			if len(targetDisks) >= 5 {
				break
			}
		}
		if len(targetDisks) == 0 {
			targetDisks = disks
			if len(targetDisks) > 5 {
				targetDisks = targetDisks[:5]
			}
		}
	}

	// Determine optimal column width
	maxMntLen := 10
	for _, d := range targetDisks {
		if len(d.MountPoint) > maxMntLen && len(d.MountPoint) < 24 {
			maxMntLen = len(d.MountPoint)
		}
	}
	maxMntLen += 2

	for _, d := range targetDisks {
		pctStr := fmt.Sprintf("%.0f%% used", d.UsagePercent)
		coloredPct := th.UsageColor(d.UsagePercent, pctStr)

		if termWidth >= 70 && !cfg.Compact {
			bar := UsageBar(d.UsagePercent, 12, th)
			usedStr := FormatBytes(d.UsedBytes)
			totStr := FormatBytes(d.TotalBytes)
			fmt.Fprintf(w, "%-*s %-12s %s  (%s / %s)\n", maxMntLen, d.MountPoint, coloredPct, bar, usedStr, totStr)
		} else {
			fmt.Fprintf(w, "%-*s %s\n", maxMntLen, d.MountPoint, coloredPct)
		}
	}
}

func renderGPUSection(w io.Writer, gpus []model.GPUInfo, th *Theme, cfg *config.Config) {
	for i, gpu := range gpus {
		name := gpu.Name
		if gpu.Vendor != "" && !strings.Contains(name, gpu.Vendor) {
			name = fmt.Sprintf("%s %s", gpu.Vendor, name)
		}
		fmt.Fprintf(w, "%-11s%s\n", "GPU", name)
		if gpu.Driver != "" {
			fmt.Fprintf(w, "%-11s%s\n", "Driver", gpu.Driver)
		}
		if gpu.VRAMTotalBytes > 0 {
			vUsed := FormatBytes(gpu.VRAMUsedBytes)
			vTot := FormatBytes(gpu.VRAMTotalBytes)
			fmt.Fprintf(w, "%-11s%s / %s\n", "VRAM", vUsed, vTot)
		}
		if gpu.UtilizationPercent > 0 {
			fmt.Fprintf(w, "%-11s%s\n", "Usage", FormatPercent(gpu.UtilizationPercent))
		}
		if gpu.TemperatureC > 0 {
			fmt.Fprintf(w, "%-11s%s\n", "Temp", FormatTemp(gpu.TemperatureC, cfg.TemperatureUnit))
		}
		if i < len(gpus)-1 {
			fmt.Fprintln(w)
		}
	}
}

func renderNetworkSection(w io.Writer, ifaces []model.NetworkInterface, th *Theme, cfg *config.Config) {
	if len(ifaces) == 0 {
		fmt.Fprintf(w, "%-11s%s\n", "Interface", "None")
		return
	}

	// Pick target interface: either from config or first active
	var target *model.NetworkInterface
	if len(cfg.NetworkInterfaces) > 0 {
		for _, req := range cfg.NetworkInterfaces {
			for _, iface := range ifaces {
				if iface.Name == req {
					target = &iface
					break
				}
			}
			if target != nil {
				break
			}
		}
	}

	if target == nil {
		// Pick first interface that is up and has an IPv4
		for _, iface := range ifaces {
			if iface.State == "up" && len(iface.IPv4) > 0 {
				target = &iface
				break
			}
		}
	}

	if target == nil {
		target = &ifaces[0]
	}

	fmt.Fprintf(w, "%-11s%s\n", "Interface", target.Name)

	if len(target.IPv4) > 0 {
		fmt.Fprintf(w, "%-11s%s\n", "IPv4", target.IPv4[0])
	} else {
		fmt.Fprintf(w, "%-11s%s\n", "IPv4", "N/A")
	}

	if len(target.IPv6) > 0 {
		v6 := target.IPv6[0]
		if len(v6) > 28 {
			v6 = v6[:25] + "..."
		}
		fmt.Fprintf(w, "%-11s%s\n", "IPv6", v6)
	}

	if target.RXBytes > 0 || target.TXBytes > 0 {
		fmt.Fprintf(w, "%-11s%s\n", "RX", FormatBytes(target.RXBytes))
		fmt.Fprintf(w, "%-11s%s\n", "TX", FormatBytes(target.TXBytes))
	}

	if target.SpeedMbps > 0 {
		fmt.Fprintf(w, "%-11s%s\n", "Speed", FormatSpeed(target.SpeedMbps))
	}
}

func renderServicesSection(w io.Writer, summary model.ServiceSummary, th *Theme, cfg *config.Config) {
	if len(summary.Items) == 0 {
		if summary.Manager != "" && summary.Manager != "unknown" && summary.Manager != "none" {
			fmt.Fprintf(w, "%s %s (active)\n", th.Check(), summary.Manager)
		} else {
			fmt.Fprintln(w, th.Gray("No services detected"))
		}
		return
	}

	// Filter or prioritize services
	var items []model.ServiceItem
	if len(cfg.ServiceFilter) > 0 {
		filterMap := make(map[string]bool)
		for _, sf := range cfg.ServiceFilter {
			filterMap[strings.ToLower(sf)] = true
		}
		for _, item := range summary.Items {
			if filterMap[strings.ToLower(item.Name)] {
				items = append(items, item)
			}
		}
	} else {
		// Prioritize failed services first, then running common services
		for _, item := range summary.Items {
			if item.Status == "failed" {
				items = append(items, item)
			}
		}
		for _, item := range summary.Items {
			if item.Active && item.Status != "failed" {
				items = append(items, item)
				if len(items) >= 8 {
					break
				}
			}
		}
	}

	if len(items) == 0 {
		items = summary.Items
		if len(items) > 6 {
			items = items[:6]
		}
	}

	for _, item := range items {
		if item.Status == "failed" {
			fmt.Fprintf(w, "%s %s\n", th.Cross(), item.Name)
		} else if item.Active {
			fmt.Fprintf(w, "%s %s\n", th.Check(), item.Name)
		} else {
			fmt.Fprintf(w, "%s %s (%s)\n", th.Bullet(), item.Name, item.Status)
		}
	}
}

func renderASCIILogo(w io.Writer, th *Theme) {
	logo := `  ___ _   _ ___ _ __   ___  ___| | __
 / __| | | / __| '_ \ / _ \/ _ \ |/ /
 \__ \ |_| \__ \ |_) |  __/  __/   < 
 |___/\__, |___/ .__/ \___|\___|_|\_\
      |___/    |_|                   `
	fmt.Fprintln(w, th.Cyan(logo))
}
