package ui

import (
	"fmt"
	"strings"
)

// FormatBytes converts raw bytes into human-readable representation (e.g., 5.4 GB, 12.4 MB).
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	val := float64(b) / float64(div)
	if val >= 10.0 {
		return fmt.Sprintf("%.1f %s", val, units[exp])
	}
	return fmt.Sprintf("%.1f %s", val, units[exp])
}

// FormatPercent formats a floating percentage value.
func FormatPercent(pct float64) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return fmt.Sprintf("%.0f%%", pct)
}

// UsageBar renders a visual progress bar e.g. [███████░░░░░░░░░░░░░].
func UsageBar(percent float64, width int, th *Theme) string {
	if width < 4 {
		width = 10
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filledLen := int((percent / 100.0) * float64(width))
	if filledLen > width {
		filledLen = width
	}
	emptyLen := width - filledLen

	filledStr := strings.Repeat("█", filledLen)
	emptyStr := strings.Repeat("░", emptyLen)

	if th != nil && th.Colored {
		return fmt.Sprintf("[%s%s]", th.UsageColor(percent, filledStr), th.Gray(emptyStr))
	}
	return fmt.Sprintf("[%s%s]", filledStr, emptyStr)
}

// FormatTemp formats Celsius temperature with target unit (C or F).
func FormatTemp(celsius float64, unit string) string {
	if celsius <= 0 {
		return "N/A"
	}
	if strings.ToUpper(unit) == "F" {
		f := (celsius * 9.0 / 5.0) + 32.0
		return fmt.Sprintf("%.1f°F", f)
	}
	return fmt.Sprintf("%.1f°C", celsius)
}

// FormatSpeed formats network interface link speed in Mbps.
func FormatSpeed(mbps int64) string {
	if mbps <= 0 {
		return "N/A"
	}
	if mbps >= 1000 {
		if mbps%1000 == 0 {
			return fmt.Sprintf("%d Gbps", mbps/1000)
		}
		return fmt.Sprintf("%.1f Gbps", float64(mbps)/1000.0)
	}
	return fmt.Sprintf("%d Mbps", mbps)
}
