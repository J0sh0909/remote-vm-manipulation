//go:build linux

package internal

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// DetectHostResources queries the host machine for CPU, RAM, and disk limits
// using /proc and standard POSIX utilities on Linux.
func DetectHostResources() (HostResources, error) {
	res := HostResources{FreeDiskGB: make(map[string]int)}

	// RAM from /proc/meminfo (MemTotal is reported in kB)
	if f, err := os.Open("/proc/meminfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					kb, _ := strconv.ParseInt(fields[1], 10, 64)
					res.TotalRAMGB = int(kb / 1024 / 1024)
				}
				break
			}
		}
	}

	// CPU threads: nproc reports logical processors
	if out, err := exec.Command("nproc").Output(); err == nil {
		res.CPUThreads, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	// CPU physical cores: count unique (physical id, core id) pairs in /proc/cpuinfo
	if f, err := os.Open("/proc/cpuinfo"); err == nil {
		defer f.Close()
		type cpuKey struct{ physID, coreID string }
		seen := make(map[cpuKey]bool)
		var physID, coreID string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "physical id") {
				if parts := strings.SplitN(line, ":", 2); len(parts) == 2 {
					physID = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "core id") {
				if parts := strings.SplitN(line, ":", 2); len(parts) == 2 {
					coreID = strings.TrimSpace(parts[1])
				}
			} else if line == "" && physID != "" && coreID != "" {
				seen[cpuKey{physID, coreID}] = true
				physID, coreID = "", ""
			}
		}
		res.CPUCores = len(seen)
		if res.CPUCores == 0 {
			// Fallback for VMs or single-socket machines that omit "physical id"
			res.CPUCores = res.CPUThreads
		}
	}

	// Disk free space: df -BG reports sizes in whole GB; column 4 is avail, column 6 is mountpoint
	if out, err := exec.Command("df", "-BG").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines[1:] { // skip header
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				mountPoint := fields[5]
				freeStr := strings.TrimSuffix(fields[3], "G")
				if free, err := strconv.Atoi(freeStr); err == nil && free > 0 {
					res.FreeDiskGB[mountPoint] = free
				}
			}
		}
	}

	return res, nil
}
