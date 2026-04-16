RIFT — Complete Project Roadmap


v1.0.0 — VMware Workstation CLI (DONE)

VM lifecycle: start, stop, suspend, reset, restart
Parallel folder operations with goroutines and retry logic
Guest exec with OS auto-detection (80+ guest OS interpreters)
Windows exec via runProgramInGuest with cmd.exe
Linux exec via runScriptInGuest with output capture
Snapshot management with duplicate detection
OVF/OVA archive pipeline with mpb progress bars
Hardware config: CPU, RAM, NIC, disk, display, CD/DVD, TPM
VM encryption support with auto-detection and -vp flag
Structured error codes (VM1xx through VM9xx)
.env based configuration
rift errors reference command
rift info, rift list, rift networks, rift isos
Bootstrap companion tool integration (create, verify, reset, revoke)
Cross-platform: Windows and Linux host support
GitHub Actions pipeline with self-hosted runner
Build script with Windows/Linux cross-compilation


v1.1.0 — VirtualBox CLI Backend (DONE)

VBox lifecycle: list, start, stop, info, delete
VBox snapshots: create, list, revert, delete
VBox import/export OVF
VBox on the Hypervisor interface (same rift list works)
VBox detection for overview
rift vbox subcommands as shortcuts


v1.2.0 — Cross-Hypervisor Migration (DONE)

rift migrate <vm> --from vmware --to vbox and vice versa
Direct disk conversion via qemu-img (VMDK↔VDI, no OVF intermediate)
Config translation: CPU, RAM, guest OS, NIC type mapping
Folder migration with --folder flag and parallel mpb progress bars
Archive naming with hypervisor identifier (-vmw-, -vbox-)


v1.3.0 — AWS EC2 Backend (DONE)

rift aws create with auto key pair, security group, elastic IP
rift aws list, rift aws start, rift aws stop, rift aws terminate
rift aws ssh with auto user detection (ubuntu, ec2-user, admin)
rift aws ip for quick IP lookup
rift aws state for resource tracking via rift-state.json
rift aws destroy --yes for full cleanup (instances, EIPs, key pairs, security groups)
--count flag for mass creation with parallel goroutines
--key flag for key pair reuse


v1.4.0 — Unified Overview and Multi-Node (DONE)

rift overview showing VMware, VirtualBox, Hyper-V, and AWS in one view
Auto-detection of installed hypervisors from .env
--hv flag to override default hypervisor
Multi-section .env with all hypervisor configs
Auto-select when 1 hypervisor detected, list all on list/overview, require --hv on mutating commands when 2+ detected
Hyper-V detection stub
VBox and Hyper-V detection for overview
Multi-runner pipeline: build on Strong, distribute to Weak and Tank
Manual dispatch runs on all three machines simultaneously
Three-node infrastructure: Omen (VMware), Pavilion (VBox), AI (Hyper-V)


v1.5.0 — JSON Output and Machine-Readable Interface

--json flag on every command
All commands return structured JSON when --json is set
Error responses in JSON format with error codes
Enables AI and automation consumers to parse rift output
rift --json list returns {"hypervisor": "workstation", "vms": [{"name": "...", "state": "running"}]}
rift --json overview returns aggregated JSON from all detected hypervisors
Prerequisite for REST API and OpenClaw integration
Interim AI workflow: OpenClaw can shell out to rift --json before full API exists


v2.0.0 — REST API and OpenClaw Integration (rift serve)

rift serve --port 8080 starts HTTP server on each node
Endpoints mirror CLI commands:

GET  /api/list         → list VMs
POST /api/start        → start VM
POST /api/stop         → stop VM
POST /api/exec         → execute command in guest
GET  /api/overview     → local hypervisor overview
POST /api/create       → create VM from JSON payload
POST /api/migrate      → start migration
GET  /api/snapshots    → list snapshots
GET  /api/capabilities → available backends, supported operations per backend, compute types

API token authentication from .env
JSON request/response format
Runs on Tailscale IP for secure mesh access
Each node exposes its local hypervisor
Designed for AI agent consumption:
  Structured responses with state, progress, and capabilities metadata
  Idempotent operations where possible
  Rich error responses with actionable context and suggested next steps
  Discovery endpoint so agents know what each backend supports
