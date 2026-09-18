//go:build linux

package collector

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectCPU collects CPU information on Linux.
func CollectCPU() (*model.CPUInfo, error) {
	modelName, vendor, physCores, logCores := parseLinuxCPUInfo()
	if logCores <= 0 {
		logCores = runtime.NumCPU()
	}
	if physCores <= 0 {
		physCores = logCores
	}

	freq := getLinuxCPUFrequency()
	loadAvg := getLinuxLoadAvg()
	usage := getLinuxCPUUsage()
	temp := getLinuxCPUTemp()

	return &model.CPUInfo{
		Model:         modelName,
		Vendor:        vendor,
		PhysicalCores: physCores,
		LogicalCores:  logCores,
		FrequencyMHz:  freq,
		UsagePercent:  usage,
		LoadAverages:  loadAvg,
		TemperatureC:  temp,
	}, nil
}

func parseLinuxCPUInfo() (modelName, vendor string, physicalCores, logicalCores int) {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "Unknown CPU", "", 0, 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	coreMap := make(map[string]bool)

	var currentPhysicalID string
	var currentCoreID string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if currentPhysicalID != "" && currentCoreID != "" {
				coreMap[currentPhysicalID+":"+currentCoreID] = true
			}
			currentPhysicalID = ""
			currentCoreID = ""
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "model name", "Hardware", "Processor":
			if modelName == "" {
				modelName = val
			}
		case "vendor_id":
			if vendor == "" {
				vendor = val
			}
		case "processor":
			logicalCores++
		case "physical id":
			currentPhysicalID = val
		case "core id":
			currentCoreID = val
		case "cpu cores":
			if physicalCores == 0 {
				if c, err := strconv.Atoi(val); err == nil {
					physicalCores = c
				}
			}
		}
	}

	if currentPhysicalID != "" && currentCoreID != "" {
		coreMap[currentPhysicalID+":"+currentCoreID] = true
	}

	if len(coreMap) > 0 {
		physicalCores = len(coreMap)
	}

	if modelName == "" {
		modelName = "Generic CPU"
	}

	return modelName, vendor, physicalCores, logicalCores
}

func getLinuxCPUFrequency() float64 {
	// Try /sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq
	data, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq")
	if err == nil {
		if khz, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err == nil {
			return khz / 1000.0
		}
	}

	// Try /proc/cpuinfo cpu MHz
	f, err := os.Open("/proc/cpuinfo")
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "cpu MHz") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					if mhz, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
						return mhz
					}
				}
			}
		}
	}

	return 0
}

func getLinuxLoadAvg() [3]float64 {
	var loads [3]float64
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return loads
	}

	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		loads[0], _ = strconv.ParseFloat(fields[0], 64)
		loads[1], _ = strconv.ParseFloat(fields[1], 64)
		loads[2], _ = strconv.ParseFloat(fields[2], 64)
	}
	return loads
}

func readLinuxStatCPU() (total, idle uint64, err error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		fields := strings.Fields(l)
		if len(fields) > 4 && fields[0] == "cpu" {
			var sum uint64
			for i := 1; i < len(fields); i++ {
				v, _ := strconv.ParseUint(fields[i], 10, 64)
				sum += v
				if i == 4 { // idle
					idle += v
				}
				if i == 5 { // iowait
					idle += v
				}
			}
			return sum, idle, nil
		}
	}
	return 0, 0, fmt.Errorf("cpu line not found")
}

func getLinuxCPUUsage() float64 {
	t1, i1, err := readLinuxStatCPU()
	if err != nil {
		return 0
	}
	time.Sleep(80 * time.Millisecond)
	t2, i2, err := readLinuxStatCPU()
	if err != nil {
		return 0
	}

	deltaTotal := float64(t2 - t1)
	deltaIdle := float64(i2 - i1)

	if deltaTotal <= 0 {
		return 0
	}

	usage := 100.0 * (deltaTotal - deltaIdle) / deltaTotal
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	return usage
}

func getLinuxCPUTemp() float64 {
	// Look through /sys/class/thermal/thermal_zone*
	zones, _ := filepath.Glob("/sys/class/thermal/thermal_zone*")
	for _, z := range zones {
		typeData, _ := os.ReadFile(filepath.Join(z, "type"))
		tName := strings.ToLower(strings.TrimSpace(string(typeData)))
		if strings.Contains(tName, "x86_pkg_temp") ||
			strings.Contains(tName, "cpu") ||
			strings.Contains(tName, "k10temp") ||
			strings.Contains(tName, "coretemp") ||
			strings.Contains(tName, "soc_thermal") {
			tempData, err := os.ReadFile(filepath.Join(z, "temp"))
			if err == nil {
				if milli, err := strconv.ParseFloat(strings.TrimSpace(string(tempData)), 64); err == nil && milli > 0 {
					return milli / 1000.0
				}
			}
		}
	}

	// Also check /sys/class/hwmon/hwmon*
	hwmons, _ := filepath.Glob("/sys/class/hwmon/hwmon*")
	for _, h := range hwmons {
		hName, _ := os.ReadFile(filepath.Join(h, "name"))
		nameStr := strings.ToLower(strings.TrimSpace(string(hName)))
		if strings.Contains(nameStr, "coretemp") || strings.Contains(nameStr, "k10temp") || strings.Contains(nameStr, "cpu") {
			inputs, _ := filepath.Glob(filepath.Join(h, "temp*_input"))
			for _, inp := range inputs {
				data, err := os.ReadFile(inp)
				if err == nil {
					if milli, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err == nil && milli > 0 {
						return milli / 1000.0
					}
				}
			}
		}
	}

	// Fallback to first thermal zone if found
	if len(zones) > 0 {
		tempData, err := os.ReadFile(filepath.Join(zones[0], "temp"))
		if err == nil {
			if milli, err := strconv.ParseFloat(strings.TrimSpace(string(tempData)), 64); err == nil && milli > 0 {
				return milli / 1000.0
			}
		}
	}

	return 0
}
