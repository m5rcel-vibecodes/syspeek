package collector

import (
	"testing"
)

func TestCollectSnapshot(t *testing.T) {
	snap, err := CollectSnapshot()
	if err != nil {
		t.Fatalf("CollectSnapshot failed: %v", err)
	}
	if snap == nil {
		t.Fatal("Snapshot is nil")
	}

	if snap.System.OS == "" {
		t.Error("expected non-empty OS name")
	}
	if snap.CPU.LogicalCores <= 0 {
		t.Errorf("expected positive logical cores, got %d", snap.CPU.LogicalCores)
	}
	if snap.Memory.TotalBytes == 0 {
		t.Logf("Notice: Total memory is 0 or unreadable in current test environment")
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		sec      uint64
		expected string
	}{
		{0, "0m"},
		{59, "0m"},
		{60, "1m"},
		{3600, "1h 0m"},
		{3665, "1h 1m"},
		{86400, "1d 0h 0m"},
		{310860, "3d 14h 21m"},
	}

	for _, tc := range tests {
		got := FormatUptime(tc.sec)
		if got != tc.expected {
			t.Errorf("FormatUptime(%d) = %q, expected %q", tc.sec, got, tc.expected)
		}
	}
}

func TestNormalizeArch(t *testing.T) {
	if got := normalizeArch("amd64"); got != "x86_64" {
		t.Errorf("expected x86_64, got %s", got)
	}
	if got := normalizeArch("arm64"); got != "aarch64" {
		t.Errorf("expected aarch64, got %s", got)
	}
}
