//go:build darwin

package collector

import (
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectCPU collects CPU information on macOS.
func CollectCPU() (*model.CPUInfo, error) {
	modelName, _ := syscall.Sysctl("machdep.cpu.brand_string")
	if modelName == "" {
		if hwModel, _ := syscall.Sysctl("hw.model"); hwModel != "" {
			modelName = "Apple " + hwModel
		} else {
			modelName = "Apple Silicon"
		}
	}

	vendor := "Apple"
	if strings.Contains(strings.ToLower(modelName), "intel") {
		vendor = "Intel"
	} else if strings.Contains(strings.ToLower(modelName), "amd") {
		vendor = "AMD"
	}

	physCores := getDarwinSysctlInt("hw.physicalcpu")
	logCores := getDarwinSysctlInt("hw.logicalcpu")
	if logCores <= 0 {
		logCores = runtime.NumCPU()
	}
	if physCores <= 0 {
		physCores = logCores
	}

	freq := getDarwinCPUFrequency()
	loadAvg := getDarwinLoadAvg()

	// Compute estimated usage based on 1-min load avg normalized by logical core count
	var usage float64
	if logCores > 0 && loadAvg[0] > 0 {
		usage = (loadAvg[0] / float64(logCores)) * 100.0
		if usage > 100.0 {
			usage = 100.0
		}
	}

	return &model.CPUInfo{
		Model:         modelName,
		Vendor:        vendor,
		PhysicalCores: physCores,
		LogicalCores:  logCores,
		FrequencyMHz:  freq,
		UsagePercent:  usage,
		LoadAverages:  loadAvg,
		TemperatureC:  0, // Temperature requires private IOKit or root on macOS
	}, nil
}

func getDarwinSysctlInt(name string) int {
	cmd := exec.Command("sysctl", "-n", name)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	val, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0
	}
	return val
}

func getDarwinCPUFrequency() float64 {
	hz := getDarwinSysctlInt("hw.cpufrequency")
	if hz > 0 {
		return float64(hz) / 1000000.0
	}
	return 0
}

var darwinLoadRegex = regexp.MustCompile(`\{\s*([\d\.]+)\s+([\d\.]+)\s+([\d\.]+)\s*\}`)

func getDarwinLoadAvg() [3]float64 {
	var loads [3]float64
	cmd := exec.Command("sysctl", "-n", "vm.loadavg")
	out, err := cmd.Output()
	if err != nil {
		return loads
	}

	matches := darwinLoadRegex.FindStringSubmatch(string(out))
	if len(matches) == 4 {
		loads[0], _ = strconv.ParseFloat(matches[1], 64)
		loads[1], _ = strconv.ParseFloat(matches[2], 64)
		loads[2], _ = strconv.ParseFloat(matches[3], 64)
	}
	return loads
}
