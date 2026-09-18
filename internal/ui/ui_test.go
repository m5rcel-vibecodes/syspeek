package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/m5rcel-vibecodes/syspeek/internal/config"
	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024 * 5, "5.0 MB"},
		{1024 * 1024 * 1024 * 16, "16.0 GB"},
	}

	for _, tc := range tests {
		got := FormatBytes(tc.bytes)
		if got != tc.expected {
			t.Errorf("FormatBytes(%d) = %q, expected %q", tc.bytes, got, tc.expected)
		}
	}
}

func TestUsageBar(t *testing.T) {
	th := NewTheme(false) // monochrome
	bar := UsageBar(50.0, 10, th)
	if !strings.Contains(bar, "█████") {
		t.Errorf("expected 5 filled blocks in 10-char bar, got %s", bar)
	}
}

func TestRenderDashboard(t *testing.T) {
	snap := &model.Snapshot{
		System: model.SystemInfo{
			Hostname: "test-host",
			OS:       "Linux",
			Kernel:   "6.1.0",
			Arch:     "x86_64",
			Uptime:   "2d 4h",
			Shell:    "zsh",
		},
		CPU: model.CPUInfo{
			Model:        "Test CPU",
			LogicalCores: 8,
			UsagePercent: 25.0,
		},
		Memory: model.MemoryInfo{
			TotalBytes:   16 * 1024 * 1024 * 1024,
			UsedBytes:    4 * 1024 * 1024 * 1024,
			UsagePercent: 25.0,
		},
	}

	cfg := config.DefaultConfig()
	cfg.Colors = false

	var buf bytes.Buffer
	RenderDashboard(&buf, snap, cfg)
	out := buf.String()

	if !strings.Contains(out, "test-host") {
		t.Error("dashboard output missing hostname")
	}
	if !strings.Contains(out, "Test CPU") {
		t.Error("dashboard output missing CPU model")
	}
}
