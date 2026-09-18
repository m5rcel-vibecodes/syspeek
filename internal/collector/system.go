package collector

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// normalizeArch maps Go runtime architectures to standard system nomenclature.
func normalizeArch(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	case "386":
		return "i686"
	case "arm":
		return "armv7l"
	default:
		return arch
	}
}

// getCurrentUser returns the username running the process.
func getCurrentUser() string {
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	if u := os.Getenv("LOGNAME"); u != "" {
		return u
	}
	if cur, err := user.Current(); err == nil && cur.Username != "" {
		parts := strings.Split(cur.Username, "\\")
		return parts[len(parts)-1]
	}
	return "unknown"
}

// getCurrentShell returns the current shell name (e.g., "zsh", "bash").
func getCurrentShell() string {
	sh := os.Getenv("SHELL")
	if sh != "" {
		return filepath.Base(sh)
	}
	return "sh"
}

// detectDesktopEnv returns common desktop environment and window manager names.
func detectDesktopEnv() (desktop string, wm string) {
	if runtime.GOOS == "darwin" {
		return "Aqua", "Quartz Compositor"
	}

	if d := os.Getenv("XDG_CURRENT_DESKTOP"); d != "" {
		desktop = d
	} else if d := os.Getenv("DESKTOP_SESSION"); d != "" {
		desktop = d
	}

	if w := os.Getenv("WAYLAND_DISPLAY"); w != "" {
		wm = "Wayland"
	} else if os.Getenv("DISPLAY") != "" {
		wm = "X11"
	}

	// Try detecting common tiling or stacking window managers from environment or session
	if strings.Contains(strings.ToLower(desktop), "sway") {
		wm = "Sway"
	} else if strings.Contains(strings.ToLower(desktop), "hyprland") {
		wm = "Hyprland"
	} else if strings.Contains(strings.ToLower(desktop), "i3") {
		wm = "i3"
	} else if strings.Contains(strings.ToLower(desktop), "kde") || strings.Contains(strings.ToLower(desktop), "plasma") {
		wm = "KWin"
	} else if strings.Contains(strings.ToLower(desktop), "gnome") {
		wm = "Mutter"
	} else if strings.Contains(strings.ToLower(desktop), "xfce") {
		wm = "Xfwm4"
	}

	return desktop, wm
}
