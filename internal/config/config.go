package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config defines the user customizable settings for syspeek.
type Config struct {
	EnabledSections   []string `json:"enabled_sections"`
	RefreshInterval   float64  `json:"refresh_interval"`
	Colors            bool     `json:"colors"`
	Compact           bool     `json:"compact"`
	ShowASCIILogo     bool     `json:"show_ascii_logo"`
	TemperatureUnit   string   `json:"temperature_unit"` // "C" or "F"
	NetworkInterfaces []string `json:"network_interfaces"`
	DiskMounts        []string `json:"disk_mounts"`
	ServiceFilter     []string `json:"service_filter"`
}

// DefaultConfig returns the default configuration settings.
func DefaultConfig() *Config {
	return &Config{
		EnabledSections: []string{
			"system",
			"cpu",
			"memory",
			"storage",
			"network",
			"services",
		},
		RefreshInterval:   2.0,
		Colors:            true,
		Compact:           false,
		ShowASCIILogo:     false,
		TemperatureUnit:   "C",
		NetworkInterfaces: []string{},
		DiskMounts:        []string{},
		ServiceFilter:     []string{},
	}
}

// Load loads the configuration from disk, searching default paths or an explicit path.
func Load(explicitPath string) (*Config, error) {
	cfg := DefaultConfig()

	targetPath := explicitPath
	if targetPath == "" {
		targetPath = FindConfigFile()
	}

	if targetPath == "" {
		// No config found, return defaults cleanly.
		return cfg, nil
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) && explicitPath == "" {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config file %s: %w", targetPath, err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", targetPath, err)
	}

	// Normalize settings
	cfg.TemperatureUnit = strings.ToUpper(strings.TrimSpace(cfg.TemperatureUnit))
	if cfg.TemperatureUnit != "F" {
		cfg.TemperatureUnit = "C"
	}
	if cfg.RefreshInterval <= 0 {
		cfg.RefreshInterval = 2.0
	}

	return cfg, nil
}

// FindConfigFile locates the configuration file in standard locations.
func FindConfigFile() string {
	if envPath := os.Getenv("SYSPEEK_CONFIG"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	candidates := []string{}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		candidates = append(candidates, filepath.Join(xdg, "syspeek", "config.json"))
	}

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "syspeek", "config.json"))
		candidates = append(candidates, filepath.Join(home, ".syspeek.json"))
	}

	candidates = append(candidates, "/etc/syspeek/config.json")

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// SectionEnabled reports whether a given dashboard section should be displayed.
func (c *Config) SectionEnabled(section string) bool {
	if len(c.EnabledSections) == 0 {
		return true
	}
	target := strings.ToLower(strings.TrimSpace(section))
	for _, s := range c.EnabledSections {
		if strings.ToLower(strings.TrimSpace(s)) == target {
			return true
		}
	}
	return false
}