OpenClaw integration:
  rift serve as the interface between AI agent and infrastructure
  Agent discovers available operations per backend via /api/capabilities
  Enables granular troubleshooting and orchestration beyond CLI limitations
  Safety layer: proposed actions shown before execution, user approves
Enables rift overview --fabric to query all nodes over HTTP
API designed to accommodate future compute types (containers, pods) from the start


v2.1.0 — Docker Integration

rift docker command group
Container lifecycle: run, stop, start, rm, restart, logs, inspect
Image management: pull, list, build, tag, push, prune
Docker Compose integration: up, down, ps, logs for multi-container stacks
Network management: create, list, inspect, connect, disconnect
Volume management: create, list, inspect, rm
rift docker exec for command execution inside containers
Container appears in rift list and rift overview alongside VMs and cloud instances
Docker detection from .env (DOCKER_HOST) or local socket
API endpoints: /api/containers/* mirrors CLI container commands
Health checks and restart policies visible in overview
Port mapping and resource limits (CPU, memory) as first-class config
Foundation for container abstraction layer used by LXC and k8s later


v2.2.0 — LXC/LXD Integration

rift lxc command group
System container lifecycle: launch, start, stop, delete, restart, exec
Image management: list, copy, publish, import from simplestreams
Profile management: create, edit, apply, list
Storage pool management: create, list, attach to containers
Network management: create, attach, detach, list bridges
Snapshot support: create, restore, delete (parallels VM snapshot interface)
LXD REST API integration (lxd exposes native REST, no CLI shelling needed)
LXC containers appear in rift list and rift overview alongside VMs and Docker containers
LXD cluster awareness for multi-node setups
API endpoints: /api/lxc/* mirrors CLI commands
System containers as a bridge between Docker app containers and full VMs
File push/pull for direct file transfer into containers


v2.3.0 — Kubernetes Integration

rift k8s command group
Cluster lifecycle: create, delete, list, status (via k3s, kubeadm, or managed providers)
Workload management: deploy, scale, rollout, rollback, logs, exec
Service management: create, expose, list, describe
Namespace management: create, delete, switch context
Ingress and load balancer configuration
kubectl wrapping with rift auth context management
Helm chart support: install, upgrade, rollback, list releases
Direct k8s API integration via client-go or HTTP
Cluster appears in rift overview with node count, pod health, resource usage
Pod-level visibility: rift k8s pods shows individual pod state and resource consumption
CKS security contexts: network policies, pod security standards, RBAC templates, secrets management
Horizontal Pod Autoscaler configuration through rift
API endpoints: /api/k8s/* for full cluster and workload management
Prerequisite: Kubestronaut certification completed


v3.0.0 — Unified Compute Abstraction

Single interface for all compute types: VMs, Docker containers, LXC containers, k8s pods, cloud instances
Compute resource model:
  type: vm | container | system-container | pod | cloud-instance
  state: running | stopped | suspended | pending | terminated
  resources: cpu, memory, disk, network normalized across all types
rift list shows everything: VMware VMs, VBox VMs, Docker containers, LXC containers, k8s pods, AWS instances
rift overview aggregates all compute types per node with unified status
Unified lifecycle: start, stop, exec, logs work the same regardless of compute type
Unified API: /api/compute/* with type parameter replaces per-backend endpoints
Resource tagging and grouping across compute types
Dependency tracking: which containers depend on which VMs, which pods need which services
Compute type metadata: capabilities, limitations, cost profile per type
Foundation for strategies engine and resource morphing


v3.1.0 — Strategies Engine

rift strategy command group
Strategy definition in YAML:

  strategy:
    name: high-scalability
    compute-preference: [container, pod, vm, cloud-instance]
    scaling:
      min-replicas: 2
      max-replicas: 20
      trigger: cpu > 70%
    resources:
      containers:
        cpu: 500m-2000m
        memory: 256Mi-2Gi
      vms:
        cpu: 2-4
        memory: 4096-8192
      cloud:
        instance-type: t3.small-t3.large
        spot-allowed: true

Built-in strategy templates:
  high-scalability: container-first, horizontal scaling, smaller resource specs
  high-availability: redundancy across nodes and providers, failover chains, health checks
  cost-optimized: spot instances, containers over VMs, aggressive right-sizing
  performance: dedicated VMs, GPU passthrough, bare metal preference
  security-hardened: CKS-compliant, network policies, isolated namespaces
  custom: user-defined mix of compute preferences, constraints, and triggers
rift strategy apply <name> --file workload.yml deploys according to policy
rift strategy evaluate <name> shows projected resource allocation without executing
rift strategy status shows current compliance of running workloads against strategy
Strategy-aware deployment: rift decides compute type, resource sizing, and placement
Health monitoring: strategy enforcer watches resource state and triggers scaling
Auto-scaling across compute types: spin up containers first, fall back to VMs, burst to cloud
Cross-provider distribution: spread workloads based on strategy rules
API endpoints: /api/strategies/* for full strategy lifecycle


v3.2.0 — Resource Morphing

rift morph command group
VM to container:
  rift morph <vm> --to container
  Analyze VM workload (running services, listening ports, filesystem layout)
  Extract relevant filesystem layers
  Generate Dockerfile or container image from VM state
  Deploy container with equivalent networking, volumes, and environment
  Verify container serves the same function (health check)
  Optionally decommission source VM after verification
Container to VM:
  rift morph <container> --to vm
  Snapshot container filesystem
  Create VM with appropriate OS on target hypervisor
  Inject container filesystem and service configuration into VM
  Configure equivalent networking and storage
  Verify VM serves the same function
  Optionally decommission source container after verification
Instance spec migration:
  rift morph <resource> --spec <new-spec>
  Cloud instances: stop, change instance type, start (or live-resize where supported)
  VMs: modify CPU/RAM in VMX or hypervisor config, restart
  k8s deployments: update resource limits, rolling update with zero downtime
  LXC containers: live resource adjustment via LXD API
Strategy-driven morphing:
  rift morph <workload> --strategy <name>
  Strategy evaluates current resource type against policy
  Recommends upgrade or downgrade based on utilization and goals
  Example: high-scalability detects idle VM, recommends morph to container
CKS compliance during morph: security context preserved, network policies migrated
Pre-morph validation: compatibility checks, driver requirements, data persistence
Rollback support: source resource kept until morph verified, one-command rollback
API endpoints: /api/morph/* for programmatic morphing and OpenClaw-driven operations


v3.3.0 — Declarative Workloads from YAML

rift deploy --file workload.yml
Strategy-aware YAML manifest format:

  workload:
    name: web-platform
    strategy: high-scalability
    services:
      - name: frontend
        type: container
        image: nginx:latest
        replicas: 3
        network: public
      - name: api
        type: pod
        image: api-server:v2
        replicas: 2
        network: internal
      - name: database
        type: vm
        os: ubuntu-64
        cpu: 4
        ram: 8192
        disks:
          - size: 100GB
        network: internal
        snapshot-policy: daily
      - name: cache
        type: container
        image: redis:alpine
        network: internal
      - name: monitoring
        type: cloud-instance
        provider: aws
        instance-type: t3.small
        network: management

Creates resources on appropriate backends based on type and strategy
rift plan --file workload.yml shows projected deployment without executing
rift plan --file workload.yml --strategy cost-optimized shows alternative allocation
Parallel creation across all compute types
Cloud-init for VMs, environment variables for containers, helm values for pods
Post-deploy health validation across all resources
Exposed as POST /api/deploy with YAML or JSON payload
Update support: rift deploy --file workload.yml detects drift and reconciles


v4.0.0 — Fabric Mode (Multi-Node Orchestration)

rift overview --fabric queries all known nodes over Tailscale
Node registry in .env or nodes.yml:

  nodes:
    - name: omen
      address: 100.82.208.41:8080
      hypervisor: workstation
      runtimes: [docker]
    - name: pavilion
      address: 100.89.48.127:8080
      hypervisor: vbox
      runtimes: [docker, lxd]
    - name: ai
      address: 100.84.84.55:8080
      hypervisor: hyperv
      runtimes: [docker]
    - name: k8s-cluster
      address: 100.90.12.34:8080
      type: kubernetes
      context: production

rift fabric list shows all compute resources across all nodes
rift fabric start <resource> finds which node owns it and starts remotely
rift fabric migrate <resource> --from omen --to pavilion migrates between physical machines
Fabric-aware strategies: distribute workloads across nodes based on strategy rules
Cross-node resource morphing: morph a VM on omen into a container on pavilion
Cross-node network visualization: which resources on different nodes share subnets
Node health monitoring: detect node failures, trigger strategy failover
Fabric-level overview: total CPU, RAM, disk across all nodes and compute types


v4.1.0 — TUI Dashboard (Bubble Tea)

rift tui launches terminal UI
Dashboard view: all compute resources across all nodes in a navigable list
Resource type indicators: VM, container, pod, cloud instance with distinct styling
Keyboard shortcuts: Enter to start/stop, S for snapshot, E for exec, M for morph
Live status updates polling each node's API
Fabric view: visualize nodes, their compute resources, and cross-node relationships
Strategy view: show active strategies, compliance status, scaling events
Morph history: recent morph operations with before/after comparison
Progress bars for migration, morph, and export within the TUI
Lip Gloss styling: colors for state (green running, red stopped), compute type icons
Runs over SSH for remote access: ssh user@omen rift tui
Restricted SSH mode: ForceCommand rift tui for locked-down access


v4.2.0 — Hyper-V Full CLI Backend

Full Hypervisor interface implementation wrapping PowerShell cmdlets
Get-VM → list with state, CPU, RAM
Start-VM / Stop-VM → lifecycle
Checkpoint-VM → snapshot create
Restore-VMCheckpoint / Remove-VMCheckpoint → snapshot revert/delete
Export-VM / Import-VM → archive operations
New-VM → VM creation with CPU, RAM, disk, network configuration
Guest exec via PowerShell Direct or WinRM
Hardware config: Set-VMProcessor, Set-VMMemory, Add-VMNetworkAdapter, Add-VMHardDiskDrive
Test on AI (Tank) with real Hyper-V VMs
Integrates into unified compute abstraction alongside other hypervisors


v4.3.0 — Migration Expansion

VMware ↔ Hyper-V migration (VMDK ↔ VHDX via qemu-img)
VirtualBox ↔ Hyper-V migration (VDI ↔ VHDX via qemu-img)
All six hypervisor migration paths: vmware↔vbox, vmware↔hyperv, vbox↔hyperv
Driver cleanup during migration (remove VMware Tools, VBox Guest Additions, inject correct drivers)
initramfs regeneration for Linux guests after migration
NIC renaming and fstab UUID updates
Boot mode preservation (BIOS vs UEFI)
Migration API endpoints in rift serve for OpenClaw-driven migrations
Container migration: move Docker containers between nodes via image export/import
LXC migration: live migration between LXD cluster members


v5.0.0 — Proxmox Backend

Full Hypervisor interface via Proxmox REST API (/api2/json/)
Authentication with API tokens
VM lifecycle: start, stop, suspend, reset
Snapshot management
VM creation with cloud-init templates and storage pools
LXC container support alongside VMs (Proxmox-native LXC integrates with rift lxc)
Cluster awareness: multi-node Proxmox clusters
Live migration between Proxmox nodes
Storage pool management: local, ZFS, Ceph
Network management: bridges, VLANs, bonds
Install Proxmox on dedicated hardware (old laptop or Tank alternative)
Proxmox appears as a fabric node with both VM and LXC capabilities


v5.1.0 — Type 2 to Type 1 Migration

rift migrate <vm> --from vmware --to proxmox
Read VMX config, convert VMDK to QCOW2 via qemu-img
SCP disk to Proxmox over SSH/Tailscale
qm importdisk on Proxmox side via API
Create VM with matching specs via Proxmox API
VirtIO driver injection for Linux guests
Windows guests: inject VirtIO drivers before migration
rift migrate <vm> --from proxmox --to vmware for reverse path
All Type 2 ↔ Type 1 paths: vmware↔proxmox, vbox↔proxmox, hyperv↔proxmox


v5.2.0 — Cloud Migration

rift migrate <vm> --to aws

Convert disk to RAW format
Upload to S3
Call aws ec2 import-image
Poll until AMI is ready
Launch instance from imported AMI
Assign security group, elastic IP, key pair

rift migrate <instance> --from aws --to vmware

Create AMI from instance
Export to S3 as VMDK
Download and register in VMware

Cloud-init and driver normalization for cloud targets
Pre-migration checks: compatible disk format, boot mode, drivers
Strategy-aware cloud migration: strategy decides when to burst to cloud or pull back on-prem


v5.3.0 — API Backends (Performance)

Replace CLI shelling (vmrun, VBoxManage, PowerShell) with direct API calls
VMware Workstation: vmrest REST API
VirtualBox: COM/XPCOM API via Go bindings
Hyper-V: WMI direct queries instead of PowerShell subprocess
Reduced overhead, faster operations, more granular control
API backends expose features CLI backends can't: custom security groups, fine-grained network rules, programmatic firewall management
Automatic backend selection: use API backend when available, fall back to CLI
Performance matters now that rift serve handles concurrent requests from fabric and strategies


v6.0.0 — Ansible and Terraform Integration

Ansible inventory plugin: rift dynamic inventory source

Returns all compute resources across all backends as Ansible hosts
Groups by compute type, hypervisor, container runtime, node, strategy

Ansible module: rift_compute for lifecycle management in playbooks
Terraform provider: terraform-provider-rift

rift_vm resource for VM creation
rift_container resource for container creation
rift_strategy resource for strategy management
rift_snapshot resource
rift_network data source
State management via Terraform state file

Post-creation Ansible playbooks: auto-configure resources after rift deploy


v6.1.0 — Network Lab Automation

Topology-aware deployments from YAML

  topology:
    routers:
      - name: R1
        image: vyos
        interfaces:
          - network: transit
          - network: lan-a
    switches:
      - name: SW1
        interfaces:
          - network: lan-a
          - network: lan-b
    clients:
      - name: PC1
        network: lan-a
      - name: PC2
        network: lan-b

Virtual network creation across hypervisors and container runtimes
Cross-compute networking: bridge VMware VM to Docker container on same subnet
Predefined templates: multi-router topology, K8s cluster, AD domain, DMZ
Network validation: ping tests, route checks, port scans after deployment


v7.0.0 — Additional Cloud Providers

Azure: rift azure create, rift azure list, managed disks, NSG rules
GCP: rift gcp create, rift gcp list, firewall rules, persistent disks
Oracle Cloud: KVM-based, similar to Proxmox integration
Cloud migration paths: AWS↔Azure, AWS↔GCP, any cloud↔any hypervisor
Multi-cloud overview in rift overview
Cost estimation: rift plan --file workload.yml --estimate shows projected cloud costs
Multi-cloud strategies: distribute workloads across cloud providers based on cost and region


v7.1.0 — Additional Hypervisors

KVM/libvirt: virsh CLI or libvirt Go bindings
XCP-ng: XenAPI
oVirt/RHV: REST API
Each follows the same pattern: implement Hypervisor interface, register in factory, appear in unified compute


v8.0.0 — AI-Driven Orchestration

rift ai command group
Local LLM (Ollama on Tank) interprets natural language:

"Start my K3S cluster" → rift start --folder K3S
"Scale the frontend to 10 replicas" → rift k8s scale frontend --replicas 10
"Migrate the Ubuntu server to a container" → rift morph H26-U-S --to container
"Apply high-availability to the web platform" → rift strategy apply high-availability --workload web-platform
"Create 3 web servers on AWS" → rift aws create --ami <id> --name web --count 3

Model reads rift --help and command documentation as context
rift serve JSON API (built in v2.0.0) is the interface between model and tool
Builds on OpenClaw integration from v2.0.0 with expanded natural language layer
Strategy recommendations: AI suggests optimal strategy based on workload analysis
Morph recommendations: AI detects underutilized VMs and suggests container morphing
Safety: rift plan shows proposed actions before execution, user approves


Ongoing — Quality and Polish

Unit test coverage for all non-hypervisor logic
Integration test suites per hypervisor and container runtime (run on dedicated machines)
README maintenance with each release
GitHub releases with compiled binaries (Windows + Linux)
Error message improvements based on user feedback
Performance profiling and optimization
Documentation site (GitHub Pages or docs folder)
Contributing guide for open source contributions
