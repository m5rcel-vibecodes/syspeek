//go:build linux

package collector

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/m5rcel-vibecodes/syspeek/internal/model"
)

var (
	cachedUIDs = make(map[string]string)
	pageSize   = uint64(os.Getpagesize())
)

func init() {
	if pageSize == 0 {
		pageSize = 4096
	}
	loadEtcPasswd()
}

func loadEtcPasswd() {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			name := parts[0]
			uid := parts[2]
			cachedUIDs[uid] = name
		}
	}
}

// CollectProcesses gathers top processes by memory/activity on Linux.
func CollectProcesses(limit int) ([]model.ProcessInfo, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	memInfo, _ := CollectMemory()
	var totalRAM uint64
	if memInfo != nil {
		totalRAM = memInfo.TotalBytes
	}
	if totalRAM == 0 {
		totalRAM = 1
	}

	var procs []model.ProcessInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue // Not a PID folder
		}

		procDir := filepath.Join("/proc", entry.Name())

		// Read /proc/[pid]/stat
		statBytes, err := os.ReadFile(filepath.Join(procDir, "stat"))
		if err != nil {
			continue
		}

		firstParen := bytes.IndexByte(statBytes, '(')
		lastParen := bytes.LastIndexByte(statBytes, ')')
		if firstParen == -1 || lastParen == -1 || lastParen <= firstParen {
			continue
		}

		comm := string(statBytes[firstParen+1 : lastParen])
		rest := strings.Fields(string(statBytes[lastParen+1:]))
		if len(rest) < 22 {
			continue
		}

		state := rest[0]
		rssPages, _ := strconv.ParseUint(rest[21], 10, 64)
		rssBytes := rssPages * pageSize

		// Read /proc/[pid]/cmdline
		cmdlineBytes, _ := os.ReadFile(filepath.Join(procDir, "cmdline"))
		cmd := strings.TrimSpace(string(bytes.ReplaceAll(cmdlineBytes, []byte{0}, []byte(" "))))
		if cmd == "" {
			cmd = "[" + comm + "]"
		}

		// Read UID from /proc/[pid]/status
		user := "root"
		if statusBytes, err := os.ReadFile(filepath.Join(procDir, "status")); err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(statusBytes))
			for scanner.Scan() {
				sLine := scanner.Text()
				if strings.HasPrefix(sLine, "Uid:") {
					uidFields := strings.Fields(sLine)
					if len(uidFields) >= 2 {
						uid := uidFields[1]
						if name, ok := cachedUIDs[uid]; ok {
							user = name
						} else {
							user = uid
						}
					}
					break
				}
			}
		}

		memPct := (float64(rssBytes) / float64(totalRAM)) * 100.0

		procs = append(procs, model.ProcessInfo{
			PID:           pid,
			User:          user,
			CPUPercent:    0, // Will be enriched or calculated
			MemoryBytes:   rssBytes,
			MemoryPercent: memPct,
			Command:       cmd,
			Status:        state,
		})
	}

	// Sort by RSS memory descending
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].MemoryBytes > procs[j].MemoryBytes
	})

	if limit > 0 && len(procs) > limit {
		procs = procs[:limit]
	}

	return procs, nil
}
