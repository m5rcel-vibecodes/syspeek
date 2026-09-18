//go:build darwin

package collector

import (
	"bufio"
	"net"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

// CollectNetwork collects network interface metrics on macOS.
func CollectNetwork() ([]model.NetworkInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	statsMap := parseDarwinNetstatIB()

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
				if ip.IsLoopback() && iface.Name != "lo0" {
					continue
				}
				if ip.To4() != nil {
					ipv4s = append(ipv4s, ip.String())
				} else if ip.To16() != nil {
					ipv6s = append(ipv6s, ip.String())
				}
			}
		}

		stat := statsMap[iface.Name]

		result = append(result, model.NetworkInterface{
			Name:      iface.Name,
			State:     state,
			IPv4:      ipv4s,
			IPv6:      ipv6s,
			MAC:       iface.HardwareAddr.String(),
			RXBytes:   stat.rxBytes,
			TXBytes:   stat.txBytes,
			RXPackets: stat.rxPackets,
			TXPackets: stat.txPackets,
			RXErrors:  stat.rxErrors,
			TXErrors:  stat.txErrors,
		})
	}

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

type darwinNetStat struct {
	rxBytes   uint64
	rxPackets uint64
	rxErrors  uint64
	txBytes   uint64
	txPackets uint64
	txErrors  uint64
}

func parseDarwinNetstatIB() map[string]darwinNetStat {
	stats := make(map[string]darwinNetStat)
	cmd := exec.Command("netstat", "-ib")
	out, err := cmd.Output()
	if err != nil {
		return stats
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	// Skip header line
	if scanner.Scan() {
		_ = scanner.Text()
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		name := fields[0]
		// Skip asterisk marked interfaces or strip *
		name = strings.TrimSuffix(name, "*")

		// Only parse the <Link#...> line where hardware link bytes are reported
		network := fields[2]
		if !strings.HasPrefix(network, "<Link#") && fields[1] != "16384" {
			// For lo0, network is "127" or "localhost", check if not already recorded
			if _, exists := stats[name]; exists {
				continue
			}
		}

		// When address column is present vs not present:
		// Format 1 (with Address e.g. MAC): Name, Mtu, Network, Address, Ipkts, Ierrs, Ibytes, Opkts, Oerrs, Obytes, Coll (11 fields)
		// Format 2 (without Address): Name, Mtu, Network, Ipkts, Ierrs, Ibytes, Opkts, Oerrs, Obytes, Coll (10 fields)
		var ibytesIdx, obytesIdx, ipktsIdx, opktsIdx, ierrsIdx, oerrsIdx int
		if len(fields) >= 11 {
			ipktsIdx = 4
			ierrsIdx = 5
			ibytesIdx = 6
			opktsIdx = 7
			oerrsIdx = 8
			obytesIdx = 9
		} else {
			ipktsIdx = 3
			ierrsIdx = 4
			ibytesIdx = 5
			opktsIdx = 6
			oerrsIdx = 7
			obytesIdx = 8
		}

		rxBytes, _ := strconv.ParseUint(fields[ibytesIdx], 10, 64)
		txBytes, _ := strconv.ParseUint(fields[obytesIdx], 10, 64)
		rxPackets, _ := strconv.ParseUint(fields[ipktsIdx], 10, 64)
		txPackets, _ := strconv.ParseUint(fields[opktsIdx], 10, 64)
		rxErrors, _ := strconv.ParseUint(fields[ierrsIdx], 10, 64)
		txErrors, _ := strconv.ParseUint(fields[oerrsIdx], 10, 64)

		// Record if higher than existing
		existing := stats[name]
		if rxBytes >= existing.rxBytes && txBytes >= existing.txBytes {
			stats[name] = darwinNetStat{
				rxBytes:   rxBytes,
				txBytes:   txBytes,
				rxPackets: rxPackets,
				txPackets: txPackets,
				rxErrors:  rxErrors,
				txErrors:  txErrors,
			}
		}
	}

	return stats
}

func networkSortScore(iface model.NetworkInterface) int64 {
	var score int64
	if iface.State == "up" {
		score += 1000000
	}
	if len(iface.IPv4) > 0 {
		score += 500000
	}
	// Favor physical network interfaces (en0, en1, etc.)
	if strings.HasPrefix(iface.Name, "en") || strings.HasPrefix(iface.Name, "eth") || strings.HasPrefix(iface.Name, "wl") {
		score += 200000
	}
	// Deprioritize virtual/bridge/tunnel interfaces
	if strings.HasPrefix(iface.Name, "bridge") || strings.HasPrefix(iface.Name, "utun") || strings.HasPrefix(iface.Name, "vmenet") || strings.HasPrefix(iface.Name, "docker") || strings.HasPrefix(iface.Name, "veth") {
		score -= 100000
	}
	if iface.Name == "lo0" || iface.Name == "lo" {
		score -= 800000
	}
	traffic := (iface.RXBytes + iface.TXBytes) / (1024 * 1024)
	if traffic > 100000 {
		traffic = 100000
	}
	score += int64(traffic)
	return score
}
