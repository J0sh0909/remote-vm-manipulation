package vbox

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/J0sh0909/rift/internal/core"
	"github.com/vbauerster/mpb/v8"
)

func init() {
	core.RegisterBackend("vbox", func(s core.Settings) (core.Hypervisor, error) {
		path := s.VBoxManagePath
		if path == "" {
			p, err := exec.LookPath("VBoxManage")
			if err != nil {
				return nil, fmt.Errorf("VBoxManage not found in PATH")
			}
			path = p
		}
		return &VBoxBackend{vboxManage: path}, nil
	})
}

// VBoxBackend wraps VBoxManage CLI commands for VirtualBox VM management.
type VBoxBackend struct {
	vboxManage string // path to VBoxManage
}

// VBoxVMInfo holds detailed info about a VirtualBox VM.
type VBoxVMInfo struct {
	Name     string
	UUID     string
	OSType   string
	CPUs     int
	MemoryMB int
	State    string
	NICs     []VBoxNIC
	Disks    []string // disk file paths
}

// VBoxNIC describes a VirtualBox network adapter.
type VBoxNIC struct {
	Index   int
	Type    string // connection type: nat, bridged, hostonly, intnet, none
	NICType string // hardware type: virtio, 82540EM, 82545EM, Am79C970A, Am79C973
}

// VBoxVM holds minimal info about a VirtualBox VM.
type VBoxVM struct {
	Name  string
	State string
}

// NewVBoxBackend creates a VBoxBackend, locating VBoxManage in PATH.
func NewVBoxBackend() (*VBoxBackend, error) {
	path, err := exec.LookPath("VBoxManage")
	if err != nil {
		return nil, fmt.Errorf("VBoxManage not found in PATH")
	}
	return &VBoxBackend{vboxManage: path}, nil
}

// VBoxManagePath returns the path to VBoxManage.
func (v *VBoxBackend) VBoxManagePath() string { return v.vboxManage }

func (v *VBoxBackend) run(args ...string) (string, error) {
	cmd := exec.Command(v.vboxManage, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("VBoxManage %v failed: %w\nOutput: %s", args, err, out)
	}
	return string(out), nil
}

// ListVMs returns all registered VMs with running state.
func (v *VBoxBackend) ListVMs() ([]VBoxVM, error) {
	allOut, err := v.run("list", "vms")
	if err != nil {
		return nil, err
	}
	runningOut, _ := v.run("list", "runningvms")
	runningSet := make(map[string]bool)
	reUUID := regexp.MustCompile(`\{(.+?)\}`)
	for _, line := range strings.Split(runningOut, "\n") {
		if m := reUUID.FindStringSubmatch(line); m != nil {
			runningSet[m[1]] = true
		}
	}

	re := regexp.MustCompile(`"(.+?)"\s+\{(.+?)\}`)
	var vms []VBoxVM
	for _, line := range strings.Split(allOut, "\n") {
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		state := "poweroff"
		if runningSet[m[2]] {
			state = "running"
		}
		vms = append(vms, VBoxVM{Name: m[1], State: state})
	}
	return vms, nil
}

// DirectStartVM starts a VM in headless mode.
func (v *VBoxBackend) DirectStartVM(nameOrUUID string) error {
	_, err := v.run("startvm", nameOrUUID, "--type", "headless")
	return err
}

// DirectStopVM stops a VM. If hard is true, uses poweroff; otherwise acpipowerbutton.
func (v *VBoxBackend) DirectStopVM(nameOrUUID string, hard bool) error {
	action := "acpipowerbutton"
	if hard {
		action = "poweroff"
	}
	_, err := v.run("controlvm", nameOrUUID, action)
	return err
}

