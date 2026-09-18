//go:build darwin

package collector

import (
	"bufio"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectDisks collects disk usage for mounted partitions on macOS.
func CollectDisks() ([]model.DiskPartition, error) {
	var partitions []model.DiskPartition
	seenMounts := make(map[string]bool)

	// Primary mounts to check directly
	checkMounts := []string{"/", "/System/Volumes/Data"}

	// Find external volumes in /Volumes
	volMatches, _ := filepath.Glob("/Volumes/*")
	for _, vm := range volMatches {
		checkMounts = append(checkMounts, vm)
	}

	// Also parse `df -k` to discover other active user filesystems
	if cmd := exec.Command("df", "-k"); cmd != nil {
		if out, err := cmd.Output(); err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				fields := strings.Fields(scanner.Text())
				if len(fields) >= 9 {
					mnt := fields[len(fields)-1]
					dev := fields[0]
					if strings.HasPrefix(dev, "/dev/") && !strings.Contains(mnt, "mnt1") {
						checkMounts = append(checkMounts, mnt)
					}
				}
			}
		}
	}

	for _, mountPoint := range checkMounts {
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

		// Extract FSType from stat.Fstypename
		var fsTypeBuilder strings.Builder
		for _, b := range stat.Fstypename {
			if b == 0 {
				break
			}
			fsTypeBuilder.WriteByte(byte(b))
		}
		fsType := fsTypeBuilder.String()
		if fsType == "" {
			fsType = "apfs"
		}

		partitions = append(partitions, model.DiskPartition{
			MountPoint:   mountPoint,
			Device:       mountPoint,
			FSType:       fsType,
			TotalBytes:   total,
			UsedBytes:    used,
			FreeBytes:    free,
			UsagePercent: usagePercent,
		})
	}

	// Sort partitions so / is first, then /System/Volumes/Data
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
