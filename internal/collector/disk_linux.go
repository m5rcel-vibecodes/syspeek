//go:build linux

package collector

import (
	"bufio"
	"os"
	"sort"
	"strings"
	"syscall"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

var ignoredFSTypes = map[string]bool{
	"sysfs":       true,
	"proc":        true,
	"devtmpfs":    true,
	"devpts":      true,
	"securityfs":  true,
	"cgroup":      true,
	"cgroup2":     true,
	"pstore":      true,
	"bpf":         true,
	"debugfs":     true,
	"tracefs":     true,
	"hugetlbfs":   true,
	"mqueue":      true,
	"configfs":    true,
	"fusectl":     true,
	"binfmt_misc": true,
	"autofs":      true,
	"nsfs":        true,
	"ramfs":       true,
}

// CollectDisks collects disk usage for mounted partitions on Linux.
func CollectDisks() ([]model.DiskPartition, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var partitions []model.DiskPartition
	seenMounts := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]

		if ignoredFSTypes[fsType] {
			continue
		}

		// Skip read-only snaps unless specified
		if strings.HasPrefix(mountPoint, "/snap/") {
			continue
		}

		// Filter out tmpfs unless root
		if fsType == "tmpfs" && mountPoint != "/" {
			continue
		}

		// Filter virtual pseudo devices that do not start with /dev and are not zfs / btrfs / fuseblk
		if !strings.HasPrefix(device, "/dev/") && fsType != "zfs" && fsType != "btrfs" && fsType != "fuseblk" && mountPoint != "/" {
			continue
		}

		if seenMounts[mountPoint] {
			continue
		}
		seenMounts[mountPoint] = true

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mountPoint, &stat); err != nil {
			continue
		}

		total := stat.Blocks * uint64(stat.Bsize)
		if total == 0 {
			continue
		}

		free := stat.Bavail * uint64(stat.Bsize)
		var used uint64
		if total >= free {
			used = total - free
		}

		usagePercent := (float64(used) / float64(total)) * 100.0

		partitions = append(partitions, model.DiskPartition{
			MountPoint:   mountPoint,
			Device:       device,
			FSType:       fsType,
			TotalBytes:   total,
			UsedBytes:    used,
			FreeBytes:    free,
			UsagePercent: usagePercent,
		})
	}

	// Sort partitions so / is first, followed by shortest mount paths
	sort.Slice(partitions, func(i, j int) bool {
		if partitions[i].MountPoint == "/" {
			return true
		}
		if partitions[j].MountPoint == "/" {
			return false
		}
		return partitions[i].MountPoint < partitions[j].MountPoint
	})

	return partitions, nil
}
