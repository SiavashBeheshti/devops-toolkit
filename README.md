<h1 align="center">DevOps Toolkit</h1>

<p align="center">
  <strong>A powerful, beautiful CLI toolkit for modern DevOps operations</strong>
</p>

<p align="center">
  <a href="https://github.com/SiavashBeheshti/devops-toolkit/releases"><img src="https://img.shields.io/github/v/release/SiavashBeheshti/devops-toolkit?style=flat-square&color=7C3AED" alt="Release"></a>
  <a href="https://github.com/SiavashBeheshti/devops-toolkit/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-green.svg?style=flat-square" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/SiavashBeheshti/devops-toolkit"><img src="https://goreportcard.com/badge/github.com/SiavashBeheshti/devops-toolkit?style=flat-square" alt="Go Report Card"></a>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <a href="https://github.com/SiavashBeheshti/devops-toolkit/actions"><img src="https://img.shields.io/github/actions/workflow/status/SiavashBeheshti/devops-toolkit/ci.yml?style=flat-square" alt="Build Status"></a>
</p>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-installation">Installation</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-documentation">Documentation</a> •
  <a href="#-contributing">Contributing</a>
</p>

---

## 🎯 Why DevOps Toolkit?

Native DevOps tools often provide minimal, hard-to-read output. **DevOps Toolkit** transforms your terminal experience with:

- **🎨 Beautiful Output** — Color-coded statuses, progress bars, and styled tables
- **📊 Enhanced Visibility** — See more information at a glance than native tools provide
- **🔄 Unified Interface** — One CLI for Kubernetes, Docker, GitLab, and compliance checks
- **⚡ Productivity Boost** — Common operations simplified into single commands

<p align="center">
  <img src="https://raw.githubusercontent.com/SiavashBeheshti/devops-toolkit/main/.github/assets/demo.gif" alt="DevOps Toolkit Demo" width="800"/>
</p>

---

## ✨ Features

### 🚀 Kubernetes Operations

| Command | Description |
|---------|-------------|
| `k8s health` | Comprehensive cluster health dashboard |
| `k8s pods` | Enhanced pod listing with status colors & restart counts |
| `k8s nodes` | Node status with resource utilization bars |
| `k8s resources` | CPU/Memory breakdown by namespace |
| `k8s cleanup` | Remove failed pods, completed jobs, orphaned resources |
| `k8s events` | Filtered event viewing with highlighting |

<details>
<summary>📸 Screenshot: Kubernetes Health Check</summary>

```
╔══════════════════════════════════════════════════════════════════╗
║  Cluster Health Summary                                          ║
╚══════════════════════════════════════════════════════════════════╝

  Component      Status          Details
  ─────────────────────────────────────────────────────────────────
  Nodes          ✓ Healthy       5/5 Ready
  Pods           ✓ Healthy       Running: 45, Pending: 0, Failed: 0
  PVCs           ✓ OK            Bound: 12/12
  Deployments    ⚠ Warning       Ready: 14/15, Unavailable: 1
  Services       ✓ OK            ClusterIP: 12, LoadBalancer: 3

╔══════════════════════════════════════════════════════════════════╗
║  Resource Utilization                                            ║
╚══════════════════════════════════════════════════════════════════╝

  Resource    Used        Capacity    Utilization
  ─────────────────────────────────────────────────────────────────
  CPU         2400m       8000m       ████████░░░░░░░░░░░░  30%
  Memory      12.4Gi      32.0Gi      ████████████░░░░░░░░  39%
```

</details>

### 🐳 Docker Operations

| Command | Description |
|---------|-------------|
| `docker containers` | Enhanced container listing with health status |
| `docker images` | Image analysis with size breakdown |
| `docker stats` | Real-time resource usage with visual bars |
| `docker clean` | Smart cleanup of unused resources |
| `docker inspect` | Beautiful, readable container details |
| `docker logs` | Syntax-highlighted log viewing |

<details>
<summary>📸 Screenshot: Docker Stats</summary>