// GetVMInfo returns detailed info about a VM by parsing showvminfo output.
func (v *VBoxBackend) GetVMInfo(nameOrUUID string) (VBoxVMInfo, error) {
	out, err := v.run("showvminfo", nameOrUUID, "--machinereadable")
	if err != nil {
		return VBoxVMInfo{}, err
	}
	info := VBoxVMInfo{}
	data := parseMachineReadable(out)

	info.Name = data["name"]
	info.UUID = data["UUID"]
	info.OSType = data["ostype"]
	info.State = data["VMState"]
	fmt.Sscanf(data["cpus"], "%d", &info.CPUs)
	fmt.Sscanf(data["memory"], "%d", &info.MemoryMB)

	// Parse NICs (up to 8).
	for i := 1; i <= 8; i++ {
		nicType := data[fmt.Sprintf("nic%d", i)]
		if nicType == "" || nicType == "none" {
			continue
		}
		hwType := data[fmt.Sprintf("nictype%d", i)]
		info.NICs = append(info.NICs, VBoxNIC{Index: i, Type: nicType, NICType: hwType})
	}

	// Collect disk paths from SATA/IDE/SCSI attachments.
	for k, val := range data {
		if strings.Contains(k, "-imageuuid-") || strings.Contains(k, "-ImageUUID-") {
			continue
		}
		if (strings.HasPrefix(k, "\"sata-") || strings.HasPrefix(k, "\"ide-") ||
			strings.HasPrefix(k, "\"scsi-") || strings.HasPrefix(k, "\"nvme-")) &&
			val != "none" && val != "emptydrive" && val != "" {
			info.Disks = append(info.Disks, val)
		}
		// Also match unquoted keys.
		if (strings.HasPrefix(k, "sata-") || strings.HasPrefix(k, "ide-") ||
			strings.HasPrefix(k, "scsi-") || strings.HasPrefix(k, "nvme-")) &&
			!strings.Contains(k, "uuid") && !strings.Contains(k, "UUID") &&
			val != "none" && val != "emptydrive" && val != "" {
			if strings.HasSuffix(val, ".vdi") || strings.HasSuffix(val, ".vmdk") ||
				strings.HasSuffix(val, ".vhd") {
				info.Disks = append(info.Disks, val)
			}
		}
	}

	return info, nil
}

// ImportOVF imports an OVF/OVA file.
func (v *VBoxBackend) ImportOVF(ovfPath string) error {
	_, err := v.run("import", ovfPath)
	return err
}

// ExportOVF exports a VM to an OVF/OVA file.
func (v *VBoxBackend) ExportOVF(nameOrUUID, outputPath string) error {
	_, err := v.run("export", nameOrUUID, "-o", outputPath)
	return err
}

// CreateVM creates and registers a new VM with the given specs.
func (v *VBoxBackend) CreateVM(name, osType string, cpus, ramMB int) error {
	if _, err := v.run("createvm", "--name", name, "--ostype", osType, "--register"); err != nil {
		return err
	}
	_, err := v.run("modifyvm", name,
		"--cpus", fmt.Sprintf("%d", cpus),
		"--memory", fmt.Sprintf("%d", ramMB),
	)
	return err
}

// AddSATAController adds a SATA controller to a VM.
func (v *VBoxBackend) AddSATAController(vmName, ctlName string) error {
	_, err := v.run("storagectl", vmName, "--name", ctlName, "--add", "sata", "--controller", "IntelAhci")
	return err
}

// AttachDisk attaches a disk image to a SATA controller port.
func (v *VBoxBackend) AttachDisk(vmName, diskPath, controller string) error {
	_, err := v.run("storageattach", vmName,
		"--storagectl", controller,
		"--port", "0",
		"--device", "0",
		"--type", "hdd",
		"--medium", diskPath,
	)
	return err
}

// SnapshotCreate takes a snapshot.
func (v *VBoxBackend) SnapshotCreate(vmName, snapName string) error {
	_, err := v.run("snapshot", vmName, "take", snapName)
	return err
}

