//go:build linux

package collector

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectGPUs collects GPU hardware and performance info on Linux.
func CollectGPUs() ([]model.GPUInfo, error) {
	var gpus []model.GPUInfo

	// 1. Try nvidia-smi if present
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		nvGpus := collectNvidiaGPUs()
		if len(nvGpus) > 0 {
			return nvGpus, nil
		}
	}

	// 2. Try lspci if present
	if _, err := exec.LookPath("lspci"); err == nil {
		pciGpus := collectLspciGPUs()
		if len(pciGpus) > 0 {
			return pciGpus, nil
		}
	}

	// 3. Fallback to /sys/class/drm
	drmGpus := collectDrmGPUs()
	if len(drmGpus) > 0 {
		return drmGpus, nil
	}

	return gpus, nil
}

func collectNvidiaGPUs() []model.GPUInfo {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=name,driver_version,memory.total,memory.used,utilization.gpu,temperature.gpu",
		"--format=csv,noheader,nounits",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var list []model.GPUInfo
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) < 6 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		driver := strings.TrimSpace(parts[1])
		memTotalMiB, _ := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64)
		memUsedMiB, _ := strconv.ParseUint(strings.TrimSpace(parts[3]), 10, 64)
		util, _ := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
		temp, _ := strconv.ParseFloat(strings.TrimSpace(parts[5]), 64)

		list = append(list, model.GPUInfo{
			Name:               name,
			Vendor:             "NVIDIA",
			Driver:             driver,
			VRAMTotalBytes:     memTotalMiB * 1024 * 1024,
			VRAMUsedBytes:      memUsedMiB * 1024 * 1024,
			UtilizationPercent: util,
			TemperatureC:       temp,
		})
	}
	return list
}

func collectLspciGPUs() []model.GPUInfo {
	cmd := exec.Command("lspci")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var list []model.GPUInfo
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		lower := strings.ToLower(line)
		if strings.Contains(lower, "vga compatible controller") ||
			strings.Contains(lower, "3d controller") ||
			strings.Contains(lower, "display controller") {

			// Example: 01:00.0 VGA compatible controller: NVIDIA Corporation GA106 [GeForce RTX 3060] (rev a1)
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) != 2 {
				continue
			}

			fullDesc := parts[1]
			vendor := "Unknown"
			if strings.Contains(strings.ToLower(fullDesc), "nvidia") {
				vendor = "NVIDIA"
			} else if strings.Contains(strings.ToLower(fullDesc), "amd") || strings.Contains(strings.ToLower(fullDesc), "ati") {
				vendor = "AMD"
			} else if strings.Contains(strings.ToLower(fullDesc), "intel") {
				vendor = "Intel"
			}

			list = append(list, model.GPUInfo{
				Name:   fullDesc,
				Vendor: vendor,
			})
		}
	}
	return list
}

func collectDrmGPUs() []model.GPUInfo {
	cards, _ := filepath.Glob("/sys/class/drm/card[0-9]")
	var list []model.GPUInfo

	for _, card := range cards {
		vendorData, err := os.ReadFile(filepath.Join(card, "device", "vendor"))
		if err != nil {
			continue
		}
		vendorHex := strings.TrimSpace(string(vendorData))

		vendor := "Unknown"
		switch vendorHex {
		case "0x10de":
			vendor = "NVIDIA"
		case "0x1002":
			vendor = "AMD"
		case "0x8086":
			vendor = "Intel"
		}

		driverLink, _ := os.Readlink(filepath.Join(card, "device", "driver"))
		driver := filepath.Base(driverLink)

		list = append(list, model.GPUInfo{
			Name:   filepath.Base(card),
			Vendor: vendor,
			Driver: driver,
		})
	}
	return list
}
