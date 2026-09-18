//go:build linux

package collector

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectMemory collects RAM and Swap information on Linux.
func CollectMemory() (*model.MemoryInfo, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		memTotal, memFree, memAvailable uint64
		buffers, cached                 uint64
		swapTotal, swapFree             uint64
	)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		valFields := strings.Fields(parts[1])
		if len(valFields) == 0 {
			continue
		}

		valKB, err := strconv.ParseUint(valFields[0], 10, 64)
		if err != nil {
			continue
		}
		valBytes := valKB * 1024

		switch key {
		case "MemTotal":
			memTotal = valBytes
		case "MemFree":
			memFree = valBytes
		case "MemAvailable":
			memAvailable = valBytes
		case "Buffers":
			buffers = valBytes
		case "Cached":
			cached = valBytes
		case "SwapTotal":
			swapTotal = valBytes
		case "SwapFree":
			swapFree = valBytes
		}
	}

	var usedBytes uint64
	if memAvailable > 0 && memTotal >= memAvailable {
		usedBytes = memTotal - memAvailable
	} else if memTotal >= (memFree + buffers + cached) {
		usedBytes = memTotal - (memFree + buffers + cached)
		memAvailable = memFree + buffers + cached
	}

	var usagePercent float64
	if memTotal > 0 {
		usagePercent = (float64(usedBytes) / float64(memTotal)) * 100.0
	}

	var swapUsed uint64
	var swapPercent float64
	if swapTotal > 0 && swapTotal >= swapFree {
		swapUsed = swapTotal - swapFree
		swapPercent = (float64(swapUsed) / float64(swapTotal)) * 100.0
	}

	return &model.MemoryInfo{
		TotalBytes:     memTotal,
		UsedBytes:      usedBytes,
		FreeBytes:      memFree,
		AvailableBytes: memAvailable,
		CachedBytes:    cached,
		BuffersBytes:   buffers,
		UsagePercent:   usagePercent,
		SwapTotalBytes: swapTotal,
		SwapUsedBytes:  swapUsed,
		SwapFreeBytes:  swapFree,
		SwapPercent:    swapPercent,
	}, nil
}
