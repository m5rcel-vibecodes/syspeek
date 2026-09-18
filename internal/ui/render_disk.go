package ui

import (
	"fmt"
	"io"

	"github.com/m5rcel-vibecodes/syspeek/internal/config"
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// RenderDiskView prints a detailed storage and partitions report.
func RenderDiskView(w io.Writer, disks []model.DiskPartition, cfg *config.Config) {
	th := NewTheme(cfg.Colors)

	fmt.Fprintln(w, th.Bold(th.Cyan("STORAGE & PARTITIONS")))
	fmt.Fprintln(w, th.Gray("─────────────────────────────────────────────────────────────────────────────"))

	if len(disks) == 0 {
		fmt.Fprintln(w, "No mounted storage partitions found.")
		return
	}

	maxMntLen := 14
	for _, d := range disks {
		if len(d.MountPoint) > maxMntLen {
			maxMntLen = len(d.MountPoint)
		}
	}
	if maxMntLen > 30 {
		maxMntLen = 30
	}

	headerFmt := fmt.Sprintf("%%-%ds %%-14s %%-8s %%-10s %%-10s %%-10s %%-7s %%s\n", maxMntLen)
	rowFmt := fmt.Sprintf("%%-%ds %%-14s %%-8s %%-10s %%-10s %%-10s %%-7s %%s\n", maxMntLen)

	fmt.Fprintf(w, headerFmt, "MOUNT POINT", "DEVICE", "FSTYPE", "TOTAL", "USED", "FREE", "USAGE", "BAR")
	fmt.Fprintln(w, th.Gray("──────────────────────────────────────────────────────────────────────────────────"))

	for _, d := range disks {
		bar := UsageBar(d.UsagePercent, 12, th)
		coloredPct := th.UsageColor(d.UsagePercent, FormatPercent(d.UsagePercent))

		mnt := d.MountPoint
		if len(mnt) > maxMntLen {
			mnt = "…" + mnt[len(mnt)-(maxMntLen-1):]
		}

		dev := d.Device
		if len(dev) > 14 {
			dev = "…" + dev[len(dev)-13:]
		}

		fmt.Fprintf(w, rowFmt,
			mnt,
			dev,
			d.FSType,
			FormatBytes(d.TotalBytes),
			FormatBytes(d.UsedBytes),
			FormatBytes(d.FreeBytes),
			coloredPct,
			bar,
		)
	}
}
