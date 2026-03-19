//go:build windows

package internal

import (
	"os/exec"
	"strconv"
	"strings"
)

// DetectHostResources queries the host machine for CPU, RAM, and disk limits
// using PowerShell/WMI on Windows.
func DetectHostResources() (HostResources, error) {
	res := HostResources{FreeDiskGB: make(map[string]int)}

	// RAM
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"(Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory").CombinedOutput()
	if err == nil {
		val := strings.TrimSpace(string(out))
		bytes, _ := strconv.ParseInt(val, 10, 64)
		if bytes > 0 {
			res.TotalRAMGB = int(bytes / 1024 / 1024 / 1024)
		}
	}

	// CPU cores
	out, err = exec.Command("powershell", "-NoProfile", "-Command",
		"(Get-CimInstance Win32_Processor).NumberOfCores").CombinedOutput()
	if err == nil {
		res.CPUCores, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	// CPU threads
	out, err = exec.Command("powershell", "-NoProfile", "-Command",
		"(Get-CimInstance Win32_Processor).NumberOfLogicalProcessors").CombinedOutput()
	if err == nil {
		res.CPUThreads, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	// Disk free space
	out, err = exec.Command("powershell", "-NoProfile", "-Command",
		"Get-CimInstance Win32_LogicalDisk | ForEach-Object { $_.DeviceID + '=' + $_.FreeSpace }").CombinedOutput()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				drive := strings.TrimSpace(parts[0])
				bytes, _ := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if bytes > 0 && drive != "" {
					res.FreeDiskGB[drive] = int(bytes / 1024 / 1024 / 1024)
				}
			}
		}
	}

	return res, nil
}
