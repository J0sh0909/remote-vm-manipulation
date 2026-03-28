package hyperv

import (
	"encoding/json"
	"os/exec"
	"runtime"
)

type HyperVVM struct {
	Name  string `json:"Name"`
	State int    `json:"State"`
}

func HyperVStateName(state int) string {
	switch state {
	case 2:
		return "running"
	case 3:
		return "off"
	case 6:
		return "saved"
	case 9:
		return "paused"
	case 10:
		return "starting"
	case 11:
		return "resetting"
	case 4:
		return "shutting down"
	default:
		return "unknown"
	}
}

func DetectHyperVVMs() ([]HyperVVM, error) {
	if runtime.GOOS != "windows" {
		return nil, nil
	}
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		`Get-VM | Select-Object Name,State | ConvertTo-Json`).Output()
	if err != nil {
		return nil, nil
	}
	trimmed := string(out)
	if len(trimmed) == 0 {
		return nil, nil
	}
	var vms []HyperVVM
	if err := json.Unmarshal(out, &vms); err != nil {
		var single HyperVVM
		if err2 := json.Unmarshal(out, &single); err2 != nil {
			return nil, nil
		}
		vms = []HyperVVM{single}
	}
	return vms, nil
}
