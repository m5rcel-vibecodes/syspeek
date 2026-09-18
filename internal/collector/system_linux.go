//go:build linux

package collector

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectSystem collects system information on Linux.
func CollectSystem() (*model.SystemInfo, error) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	distro, version := parseLinuxOSRelease()
	kernel := getLinuxKernel()
	uptimeSec := getLinuxUptime()
	initSys := getLinuxInitSystem()
	pkgMgr := detectLinuxPackageManager()
	desktop, wm := detectDesktopEnv()

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
		InitSystem:     initSys,
		PackageManager: pkgMgr,
	}

	return info, nil
}

func parseLinuxOSRelease() (string, string) {
	paths := []string{"/etc/os-release", "/usr/lib/os-release"}
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()

		var prettyName, name, versionID string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := parts[0]
			val := strings.Trim(parts[1], `"'`)
			switch key {
			case "PRETTY_NAME":
				prettyName = val
			case "NAME":
				name = val
			case "VERSION_ID":
				versionID = val
			}
		}

		if prettyName != "" {
			return prettyName, versionID
		}
		if name != "" {
			return name, versionID
		}
	}
	return "Linux", ""
}

func getLinuxKernel() string {
	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err == nil {
		return utsnameToString(uts.Release[:])
	}

	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err == nil {
		return strings.TrimSpace(string(data))
	}
	return "unknown"
}

func utsnameToString(in []int8) string {
	b := make([]byte, 0, len(in))
	for _, v := range in {
		if v == 0 {
			break
		}
		b = append(b, byte(v))
	}
	return string(b)
}

func getLinuxUptime() uint64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) > 0 {
		if sec, err := strconv.ParseFloat(fields[0], 64); err == nil {
			return uint64(sec)
		}
	}
	return 0
}

func getLinuxInitSystem() string {
	data, err := os.ReadFile("/proc/1/comm")
	if err == nil {
		name := strings.TrimSpace(string(data))
		if name != "" {
			return name
		}
	}

	// Fallback to checking symlink target of /sbin/init
	target, err := os.Readlink("/sbin/init")
	if err == nil {
		if strings.Contains(target, "systemd") {
			return "systemd"
		}
		if strings.Contains(target, "openrc") {
			return "openrc"
		}
	}

	return "unknown"
}

func detectLinuxPackageManager() string {
	managers := []string{
		"pacman",
		"apt",
		"dnf",
		"zypper",
		"apk",
		"emerge",
		"nix-env",
		"flatpak",
		"snap",
	}

	var found []string
	for _, m := range managers {
		if _, err := exec.LookPath(m); err == nil {
			found = append(found, m)
			if len(found) >= 2 {
				break
			}
		}
	}

	if len(found) > 0 {
		return strings.Join(found, ", ")
	}
	return ""
}
