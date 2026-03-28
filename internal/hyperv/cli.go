package hyperv

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/J0sh0909/rift/internal/core"
	"github.com/vbauerster/mpb/v8"
)

// HyperVBackend implements core.Hypervisor using PowerShell cmdlets.
type HyperVBackend struct {
	s core.Settings
}

func init() {
	core.RegisterBackend("hyperv", func(s core.Settings) (core.Hypervisor, error) {
		if runtime.GOOS != "windows" {
			return nil, fmt.Errorf("Hyper-V is only available on Windows")
		}
		return &HyperVBackend{s: s}, nil
	})
}

func ps(script string) (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("powershell failed: %w\nOutput: %s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

type hvVM struct {
	Name  string `json:"Name"`
	State int    `json:"State"`
}

func (h *HyperVBackend) GetPowerState() ([]core.VM, error) {
	out, err := ps(`Get-VM | Select-Object Name,State | ConvertTo-Json`)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var vms []hvVM
	if err := json.Unmarshal([]byte(out), &vms); err != nil {
		var single hvVM
		if err2 := json.Unmarshal([]byte(out), &single); err2 != nil {
			return nil, fmt.Errorf("parsing Hyper-V VM list: %w", err)
		}
		vms = []hvVM{single}
	}
	var result []core.VM
	for _, vm := range vms {
		result = append(result, core.VM{
			Name:    vm.Name,
			Path:    vm.Name,
			Running: vm.State == 2,
		})
	}
	return result, nil
}

func (h *HyperVBackend) EnsureVMwareRunning() error { return nil }

func (h *HyperVBackend) StartVM(vmName string) error {
	_, err := ps(fmt.Sprintf(`Start-VM -Name '%s'`, vmName))
	return err
}

func (h *HyperVBackend) StopVM(vmName string, mode ...string) error {
	if len(mode) > 0 && mode[0] == "hard" {
		_, err := ps(fmt.Sprintf(`Stop-VM -Name '%s' -Force`, vmName))
		return err
	}
	_, err := ps(fmt.Sprintf(`Stop-VM -Name '%s'`, vmName))
	return err
}

func (h *HyperVBackend) SuspendVM(vmName string) error {
	_, err := ps(fmt.Sprintf(`Save-VM -Name '%s'`, vmName))
	return err
}

func (h *HyperVBackend) ResetVM(vmName string) error {
	_, err := ps(fmt.Sprintf(`Restart-VM -Name '%s' -Force`, vmName))
	return err
}

func (h *HyperVBackend) RunGuestCommand(vmName, user, pass, interpreter, script, adminUser, adminPass string) (string, error) {
	return "", fmt.Errorf("guest operations not supported on Hyper-V - use PowerShell Direct or SSH")
}

func (h *HyperVBackend) RunGuestProgram(vmName, user, pass, adminUser, adminPass, program string, args ...string) (string, error) {
	return "", fmt.Errorf("guest operations not supported on Hyper-V - use PowerShell Direct or SSH")
}

func (h *HyperVBackend) CopyFileFromGuest(vmName, user, pass, adminUser, adminPass, guestPath, hostPath string) error {
	return fmt.Errorf("guest operations not supported on Hyper-V - use PowerShell Direct or SSH")
}

func (h *HyperVBackend) DeleteFileInGuest(vmName, user, pass, adminUser, adminPass, guestPath string) error {
	return fmt.Errorf("guest operations not supported on Hyper-V - use PowerShell Direct or SSH")
}

func (h *HyperVBackend) ListGuestProcesses(vmName, user, pass, adminUser, adminPass string) error {
	return fmt.Errorf("guest operations not supported on Hyper-V - use PowerShell Direct or SSH")
}

func (h *HyperVBackend) CreateSnapshot(vmName, name string) error {
	_, err := ps(fmt.Sprintf(`Checkpoint-VM -Name '%s' -SnapshotName '%s'`, vmName, name))
	return err
}

func (h *HyperVBackend) RevertToSnapshot(vmName, name string) error {
	_, err := ps(fmt.Sprintf(`Restore-VMCheckpoint -VMName '%s' -Name '%s' -Confirm:$false`, vmName, name))
	return err
}

func (h *HyperVBackend) DeleteSnapshot(vmName, name string) error {
	_, err := ps(fmt.Sprintf(`Remove-VMCheckpoint -VMName '%s' -Name '%s' -Confirm:$false`, vmName, name))
	return err
}

func (h *HyperVBackend) ListSnapshots(vmName string) ([]string, error) {
	out, err := ps(fmt.Sprintf(`Get-VMCheckpoint -VMName '%s' | Select-Object -ExpandProperty Name`, vmName))
	if err != nil {
		if strings.Contains(err.Error(), "does not have any checkpoints") {
			return nil, nil
		}
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var names []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			names = append(names, line)
		}
	}
	return names, nil
}

func (h *HyperVBackend) FindOvftool() (string, error) {
	return "", fmt.Errorf("ovftool not applicable for Hyper-V")
}

func (h *HyperVBackend) ExportVM(vmName, destPath string) error {
	_, err := ps(fmt.Sprintf(`Export-VM -Name '%s' -Path '%s'`, vmName, destPath))
	return err
}

func (h *HyperVBackend) ExportVMWithBar(vmName, destPath string, bar *mpb.Bar) error {
	err := h.ExportVM(vmName, destPath)
	if err == nil {
		bar.SetCurrent(100)
	}
	return err
}

func (h *HyperVBackend) ImportVM(srcPath, destPath string) error {
	_, err := ps(fmt.Sprintf(`Import-VM -Path '%s' -Copy -GenerateNewId`, srcPath))
	return err
}

func (h *HyperVBackend) WarmEncryptionCache(_ []string) {}
