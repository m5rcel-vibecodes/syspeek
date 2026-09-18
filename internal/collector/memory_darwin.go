//go:build darwin

package collector

import (
	"bufio"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectMemory collects RAM and Swap information on macOS.
func CollectMemory() (*model.MemoryInfo, error) {
	totalBytes := getDarwinMemTotal()
	usedBytes, freeBytes, availBytes, cachedBytes := getDarwinVMStat(totalBytes)

	var usagePercent float64
	if totalBytes > 0 {
		usagePercent = (float64(usedBytes) / float64(totalBytes)) * 100.0
	}

	swapTotal, swapUsed, swapFree := getDarwinSwap()
	var swapPercent float64
	if swapTotal > 0 {
		swapPercent = (float64(swapUsed) / float64(swapTotal)) * 100.0
	}

	return &model.MemoryInfo{
		TotalBytes:     totalBytes,
		UsedBytes:      usedBytes,
		FreeBytes:      freeBytes,
		AvailableBytes: availBytes,
		CachedBytes:    cachedBytes,
		UsagePercent:   usagePercent,
		SwapTotalBytes: swapTotal,
		SwapUsedBytes:  swapUsed,
		SwapFreeBytes:  swapFree,
		SwapPercent:    swapPercent,
	}, nil
}

func getDarwinMemTotal() uint64 {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	val, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0
	}
	return val
}

var pageSizeRegex = regexp.MustCompile(`page size of (\d+) bytes`)

func getDarwinVMStat(totalBytes uint64) (used, free, avail, cached uint64) {
	cmd := exec.Command("vm_stat")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, 0, 0
	}

	var pageSize uint64 = 4096
	matches := pageSizeRegex.FindStringSubmatch(string(out))
	if len(matches) > 1 {
		if ps, err := strconv.ParseUint(matches[1], 10, 64); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	var freePages, activePages, inactivePages, speculativePages, wiredPages, compressorPages, fileBackedPages uint64

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valStr := strings.Trim(strings.TrimSpace(parts[1]), ".")
		v, err := strconv.ParseUint(valStr, 10, 64)
		if err != nil {
			continue
		}

		switch key {
		case "Pages free":
			freePages = v
		case "Pages active":
			activePages = v
		case "Pages inactive":
			inactivePages = v
		case "Pages speculative":
			speculativePages = v
		case "Pages wired down":
			wiredPages = v
		case "Pages occupied by compressor":
			compressorPages = v
		case "File-backed pages":
			fileBackedPages = v
		}
	}

	used = (activePages + wiredPages + compressorPages) * pageSize
	free = (freePages + speculativePages) * pageSize
	cached = fileBackedPages * pageSize
	if totalBytes > used {
		avail = totalBytes - used
	} else {
		avail = free + (inactivePages * pageSize)
	}

	return used, free, avail, cached
}

var swapRegex = regexp.MustCompile(`total\s*=\s*([\d\.]+[KMGTP]?)\s+used\s*=\s*([\d\.]+[KMGTP]?)\s+free\s*=\s*([\d\.]+[KMGTP]?)`)

func getDarwinSwap() (total, used, free uint64) {
	cmd := exec.Command("sysctl", "-n", "vm.swapusage")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, 0
	}

	matches := swapRegex.FindStringSubmatch(string(out))
	if len(matches) == 4 {
		total = parseMemSizeUnit(matches[1])
		used = parseMemSizeUnit(matches[2])
		free = parseMemSizeUnit(matches[3])
	}
	return total, used, free
}

func parseMemSizeUnit(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	unit := s[len(s)-1]
	valStr := s[:len(s)-1]
	f, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0
	}

	switch unit {
	case 'K', 'k':
		return uint64(f * 1024)
	case 'M', 'm':
		return uint64(f * 1024 * 1024)
	case 'G', 'g':
		return uint64(f * 1024 * 1024 * 1024)
	case 'T', 't':
		return uint64(f * 1024 * 1024 * 1024 * 1024)
	default:
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return uint64(f)
		}
		return 0
	}
}
