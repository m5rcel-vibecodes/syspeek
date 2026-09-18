//go:build linux

package collector

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectNetwork collects network interface metrics on Linux.
func CollectNetwork() ([]model.NetworkInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	statsMap := parseLinuxProcNetDev()

	var result []model.NetworkInterface
	for _, iface := range ifaces {
		state := "down"
		if iface.Flags&net.FlagUp != 0 {
			state = "up"
		}

		var ipv4s, ipv6s []string
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				ip := ipNet.IP
				if ip.IsLoopback() && iface.Name != "lo" {
					continue
				}
				if ip.To4() != nil {
					ipv4s = append(ipv4s, ip.String())
				} else if ip.To16() != nil {
					ipv6s = append(ipv6s, ip.String())
				}
			}
		}

		speed := getLinuxInterfaceSpeed(iface.Name)
		stat := statsMap[iface.Name]

		result = append(result, model.NetworkInterface{
			Name:      iface.Name,
			State:     state,
			IPv4:      ipv4s,
			IPv6:      ipv6s,
			MAC:       iface.HardwareAddr.String(),
			SpeedMbps: speed,
			RXBytes:   stat.rxBytes,
			TXBytes:   stat.txBytes,
			RXPackets: stat.rxPackets,
			TXPackets: stat.txPackets,
			RXErrors:  stat.rxErrors,
			TXErrors:  stat.txErrors,
		})
	}

	// Sort active interfaces with IPv4 addresses first, excluding loopback from top if others exist
	sort.Slice(result, func(i, j int) bool {
		iScore := networkSortScore(result[i])
		jScore := networkSortScore(result[j])
		if iScore != jScore {
			return iScore > jScore
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}

type netDevStat struct {
	rxBytes   uint64
	rxPackets uint64
	rxErrors  uint64
	txBytes   uint64
	txPackets uint64
	txErrors  uint64
}

func parseLinuxProcNetDev() map[string]netDevStat {
	stats := make(map[string]netDevStat)
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return stats
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		name := strings.TrimSpace(line[:colonIdx])
		fields := strings.Fields(line[colonIdx+1:])
		if len(fields) < 16 {
			continue
		}

		rxBytes, _ := strconv.ParseUint(fields[0], 10, 64)
		rxPackets, _ := strconv.ParseUint(fields[1], 10, 64)
		rxErrors, _ := strconv.ParseUint(fields[2], 10, 64)
		txBytes, _ := strconv.ParseUint(fields[8], 10, 64)
		txPackets, _ := strconv.ParseUint(fields[9], 10, 64)
		txErrors, _ := strconv.ParseUint(fields[10], 10, 64)

		stats[name] = netDevStat{
			rxBytes:   rxBytes,
			rxPackets: rxPackets,
			rxErrors:  rxErrors,
			txBytes:   txBytes,
			txPackets: txPackets,
			txErrors:  txErrors,
		}
	}
	return stats
}

func getLinuxInterfaceSpeed(name string) int64 {
	speedPath := filepath.Join("/sys/class/net", name, "speed")
	data, err := os.ReadFile(speedPath)
	if err == nil {
		if spd, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil && spd > 0 {
			return spd
		}
	}
	return 0
}

func networkSortScore(iface model.NetworkInterface) int64 {
	var score int64
	if iface.State == "up" {
		score += 1000000
	}
	if len(iface.IPv4) > 0 {
		score += 500000
	}
	// Favor physical network interfaces (eth, enp, wlan, wlp, etc.)
	if strings.HasPrefix(iface.Name, "eth") || strings.HasPrefix(iface.Name, "en") || strings.HasPrefix(iface.Name, "wl") {
		score += 200000
	}
	// Deprioritize virtual/bridge/docker interfaces
	if strings.HasPrefix(iface.Name, "docker") || strings.HasPrefix(iface.Name, "br-") || strings.HasPrefix(iface.Name, "virbr") || strings.HasPrefix(iface.Name, "veth") || strings.HasPrefix(iface.Name, "tun") || strings.HasPrefix(iface.Name, "tap") {
		score -= 100000
	}
	if iface.Name == "lo" {
		score -= 800000
	}
	traffic := (iface.RXBytes + iface.TXBytes) / (1024 * 1024)
	if traffic > 100000 {
		traffic = 100000
	}
	score += int64(traffic)
	return score
}
