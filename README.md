# rift

A cross-hypervisor VM orchestration CLI built in Go.

## Features

- **VM lifecycle management** - start, stop, suspend, reset, and query power state
- **Parallel folder operations** - target all VMs in a VMware folder with -f <folder>, running eligible operations concurrently
- **Guest exec with OS auto-detection** - run commands inside VMs; interpreter is inferred from the VMX guestOS key (Linux -> /bin/bash, Windows -> cmd.exe)
- **Bootstrap provisioning** - provision the runner automation user on guest VMs in one command
- **Snapshot management** - create (with duplicate detection), list, revert, and delete snapshots
- **Cross-hypervisor migration** - migrate VMs between VMware Workstation and VirtualBox with rift migrate
- **OVF/OVA archive pipeline** - export and import VMs via ovftool with per-VM progress bars
- **Hardware configuration** - edit CPU, RAM, NIC, disk, CD/DVD, and display settings
- **Structured error codes** - every failure prints a [VMxxx] code; rift errors lists all codes
- **GitHub Actions pipeline** - trigger any rift command against self-hosted runners via workflow_dispatch
- **AWS EC2 management** - create, start, stop, terminate, and SSH into EC2 instances
- **VirtualBox management** - list, start, stop, snapshot, and delete VirtualBox VMs via rift vbox
- **Cross-platform host support** - compiles and runs on both Windows and Linux hosts
- **Dual-binary architecture** - separate rift.exe (Windows) and rift-linux (Linux/WSL) binaries

## Supported Platforms

| Platform | Scope | Lifecycle | Snapshots | Exec | Config | Export | Migrate | SSH |
|---|---|---|---|---|---|---|---|---|
| VMware Workstation | rift <cmd> | start, stop, suspend, reset | create, list, revert, delete | guest exec | cpu, ram, nic, disk, cdrom, display, os | ovf/ova | to/from vbox | - |
| VirtualBox | rift vbox <cmd> | start, stop | take, list, restore, delete | - | - | - | to/from vmware | - |
| AWS EC2 | rift aws <cmd> | start, stop, create, terminate | - | - | - | - | - | ssh, ip |
| Hyper-V | detection only | - | - | - | - | - | - | - |
| Proxmox VE | planned | - | - | - | - | - | - | - |

## WSL/Windows Interoperability

Rift supports both native Windows execution and WSL/Linux execution through a dual-binary approach:

- **Windows Host**: Use rift.exe with .env file containing Windows paths
- **WSL/Linux Host**: Use rift-linux with .env.wsl file containing WSL paths
- **Limitation**: Some commands requiring file access to Windows user directories may not work when calling rift.exe from WSL
- **Workaround**: Use rift-linux binary within WSL with appropriate .env.wsl configuration

## Setup

```bash
git clone https://github.com/J0sh0909/remote-vm-manipulation
cd remote-vm-manipulation
```

### Environment Configuration

Create .env (Windows) or .env.wsl (WSL/Linux) files with your configuration:

```bash
VMRUN_PATH=path/to/vmrun
VM_DIRECTORY=path/to/vms
INVENTORY_PATH=path/to/inventory.vmls
NETMAP_PATH=path/to/netmap.conf
ISO_DIRECTORY=path/to/isos
VDISK_PATH=path/to/vdiskmanager
ARCHIVE_PATH=path/to/archive
QEMU_IMG_PATH=path/to/qemu-img
VM_DEFAULT_USER=runner
VM_DEFAULT_PASS=your_password
HYPERVISOR=workstation
AWS_REGION=us-east-1
AWS_KEY_DIR=path/to/ssh/keys
```

### Building

```bash
# Windows
go build -o rift.exe .

# Linux/WSL
go build -o rift-linux .
```

## Architecture

Rift is built around a Hypervisor interface that abstracts backend operations:

- internal/vmware/ - VMware Workstation backend
- internal/vbox/ - VirtualBox backend
- internal/aws/ - AWS EC2 backend
- internal/hyperv/ - Hyper-V detection
- internal/proxmox/ - Proxmox VE (planned)

## License

MIT License - Copyright 2026 J0sh0909
