//go:build darwin

package collector

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/m5rcel-vibecodes/syspeekinternal/model"
)

// CollectSystem collects system information on macOS.
func CollectSystem() (*model.SystemInfo, error) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	distro, version := getDarwinVersion()
	kernel, _ := syscall.Sysctl("kern.osrelease")
	if kernel == "" {
		kernel = "Darwin"
	}

	uptimeSec := getDarwinUptime()
	desktop, wm := detectDesktopEnv()
	pkgMgr := detectDarwinPackageManager()

	info := &model.SystemInfo{
		Hostname:       hostname,
		OS:             distro,
		Distribution:   distro,
		DistroVersion:  version,
		Kernel:         kernel,
		Arch:           normalizeArch(runtime.GOARCH),
		UptimeSeconds:  uptimeSec,
		Uptime:         FormatUptime(uptimeSec),
		CurrentUser:    getCurrentUser(),
		Shell:          getCurrentShell(),
		Desktop:        desktop,
		WindowManager:  wm,
		InitSystem:     "launchd",
		PackageManager: pkgMgr,
	}

	return info, nil
}

func getDarwinVersion() (string, string) {
	cmd := exec.Command("sw_vers")
	out, err := cmd.Output()
	if err != nil {
		return "macOS", ""
	}

	var product, version string
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		parts := strings.SplitN(l, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		switch k {
		case "ProductName":
			product = v
		case "ProductVersion":
			version = v
		}
	}

	if product == "" {
		product = "macOS"
	}
	return product, version
}

var darwinBoottimeRegex = regexp.MustCompile(`sec\s*=\s*(\d+)`)

func getDarwinUptime() uint64 {
	cmd := exec.Command("sysctl", "-n", "kern.boottime")
	out, err := cmd.Output()
	if err != nil {
		return 0
	}

	matches := darwinBoottimeRegex.FindStringSubmatch(string(out))
	if len(matches) > 1 {
		if bootSec, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			nowSec := time.Now().Unix()
			if nowSec > bootSec {
				return uint64(nowSec - bootSec)
			}
		}
	}
	return 0
}

func detectDarwinPackageManager() string {
	var found []string
	if _, err := exec.LookPath("brew"); err == nil {
		found = append(found, "brew")
	}
	if _, err := exec.LookPath("port"); err == nil {
		found = append(found, "port")
	}
	if _, err := exec.LookPath("nix-env"); err == nil {
		found = append(found, "nix")
	}
	return strings.Join(found, ", ")
}
