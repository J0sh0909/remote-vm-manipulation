package core

// Test commit to trigger CI after adding new file.
import (
	"fmt"
	"os"
	"os/exec"

	"github.com/vbauerster/mpb/v8"
)

// Hypervisor abstracts VM management operations across different backends.
type Hypervisor interface {
	// GetPowerState returns all VMs with their running state.
	GetPowerState() ([]VM, error)

	// EnsureVMwareRunning ensures the hypervisor application is running before
	// VMs are started. No-op on non-Windows platforms and non-Workstation backends.
	EnsureVMwareRunning() error

	// Power operations
	StartVM(vmxPath string) error
	StopVM(vmxPath string, mode ...string) error
	SuspendVM(vmxPath string) error
	ResetVM(vmxPath string) error

	// Guest operations
	// adminUser/adminPass are optional fallback credentials for hostname-prefixed
	// auth retry on Windows guests. Pass empty strings to skip the retry.
	RunGuestCommand(vmxPath, user, pass, interpreter, script, adminUser, adminPass string) (string, error)
	RunGuestProgram(vmxPath, user, pass, adminUser, adminPass, program string, args ...string) (string, error)
	CopyFileFromGuest(vmxPath, user, pass, adminUser, adminPass, guestPath, hostPath string) error
	DeleteFileInGuest(vmxPath, user, pass, adminUser, adminPass, guestPath string) error
	ListGuestProcesses(vmxPath, user, pass, adminUser, adminPass string) error

	// Snapshot operations
	CreateSnapshot(vmxPath, name string) error
	RevertToSnapshot(vmxPath, name string) error
	DeleteSnapshot(vmxPath, name string) error
	ListSnapshots(vmxPath string) ([]string, error)

	// Archive operations
	FindOvftool() (string, error)
	ExportVM(vmxPath, destPath string) error
	ExportVMWithBar(vmxPath, destPath string, bar *mpb.Bar) error
	ImportVM(srcPath, destVmxPath string) error

	// WarmEncryptionCache pre-reads VMX files sequentially so that parallel
	// power operations don't all hit the filesystem at once.
	WarmEncryptionCache(vmxPaths []string)
}

// HypervisorFactory is a function that creates a Hypervisor from Settings.
// Backends register themselves via RegisterBackend.
type HypervisorFactory func(s Settings) (Hypervisor, error)

var backendFactories = map[string]HypervisorFactory{}

// RegisterBackend registers a backend factory by name.
func RegisterBackend(name string, factory HypervisorFactory) {
	backendFactories[name] = factory
}

// NewHypervisor creates a Hypervisor from settings, with auto-detection.
// If hvFlag is non-empty, it overrides both the .env HYPERVISOR and auto-detection.
func NewHypervisor(s Settings, hvFlag string) (Hypervisor, error) {
	// Step 1: determine which backend to use
	backend := s.Hypervisor
	if hvFlag != "" {
		backend = hvFlag
	}

	// Step 2: if explicitly set, use that backend directly
	if backend != "" {
		factory, ok := backendFactories[backend]
		if !ok {
			return nil, fmt.Errorf("unknown hypervisor backend %q", backend)
		}
		return factory(s)
	}

	// Step 3: auto-detect installed hypervisors
	var detected []string

	// VMware: VMRUN_PATH is set and the file exists
	if s.VmrunPath != "" {
		if _, err := os.Stat(s.VmrunPath); err == nil {
			detected = append(detected, "workstation")
		}
	}

	// VirtualBox: VBOX_MANAGE_PATH is set and exists, OR VBoxManage is in PATH
	if s.VBoxManagePath != "" {
		if _, err := os.Stat(s.VBoxManagePath); err == nil {
			detected = append(detected, "vbox")
		}
	} else if _, err := exec.LookPath("VBoxManage"); err == nil {
		detected = append(detected, "vbox")
	}

	// Hyper-V: only when explicitly enabled via HYPERV_ENABLED=true in .env
	if s.HyperVEnabled {
		detected = append(detected, "hyperv")
	}

	// Store detected hypervisors on settings for overview iteration
	s.DetectedHypervisors = detected

	switch len(detected) {
	case 0:
		return nil, fmt.Errorf("no hypervisor configured - check .env")
	case 1:
		factory, ok := backendFactories[detected[0]]
		if !ok {
			return nil, fmt.Errorf("backend %q detected but not registered", detected[0])
		}
		return factory(s)
	default:
		return nil, fmt.Errorf("multiple hypervisors detected (%s) - use --hv flag or set HYPERVISOR in .env", joinNames(detected))
	}
}

// DetectHypervisors returns the list of detected hypervisor backend names.
// Used by overview to iterate all detected backends.
func DetectHypervisors(s Settings) []string {
	var detected []string

	if s.VmrunPath != "" {
		if _, err := os.Stat(s.VmrunPath); err == nil {
			detected = append(detected, "workstation")
		}
	}

	if s.VBoxManagePath != "" {
		if _, err := os.Stat(s.VBoxManagePath); err == nil {
			detected = append(detected, "vbox")
		}
	} else if _, err := exec.LookPath("VBoxManage"); err == nil {
		detected = append(detected, "vbox")
	}

	// Hyper-V: only when explicitly enabled via HYPERV_ENABLED=true in .env
	if s.HyperVEnabled {
		detected = append(detected, "hyperv")
	}

	return detected
}

// CreateBackend creates a backend by name from settings without auto-detection.
func CreateBackend(name string, s Settings) (Hypervisor, error) {
	factory, ok := backendFactories[name]
	if !ok {
		return nil, fmt.Errorf("unknown hypervisor backend %q", name)
	}
	return factory(s)
}

func joinNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	result := names[0]
	for _, n := range names[1:] {
		result += ", " + n
	}
	return result
}
