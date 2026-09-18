package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("expected default config, got nil")
	}
	if !cfg.SectionEnabled("cpu") {
		t.Errorf("expected cpu section enabled")
	}
	if !cfg.SectionEnabled("SYSTEM") {
		t.Errorf("expected case-insensitive matching for section")
	}
	if cfg.RefreshInterval != 2.0 {
		t.Errorf("expected refresh interval 2.0, got %f", cfg.RefreshInterval)
	}
	if cfg.TemperatureUnit != "C" {
		t.Errorf("expected temperature unit C, got %s", cfg.TemperatureUnit)
	}
}

func TestLoadCustomConfig(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "config.json")

	content := `{
		"enabled_sections": ["cpu", "memory"],
		"refresh_interval": 1.5,
		"colors": false,
		"compact": true,
		"temperature_unit": "F"
	}`

	if err := os.WriteFile(confPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(confPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if !cfg.SectionEnabled("cpu") {
		t.Error("expected cpu enabled")
	}
	if cfg.SectionEnabled("disk") {
		t.Error("expected disk disabled")
	}
	if cfg.RefreshInterval != 1.5 {
		t.Errorf("expected 1.5, got %f", cfg.RefreshInterval)
	}
	if cfg.Colors {
		t.Error("expected colors false")
	}
	if !cfg.Compact {
		t.Error("expected compact true")
	}
	if cfg.TemperatureUnit != "F" {
		t.Errorf("expected temperature unit F, got %s", cfg.TemperatureUnit)
	}
}
