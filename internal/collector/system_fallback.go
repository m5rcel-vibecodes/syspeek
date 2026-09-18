//go:build !linux && !darwin

package collector

import (
	"os"
	"runtime"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectSystem is the fallback collector for unsupported platforms.
func CollectSystem() (*model.SystemInfo, error) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	return &model.SystemInfo{
		Hostname:      hostname,
		OS:            runtime.GOOS,
		Distribution:  runtime.GOOS,
		Kernel:        "unknown",
		Arch:          normalizeArch(runtime.GOARCH),
		UptimeSeconds: 0,
		Uptime:        "N/A",
		CurrentUser:   getCurrentUser(),
		Shell:         getCurrentShell(),
		Desktop:       "N/A",
		WindowManager: "N/A",
		InitSystem:    "N/A",
	}, nil
}