// SnapshotList returns snapshot names.
func (v *VBoxBackend) SnapshotList(vmName string) ([]string, error) {
	out, err := v.run("snapshot", vmName, "list")
	if err != nil {
		if strings.Contains(err.Error(), "does not have any snapshots") {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	re := regexp.MustCompile(`Name:\s+(.+?)\s+\(UUID:`)
	for _, m := range re.FindAllStringSubmatch(out, -1) {
		names = append(names, m[1])
	}
	return names, nil
}

// SnapshotRevert restores a snapshot.
func (v *VBoxBackend) SnapshotRevert(vmName, snapName string) error {
	_, err := v.run("snapshot", vmName, "restore", snapName)
	return err
}

// SnapshotDelete deletes a snapshot.
func (v *VBoxBackend) SnapshotDelete(vmName, snapName string) error {
	_, err := v.run("snapshot", vmName, "delete", snapName)
	return err
}

// DeleteVM unregisters and deletes a VM and its files.
func (v *VBoxBackend) DeleteVM(nameOrUUID string) error {
	_, err := v.run("unregistervm", nameOrUUID, "--delete")
	return err
}

// parseMachineReadable parses VBoxManage --machinereadable output into a map.
func parseMachineReadable(output string) map[string]string {
	data := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := line[:idx]
		val := strings.Trim(line[idx+1:], "\"")
		data[key] = val
	}
	return data
}

// ---------------------------------------------------------------------------
// core.Hypervisor interface implementation
// ---------------------------------------------------------------------------

func (v *VBoxBackend) GetPowerState() ([]core.VM, error) {
	vms, err := v.ListVMs()
	if err != nil {
		return nil, err
	}
	var result []core.VM
	for _, vm := range vms {
		result = append(result, core.VM{
			Name:    vm.Name,
			Path:    vm.Name, // VBox uses names as identifiers
			Running: vm.State == "running",
		})
	}
	return result, nil
}

func (v *VBoxBackend) EnsureVMwareRunning() error { return nil }

// StartVM implements core.Hypervisor - delegates to headless start.
func (v *VBoxBackend) StartVM(vmxPath string) error {
	_, err := v.run("startvm", vmxPath, "--type", "headless")
	return err
}

// StopVM implements core.Hypervisor.
func (v *VBoxBackend) StopVM(vmxPath string, mode ...string) error {
	action := "acpipowerbutton"
	if len(mode) > 0 && mode[0] == "hard" {
		action = "poweroff"
	}
	_, err := v.run("controlvm", vmxPath, action)
	return err
}

func (v *VBoxBackend) SuspendVM(vmxPath string) error {
	_, err := v.run("controlvm", vmxPath, "savestate")
	return err
}

func (v *VBoxBackend) ResetVM(vmxPath string) error {
	_, err := v.run("controlvm", vmxPath, "reset")
	return err
}

func (v *VBoxBackend) RunGuestCommand(vmxPath, user, pass, interpreter, script, adminUser, adminPass string) (string, error) {
	return "", fmt.Errorf("not supported on VirtualBox - use SSH")
}

func (v *VBoxBackend) RunGuestProgram(vmxPath, user, pass, adminUser, adminPass, program string, args ...string) (string, error) {
	return "", fmt.Errorf("not supported on VirtualBox - use SSH")
}

func (v *VBoxBackend) CopyFileFromGuest(vmxPath, user, pass, adminUser, adminPass, guestPath, hostPath string) error {
	return fmt.Errorf("not supported on VirtualBox - use SSH")
}

func (v *VBoxBackend) DeleteFileInGuest(vmxPath, user, pass, adminUser, adminPass, guestPath string) error {
	return fmt.Errorf("not supported on VirtualBox - use SSH")
}

func (v *VBoxBackend) ListGuestProcesses(vmxPath, user, pass, adminUser, adminPass string) error {
	return fmt.Errorf("not supported on VirtualBox - use SSH")
}

func (v *VBoxBackend) CreateSnapshot(vmxPath, name string) error {
	_, err := v.run("snapshot", vmxPath, "take", name)
	return err
}

func (v *VBoxBackend) RevertToSnapshot(vmxPath, name string) error {
	_, err := v.run("snapshot", vmxPath, "restore", name)
	return err
}

func (v *VBoxBackend) DeleteSnapshot(vmxPath, name string) error {
	_, err := v.run("snapshot", vmxPath, "delete", name)
	return err
}

func (v *VBoxBackend) ListSnapshots(vmxPath string) ([]string, error) {
	return v.SnapshotList(vmxPath)
}

func (v *VBoxBackend) FindOvftool() (string, error) {
	return "", fmt.Errorf("ovftool not applicable for VirtualBox - use VBoxManage export/import")
}

func (v *VBoxBackend) ExportVM(vmxPath, destPath string) error {
	return v.ExportOVF(vmxPath, destPath)
}

func (v *VBoxBackend) ExportVMWithBar(vmxPath, destPath string, bar *mpb.Bar) error {
	err := v.ExportOVF(vmxPath, destPath)
	if err == nil {
		bar.SetCurrent(100)
	}
	return err
}

func (v *VBoxBackend) ImportVM(srcPath, destVmxPath string) error {
	return v.ImportOVF(srcPath)
}

func (v *VBoxBackend) WarmEncryptionCache(_ []string) {}
