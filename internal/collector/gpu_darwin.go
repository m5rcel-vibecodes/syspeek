//go:build darwin

package collector

import (
	"bufio"
	"os/exec"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectGPUs collects GPU hardware info on macOS.
func CollectGPUs() ([]model.GPUInfo, error) {
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var gpus []model.GPUInfo
	var currentGPU *model.GPUInfo

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Chipset Model:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				if currentGPU != nil {
					gpus = append(gpus, *currentGPU)
				}
				currentGPU = &model.GPUInfo{
					Name:   strings.TrimSpace(parts[1]),
					Vendor: "Apple",
				}
			}
		} else if currentGPU != nil {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				switch k {
				case "Vendor":
					if strings.Contains(strings.ToLower(v), "apple") {
						currentGPU.Vendor = "Apple"
					} else if strings.Contains(strings.ToLower(v), "intel") {
						currentGPU.Vendor = "Intel"
					} else if strings.Contains(strings.ToLower(v), "amd") {
						currentGPU.Vendor = "AMD"
					} else if strings.Contains(strings.ToLower(v), "nvidia") {
						currentGPU.Vendor = "NVIDIA"
					} else {
						currentGPU.Vendor = v
					}
				case "Metal Support":
					currentGPU.Driver = v
				case "VRAM (Total)", "Total VRAM":
					currentGPU.VRAMTotalBytes = parseMemSizeUnit(v)
				}
			}
		}
	}

	if currentGPU != nil {
		gpus = append(gpus, *currentGPU)
	}

	return gpus, nil
}