```
╔══════════════════════════════════════════════════════════════════╗
║  Container Statistics                                            ║
╚══════════════════════════════════════════════════════════════════╝

  Container       CPU %                 Mem %                 Net I/O
  ─────────────────────────────────────────────────────────────────────
  nginx           ████░░░░░░░░░░░  23%  ██░░░░░░░░░░░░░  12%  1.2MB / 890KB
  postgres        █░░░░░░░░░░░░░░   8%  ████████████░░░  78%  45MB / 12MB
  redis           ░░░░░░░░░░░░░░░   2%  █░░░░░░░░░░░░░░   5%  230KB / 180KB
  api-server      ██████░░░░░░░░░  45%  ██████░░░░░░░░░  42%  890MB / 1.2GB

▸ Resource Summary
  Total CPU: 78%
  Total Memory: 8.2 GB / 16.0 GB (51%)

▸ Alerts
  ⚠ postgres: High memory usage (78%)
```

</details>

### 🦊 GitLab CI/CD

| Command | Description |
|---------|-------------|
| `gitlab pipelines` | List pipelines with status indicators |
| `gitlab jobs` | View jobs grouped by stage |
| `gitlab trigger` | Trigger new pipelines with variables |
| `gitlab artifacts` | Manage pipeline artifacts |
| `gitlab status` | Project CI/CD dashboard |

<details>
<summary>📸 Screenshot: GitLab Pipelines</summary>

```
╔══════════════════════════════════════════════════════════════════╗
║  CI/CD Pipelines                                                 ║
╚══════════════════════════════════════════════════════════════════╝

  ID        Status           Ref              Commit      Duration
  ─────────────────────────────────────────────────────────────────
  #1234     ✓ success        main             a1b2c3d4    5m 23s
  #1233     ✓ success        main             e5f6g7h8    4m 12s
  #1232     ✗ failed         feature/auth     i9j0k1l2    2m 45s
  #1231     ● running        develop          m3n4o5p6    3m 10s
  #1230     ○ pending        hotfix/bug       q7r8s9t0    -

▸ Pipeline Summary
  ✓ Success: 2
  ✗ Failed: 1
  ● Running: 1
  ○ Pending: 1
```

</details>

### 🔒 Compliance & Security

| Command | Description |
|---------|-------------|
| `compliance check k8s` | Kubernetes security best practices |
| `compliance check docker` | Container security analysis |
| `compliance check files` | Validate manifests & Dockerfiles |
| `compliance report [target]` | Generate HTML/JSON/JUnit reports (k8s, docker, files, all) |
| `compliance policies` | List all available policies |

<details>
<summary>📸 Screenshot: Compliance Check</summary>

```
╔══════════════════════════════════════════════════════════════════╗
║  Compliance Check                                                ║
╚══════════════════════════════════════════════════════════════════╝

▸ Kubernetes Security

  Status  Severity  Rule           Resource              Message
  ─────────────────────────────────────────────────────────────────
  ✗       CRIT      K8S-SEC-001    default/nginx         Container running privileged
  ✗       HIGH      K8S-SEC-002    default/api           Running as root user
  ✓       MED       K8S-SEC-003    default/worker        Read-only filesystem enabled
  ⚠       MED       K8S-RES-001    default/nginx         No CPU limits defined

▸ Summary
  Total Checks: 24
  ✓ Passed: 18
  ✗ Failed: 4
  ⚠ Warnings: 2

  Compliance Score: ████████████████░░░░ 75%
```

</details>

---

## 📦 Installation

### Go Install

```bash
go install github.com/SiavashBeheshti/devops-toolkit@latest
```

### Download Binary

Download the latest binary from the [Releases](https://github.com/SiavashBeheshti/devops-toolkit/releases) page.

```bash
# Set the version you want to install
VERSION="1.0.0"

# Linux (amd64)
curl -LO "https://github.com/SiavashBeheshti/devops-toolkit/releases/download/v${VERSION}/devops-toolkit_${VERSION}_linux_amd64"
chmod +x devops-toolkit_${VERSION}_linux_amd64
sudo mv devops-toolkit_${VERSION}_linux_amd64 /usr/local/bin/devops-toolkit

# Linux (arm64)
curl -LO "https://github.com/SiavashBeheshti/devops-toolkit/releases/download/v${VERSION}/devops-toolkit_${VERSION}_linux_arm64"
chmod +x devops-toolkit_${VERSION}_linux_arm64
sudo mv devops-toolkit_${VERSION}_linux_arm64 /usr/local/bin/devops-toolkit

# macOS (Intel)
curl -LO "https://github.com/SiavashBeheshti/devops-toolkit/releases/download/v${VERSION}/devops-toolkit_${VERSION}_darwin_amd64"
chmod +x devops-toolkit_${VERSION}_darwin_amd64
sudo mv devops-toolkit_${VERSION}_darwin_amd64 /usr/local/bin/devops-toolkit

# macOS (Apple Silicon)
curl -LO "https://github.com/SiavashBeheshti/devops-toolkit/releases/download/v${VERSION}/devops-toolkit_${VERSION}_darwin_arm64"
chmod +x devops-toolkit_${VERSION}_darwin_arm64
sudo mv devops-toolkit_${VERSION}_darwin_arm64 /usr/local/bin/devops-toolkit

# Windows (amd64) - using PowerShell
# Invoke-WebRequest -Uri "https://github.com/SiavashBeheshti/devops-toolkit/releases/download/v${VERSION}/devops-toolkit_${VERSION}_windows_amd64.exe" -OutFile "devops-toolkit.exe"
```

### Build from Source

```bash
git clone https://github.com/SiavashBeheshti/devops-toolkit.git
cd devops-toolkit
make build
sudo make install-local
```

### Docker

```bash
docker pull ghcr.io/SiavashBeheshti/devops-toolkit:latest

# Run with kubectl config
docker run -it --rm \
  -v ~/.kube:/root/.kube \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/SiavashBeheshti/devops-toolkit k8s health
```

### Shell Alias (Optional)

For convenience, you can set up a shorter alias:

```bash
# Bash (~/.bashrc)
echo 'alias dtk="devops-toolkit"' >> ~/.bashrc
source ~/.bashrc

# Zsh (~/.zshrc)
echo 'alias dtk="devops-toolkit"' >> ~/.zshrc
source ~/.zshrc

# Fish (~/.config/fish/config.fish)
echo 'alias dtk="devops-toolkit"' >> ~/.config/fish/config.fish
source ~/.config/fish/config.fish
```

Then use `dtk` instead of `devops-toolkit`:

```bash
dtk k8s health
dtk docker stats
dtk gitlab pipelines
```

### Shell Completion

DevOps Toolkit supports shell auto-completion for commands, flags, and resource names (pods, containers, namespaces, etc.).

#### Bash

```bash
# Linux
devops-toolkit completion bash > /etc/bash_completion.d/devops-toolkit

# macOS (with Homebrew)
devops-toolkit completion bash > $(brew --prefix)/etc/bash_completion.d/devops-toolkit

# Or load for current session only
source <(devops-toolkit completion bash)
```

#### Zsh

```bash
# If shell completion is not already enabled:
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Generate completion script
devops-toolkit completion zsh > "${fpath[1]}/_devops-toolkit"

# For Oh My Zsh users:
devops-toolkit completion zsh > ~/.oh-my-zsh/completions/_devops-toolkit

# Or load for current session only
source <(devops-toolkit completion zsh)
```

#### Fish

```bash
devops-toolkit completion fish > ~/.config/fish/completions/devops-toolkit.fish

# Or load for current session only
devops-toolkit completion fish | source
```

#### PowerShell

```powershell
# Load for current session
devops-toolkit completion powershell | Out-String | Invoke-Expression

# Add to profile for persistent loading
devops-toolkit completion powershell >> $PROFILE
```

#### What Gets Completed

- **Commands & Subcommands**: `devops-toolkit k8s <TAB>` shows pods, nodes, health, etc.
- **Flags**: `devops-toolkit k8s pods --<TAB>` shows available flags
- **Kubernetes Resources**: Pod names, namespace names, container names, context names
- **Docker Resources**: Container names/IDs, image names, volume names, network names
- **Flag Values**: `--namespace <TAB>` lists namespaces, `--format <TAB>` shows format options

---

## 🚀 Quick Start

### Prerequisites

- **Kubernetes**: `kubectl` configured with cluster access
- **Docker**: Docker daemon running
- **GitLab**: Access token with API permissions

### Basic Usage

```bash
# Check Kubernetes cluster health
devops-toolkit k8s health

# List all pods with enhanced output
devops-toolkit k8s pods -A

# Show Docker container statistics
devops-toolkit docker stats

# List GitLab pipelines
devops-toolkit gitlab pipelines -p mygroup/myproject

# Run compliance checks
devops-toolkit compliance check k8s
```

---

## 📖 Documentation

### Kubernetes Commands

```bash
# ═══════════════════════════════════════════════════════════════════
# CLUSTER HEALTH
# ═══════════════════════════════════════════════════════════════════

# Full cluster health overview
devops-toolkit k8s health

# Health check for specific namespace
devops-toolkit k8s health -n production

# ═══════════════════════════════════════════════════════════════════
# POD MANAGEMENT
# ═══════════════════════════════════════════════════════════════════

# List pods in current namespace
devops-toolkit k8s pods

# List pods in all namespaces
devops-toolkit k8s pods -A

# Show only problematic pods
devops-toolkit k8s pods --problems

# Sort by restarts (descending)
devops-toolkit k8s pods -s restarts

# Wide output with node and IP
devops-toolkit k8s pods --wide

# Filter by label
devops-toolkit k8s pods -l app=nginx

# ═══════════════════════════════════════════════════════════════════
# NODE ANALYSIS
# ═══════════════════════════════════════════════════════════════════

# List nodes with basic info
devops-toolkit k8s nodes

# Show resource utilization
devops-toolkit k8s nodes --resources

# Wide output with OS and kernel info
devops-toolkit k8s nodes --wide

# ═══════════════════════════════════════════════════════════════════
# RESOURCE USAGE
# ═══════════════════════════════════════════════════════════════════

# Cluster-wide resource summary
devops-toolkit k8s resources

# Show top resource-consuming pods
devops-toolkit k8s resources --top-pods

# Limit to top 5 pods
devops-toolkit k8s resources --top-pods --limit 5

# ═══════════════════════════════════════════════════════════════════
# CLEANUP
# ═══════════════════════════════════════════════════════════════════

# Dry run - see what would be deleted
devops-toolkit k8s cleanup

# Actually perform cleanup
devops-toolkit k8s cleanup --dry-run=false

# Cleanup specific resource types
devops-toolkit k8s cleanup --completed-pods --failed-pods --dry-run=false

# Include orphaned ReplicaSets
devops-toolkit k8s cleanup --orphan-rs --dry-run=false

# ═══════════════════════════════════════════════════════════════════
# EVENTS
# ═══════════════════════════════════════════════════════════════════

# Show recent events
devops-toolkit k8s events

# Show only warnings
devops-toolkit k8s events --warnings-only

# Filter by reason
devops-toolkit k8s events --reason BackOff

# Limit number of events
devops-toolkit k8s events --limit 20
```

### Docker Commands

```bash
# ═══════════════════════════════════════════════════════════════════
# CONTAINERS
# ═══════════════════════════════════════════════════════════════════

# List running containers
devops-toolkit docker containers

# List all containers (including stopped)
devops-toolkit docker containers -a

# Wide output with command and created time
devops-toolkit docker containers --wide

# Show container sizes
devops-toolkit docker containers --size

# ═══════════════════════════════════════════════════════════════════
# IMAGES
# ═══════════════════════════════════════════════════════════════════

# List images sorted by size
devops-toolkit docker images

# Show only dangling images
devops-toolkit docker images --dangling

# Sort by name or created time
devops-toolkit docker images -s name
devops-toolkit docker images -s created

# Show image digests
devops-toolkit docker images --digest

# ═══════════════════════════════════════════════════════════════════
# STATISTICS
# ═══════════════════════════════════════════════════════════════════

# Show real-time container stats
devops-toolkit docker stats

# ═══════════════════════════════════════════════════════════════════
# CLEANUP
# ═══════════════════════════════════════════════════════════════════

# Dry run - see what would be cleaned
devops-toolkit docker clean

# Actually perform cleanup
devops-toolkit docker clean --dry-run=false

# Include unused volumes (dangerous!)
devops-toolkit docker clean --volumes --dry-run=false

# Remove all unused images (not just dangling)
devops-toolkit docker clean --all-images --dry-run=false

# ═══════════════════════════════════════════════════════════════════
# INSPECT & LOGS
# ═══════════════════════════════════════════════════════════════════

# Inspect container with beautiful output
devops-toolkit docker inspect mycontainer

# Show all details (env, mounts, network)
devops-toolkit docker inspect mycontainer --all

# View logs with highlighting
devops-toolkit docker logs mycontainer

# Tail last 50 lines
devops-toolkit docker logs mycontainer -n 50

# Follow logs
devops-toolkit docker logs mycontainer -f

# Show timestamps
devops-toolkit docker logs mycontainer --timestamps
```

### GitLab Commands

```bash
# ═══════════════════════════════════════════════════════════════════
# SETUP
# ═══════════════════════════════════════════════════════════════════

# Set credentials via environment
export GITLAB_TOKEN=your-personal-access-token
export GITLAB_PROJECT=mygroup/myproject
export GITLAB_URL=https://gitlab.com  # Optional, defaults to gitlab.com

# Or use flags
devops-toolkit gitlab pipelines --token $TOKEN --project mygroup/myproject

# ═══════════════════════════════════════════════════════════════════
# PIPELINES
# ═══════════════════════════════════════════════════════════════════

# List recent pipelines
devops-toolkit gitlab pipelines

# Filter by status
devops-toolkit gitlab pipelines -s running
devops-toolkit gitlab pipelines -s failed

# Filter by branch
devops-toolkit gitlab pipelines -r main

# Show more pipelines
devops-toolkit gitlab pipelines -n 50

# ═══════════════════════════════════════════════════════════════════
# JOBS
# ═══════════════════════════════════════════════════════════════════

# List jobs for a pipeline
devops-toolkit gitlab jobs -i 12345

# Show only failed jobs
devops-toolkit gitlab jobs -i 12345 --failed

# Filter by stage
devops-toolkit gitlab jobs -i 12345 --stage test

# ═══════════════════════════════════════════════════════════════════
# TRIGGER
# ═══════════════════════════════════════════════════════════════════

# Trigger pipeline on branch
devops-toolkit gitlab trigger -r main

# Trigger with variables
devops-toolkit gitlab trigger -r main -v ENV=production -v DEBUG=true

# Trigger and wait for completion
devops-toolkit gitlab trigger -r main --wait

# ═══════════════════════════════════════════════════════════════════
# STATUS & ARTIFACTS
# ═══════════════════════════════════════════════════════════════════

# Project CI/CD overview
devops-toolkit gitlab status

# List pipeline artifacts
devops-toolkit gitlab artifacts -i 12345
```

### Compliance Commands

```bash
# ═══════════════════════════════════════════════════════════════════
# CHECKS
# ═══════════════════════════════════════════════════════════════════

# Check Kubernetes resources
devops-toolkit compliance check k8s

# Check specific namespace
devops-toolkit compliance check k8s -n production

# Check Docker containers and images
devops-toolkit compliance check docker

# Check specific image
devops-toolkit compliance check docker --image nginx:latest

# Check configuration files
devops-toolkit compliance check files --path ./manifests

# Run all checks
devops-toolkit compliance check all

# Skip specific rules
devops-toolkit compliance check k8s --skip K8S-SEC-001,K8S-SEC-002

# Only run specific rules
devops-toolkit compliance check k8s --only K8S-SEC-001

# Set minimum severity
devops-toolkit compliance check k8s --severity high

# Fail on warnings (for CI)
devops-toolkit compliance check k8s --fail-on-warn

# ═══════════════════════════════════════════════════════════════════
# REPORTS
# ═══════════════════════════════════════════════════════════════════

# Generate HTML report for all checks
devops-toolkit compliance report -f html -o report.html

# Generate report for Kubernetes checks only
devops-toolkit compliance report k8s -f html -o k8s-report.html

# Generate report for Docker checks only
devops-toolkit compliance report docker -f json -o docker-report.json

# Generate report for file checks only
devops-toolkit compliance report files -f html -o files-report.html

# Generate JUnit XML (for CI integration)
devops-toolkit compliance report -f junit -o results.xml

# Generate report for specific namespace
devops-toolkit compliance report k8s -n production -f html -o prod-report.html

# Exclude passed checks from report
devops-toolkit compliance report --include-passed=false

# ═══════════════════════════════════════════════════════════════════
# POLICIES
# ═══════════════════════════════════════════════════════════════════

# List all available policies
devops-toolkit compliance policies

# Filter by category
devops-toolkit compliance policies --category "Kubernetes Security"

# Filter by severity
devops-toolkit compliance policies --severity critical
```

---

## ⚙️ Configuration

### Configuration File

Create `~/.devops-toolkit.yaml`:

```yaml
# GitLab Configuration
gitlab:
  url: https://gitlab.com
  token: glpat-xxxxxxxxxxxxxxxxxxxx
  project: mygroup/myproject

# Default Settings
defaults:
  output: table      # table, json, yaml
  verbose: false
  
# Kubernetes Settings
kubernetes:
  context: ""        # Use specific context
  namespace: ""      # Default namespace

# Compliance Settings  
compliance:
  policy_dir: ~/.devops-toolkit/policies
  skip_rules: []
  severity: low      # Minimum severity to report
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GITLAB_TOKEN` | GitLab personal access token | - |
| `GITLAB_URL` | GitLab instance URL | `https://gitlab.com` |
| `GITLAB_PROJECT` | Default project ID or path | - |
| `KUBECONFIG` | Kubernetes config file path | `~/.kube/config` |
| `DEVOPS_TOOLKIT_CONFIG` | Config file path | `~/.devops-toolkit.yaml` |

---

## 🏗️ Architecture

```
devops-toolkit/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command & global flags
│   ├── completion.go      # Shell completion command
│   ├── k8s/               # Kubernetes subcommands
│   ├── docker/            # Docker subcommands
│   ├── gitlab/            # GitLab subcommands
│   └── compliance/        # Compliance subcommands
│
├── pkg/                    # Reusable packages
│   ├── output/            # Terminal output formatting
│   │   ├── theme.go       # Colors & styles (Lipgloss)
│   │   ├── table.go       # Table rendering
│   │   └── printer.go     # Print utilities & spinners
│   ├── completion/        # Shell completion helpers
│   │   ├── k8s.go         # K8s resource completions
│   │   ├── docker.go      # Docker resource completions
│   │   └── common.go      # Common completions
│   ├── k8s/               # Kubernetes client wrapper
│   ├── docker/            # Docker client wrapper
│   ├── gitlabclient/      # GitLab API client
│   └── compliance/        # Compliance engine
│       ├── k8s_checker.go
│       ├── docker_checker.go
│       └── file_checker.go
│
├── main.go                # Entry point
├── go.mod                 # Go modules
├── Makefile              # Build automation
└── .goreleaser.yaml      # Release configuration
```

### Tech Stack

| Component | Technology |
|-----------|------------|
| CLI Framework | [Cobra](https://github.com/spf13/cobra) |
| Configuration | [Viper](https://github.com/spf13/viper) |
| Terminal Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| Tables | [Tablewriter](https://github.com/olekukonko/tablewriter) |
| Spinners | [Spinner](https://github.com/briandowns/spinner) |
| Kubernetes | [client-go](https://github.com/kubernetes/client-go) |
| Docker | [Docker SDK](https://github.com/docker/docker) |
| GitLab | [go-gitlab](https://github.com/xanzy/go-gitlab) |

---

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Docker (for testing Docker commands)
- kubectl configured (for testing K8s commands)

### Setup

```bash
# Clone repository
git clone https://github.com/SiavashBeheshti/devops-toolkit.git
cd devops-toolkit

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Lint
make lint
```

### Make Commands

```bash
make help          # Show available commands
make build         # Build for current platform
make build-all     # Build for all platforms
make install       # Install to GOPATH/bin
make test          # Run tests
make test-coverage # Run tests with coverage
make lint          # Run linter
make fmt           # Format code
make clean         # Clean build artifacts
```

---

## 🗺️ Roadmap

- [x] Kubernetes operations
- [x] Docker operations
- [x] GitLab CI/CD integration
- [x] Compliance checking
- [x] Shell auto-completion (bash, zsh, fish, powershell)
- [ ] GitHub Actions integration
- [ ] AWS/GCP/Azure cloud operations
- [ ] Terraform state viewer
- [ ] Helm chart analysis
- [ ] Interactive TUI mode
- [ ] Plugin system
- [ ] Prometheus metrics querying
- [ ] Log aggregation (Loki/ELK)

See the [open issues](https://github.com/SiavashBeheshti/devops-toolkit/issues) for a full list of proposed features.

---

## 🤝 Contributing

Contributions make the open-source community amazing! Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

---

## 📄 License

Distributed under the MIT License. See [LICENSE](LICENSE) for more information.

---

## 🙏 Acknowledgements

- [Cobra](https://github.com/spf13/cobra) - Powerful CLI library
- [Charm](https://github.com/charmbracelet) - Beautiful terminal UI libraries
- [Kubernetes](https://kubernetes.io/) - Container orchestration
- [Docker](https://www.docker.com/) - Containerization platform
- [GitLab](https://gitlab.com/) - DevOps platform

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/SiavashBeheshti">@SiavashBeheshti</a>
</p>

<p align="center">
  <a href="https://github.com/SiavashBeheshti/devops-toolkit/stargazers">⭐ Star this repo</a> if you find it useful!
</p>
