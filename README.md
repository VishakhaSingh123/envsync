# Environment Synchronization Tool

Student Name: Vishakha Singh
Registration No: 23FE10CSE00134
Course: CSE3253 DevOps [PE6]
Semester: VI (2025–2026)
Project Type: DevOps Fundamentals & Ecosystem
Difficulty: Intermediate

---

## 📌 Project Overview

### Problem Statement
In modern software development, teams maintain multiple environments — Dev, Staging, and Production. Over time, small manual changes cause **configuration drift**, where environments behave differently from each other. This causes the classic *"it works on my machine"* problem, leading to failed deployments, production bugs, and wasted debugging time.

The **Environment Synchronization Tool (EnvSync)** solves this by continuously auditing, comparing, and synchronizing environment variables, configuration files, and runtime versions across all environments — ensuring parity throughout the DevOps pipeline.

### Objectives
- [x] Detect configuration drift between Dev, Staging, and Production environments
- [x] Securely sync environment variables without exposing secrets in plain text
- [x] Validate runtime versions (Node, Python, Go, etc.) across environments
- [x] Auto-snapshot environments before any sync to enable safe rollbacks
- [x] Integrate with CI/CD pipelines to block deployments when drift exceeds threshold
- [x] Provide a web dashboard for visual drift monitoring and one-click sync

### Key Features
- **Drift Detection** — Compare any two environments and get a detailed report showing MISSING, MISMATCH, and EXTRA keys
- **Secret Management** — AES-256-GCM encryption ensures secrets are never stored in plain text on disk
- **Runtime Validation** — Verify Node.js, Python, Go versions match the required spec
- **Auto Snapshot & Rollback** — Automatically snapshots before every sync; rollback with one command
- **CI/CD Drift Gate** — Blocks deployment if drift count exceeds a configurable threshold
- **Strict Mode for Production** — Production syncs require explicit approval (PR-style gate)
- **Web Dashboard** — Visual heatmap, side-by-side diff, snapshot timeline, and sync simulator

---

## 🛠️ Technology Stack

### Core Technologies
- **Programming Language:** Go 1.21+
- **Framework:** Cobra (CLI framework)
- **Database:** None (file-based state + optional Redis for caching)

### DevOps Tools
- **Version Control:** Git
- **CI/CD:** GitHub Actions
- **Containerization:** Docker
- **Orchestration:** Kubernetes (manifests included)
- **Secret Management:** HashiCorp Vault (local dev via Docker)
- **Monitoring:** Nagios / custom health checks
- **Configuration Management:** envsync.yaml state file

---

## 🚀 Getting Started

### Prerequisites
- [ ] Go 1.21+ — https://go.dev/dl/
- [ ] Git 2.30+
- [ ] Docker Desktop v20.10+ (optional, for Vault + dashboard)
- [ ] VSCode with Go extension (golang.go)
- [ ] OpenSSL (for generating encryption key)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/vishakhasingh/devopsprojectenvironmentsynchronizationtool.git
   cd devopsprojectenvironmentsynchronizationtool
   ```

2. Install Go dependencies:
   ```bash
   go mod tidy
   ```

3. Build the binary:
   ```bash
   # Linux / macOS
   go build -o envsync .

   # Windows
   go build -o envsync.exe .
   ```

4. Set encryption key:
   ```bash
   # Linux / macOS
   export ENVSYNC_KEY=$(openssl rand -base64 32)

   # Windows PowerShell
   $env:ENVSYNC_KEY = [Convert]::ToBase64String((1..32 | ForEach-Object { [byte](Get-Random -Max 256) }))
   ```

5. Run it:
   ```bash
   ./envsync --help
   ```

### Build and run using Docker:
```bash
docker-compose up --build
```

### Access the application:
- **Web Dashboard:** http://localhost:8080
- **HashiCorp Vault UI:** http://localhost:8200 (token: root)

### Alternative Installation (Without Docker)
```bash
go mod tidy
go build -o envsync .
./envsync init
./envsync audit --env dev
```

---

## 📁 Project Structure

```
devopsprojectenvironmentsynchronizationtool/
│
├── README.md                           Main project documentation
├── .gitignore                          Git ignore file
├── LICENSE                             MIT License
│
├── src/                                Source code
│   ├── main/                           Main implementation
│   │   ├── main.go                     Entry point
│   │   └── config/                     Configuration files
│   │       └── envsync.yaml            Master config
│   ├── cmd/                            CLI commands
│   │   ├── root.go
│   │   ├── diff.go
│   │   ├── sync.go
│   │   ├── audit.go
│   │   ├── snapshot.go
│   │   ├── rollback.go
│   │   └── validate.go
│   ├── internal/                       Internal packages
│   │   ├── parser/parser.go
│   │   ├── comparator/comparator.go
│   │   ├── crypto/crypto.go
│   │   ├── sync/sync.go
│   │   └── snapshot/snapshot.go
│   └── scripts/
│       ├── install.sh
│       └── vault-init.sh
│
├── docs/                               Documentation
│   ├── projectplan.md
│   ├── designdocument.md
│   ├── userguide.md
│   ├── apidocumentation.md
│   └── screenshots/
│
├── infrastructure/                     Infrastructure as Code
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── kubernetes/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── configmap.yaml
│   ├── puppet/
│   └── terraform/
│
├── pipelines/                          CI/CD Pipeline definitions
│   ├── Jenkinsfile
│   ├── .github/workflows/
│   │   ├── envsync.yml
│   │   ├── drift-check.yml
│   │   └── pre-deploy.yml
│   └── gitlab-ci.yml
│
├── tests/
│   ├── unit/
│   ├── integration/
│   └── testdata/
│
├── monitoring/
│   ├── nagios/
│   ├── alerts/
│   └── dashboards/
│
├── web/dashboard/
│   └── index.html
│
├── presentations/
│   ├── project-presentation.pptx
│   └── demo-script.md
│
└── deliverables/
    ├── demo-video.mp4
    ├── final-report.pdf
    └── assessment/
```

---

## ⚙️ Configuration

### Environment Variables
Create a `.env` file in the root directory:

```env
APP_ENV=development
ENVSYNC_KEY=your_generated_key_here
VAULT_ADDR=http://localhost:8200
VAULT_TOKEN=root
REDIS_PASSWORD=envsync-dev
```

### Key Configuration Files
1. `envsync.yaml` — Master config pointing to all environments
2. `docker-compose.yml` — Multi-container setup
3. `infrastructure/kubernetes/` — K8s deployment files
4. `.env.example` — Source of truth (all required keys listed here)

---

## 🔄 CI/CD Pipeline

### Pipeline Stages
1. **Code Quality Check** — Go linting, static analysis
2. **Build** — Compile Go binary, build Docker image
3. **Test** — Run unit and integration tests
4. **Security Scan** — Trivy vulnerability scan
5. **Deploy to Staging** — Auto drift-check + sync + snapshot
6. **Deploy to Production** — Manual approval required + zero-tolerance audit

### Pipeline Status
![Pipeline Status](https://img.shields.io/badge/pipeline-passing-brightgreen)

### GitHub Actions Secrets Required

| Secret | Description |
|--------|-------------|
| `ENV_DEV` | Contents of `.env.dev` |
| `ENV_STAGING` | Contents of `.env.staging` |
| `ENV_PRODUCTION` | Contents of `.env.production` |
| `ENVSYNC_KEY` | AES-256 encryption key |

---

## 🧪 Testing

### Test Types
- **Unit Tests:** `go test ./...`
- **Integration Tests:** `go test ./... -tags=integration`
- **Drift Tests:** `./envsync diff dev staging`

### Running Tests
```bash
go test ./... -v -race
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage

| Package | Coverage |
|---------|---------|
| internal/comparator | 95% |
| internal/crypto | 92% |
| internal/parser | 88% |
| internal/sync | 85% |
| internal/snapshot | 90% |

---

## 📊 Monitoring & Logging

### Monitoring Setup
- **Nagios:** System and environment health monitoring
- **Custom Metrics:** Drift count, sync frequency, snapshot count
- **Alerts:** Notifications when drift exceeds threshold

### Logging
- Structured colored terminal logging (INFO, WARN, ERROR)
- Log retention: 30 days

---

## 🐳 Docker & Kubernetes

### Docker Images
```bash
# Build image
docker build -t devopsprojectenvsync:latest .

# Run container
docker run -p 8080:8080 -e ENVSYNC_KEY=$ENVSYNC_KEY devopsprojectenvsync:latest
```

### Kubernetes Deployment
```bash
# Apply K8s manifests
kubectl apply -f infrastructure/kubernetes/

# Check deployment status
kubectl get pods,svc,deploy
```

---

## 📈 Performance Metrics

| Metric | Target | Current |
|--------|--------|---------|
| Build Time | < 5 min | ~45 sec |
| Test Coverage | > 80% | 90% |
| Deployment Frequency | Daily | On every push |
| Mean Time to Recovery | < 1 hour | ~2 min (rollback) |
| Drift Detection Speed | < 1 sec | ~200ms |

---

## 📚 Documentation

### User Documentation
- [User Guide](docs/userguide.md)
- [API Documentation](docs/apidocumentation.md)

### Technical Documentation
- [Design Document](docs/designdocument.md)
- [CI/CD Setup Guide](CICD_SETUP.md)

---

## 🎥 Demo

### Demo Video
[Link to 5–10 minute demo video in deliverables/]

### CLI Commands Demo
```bash
./envsync diff dev staging
./envsync audit --env staging --fail-on-missing
./envsync sync dev staging --dry-run
./envsync sync dev staging
./envsync snapshot create staging
./envsync rollback staging
./envsync validate --env dev
```

---

## 🌿 Development Workflow

### Git Branching Strategy
```
main
├── develop
│   ├── feature/drift-detection
│   ├── feature/crypto-layer
│   ├── feature/web-dashboard
│   └── hotfix/sync-conflict-fix
└── release/v1.0.0
```

### Commit Convention
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `test:` Test-related
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

---

## 🔐 Security

### Security Measures Implemented
- [x] AES-256-GCM encryption for all snapshots at rest
- [x] Secrets never stored in plain text
- [x] Sensitive key values masked in all terminal output
- [x] Strict Mode for production requiring explicit confirmation
- [x] GitHub Environment protection rules for production
- [x] `.env.*` files excluded from git via `.gitignore`

### Security Scanning
```bash
trivy image devopsprojectenvsync:latest
```

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'feat: add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

## 💡 Project Challenges

1. **Cross-platform secret encryption** — Solved by using AES-256-GCM with key derived from environment variable, never touching the filesystem
2. **Conflict resolution during sync** — Solved by building an interactive `[S]ource / [T]arget / [K]eep` prompt and `--overwrite` flag for CI/CD
3. **Preventing accidental production overwrites** — Solved by Strict Mode + GitHub Environment required reviewers as a two-layer approval gate

## 📖 Learnings
- How configuration drift causes real-world production incidents and how to detect it programmatically
- AES-256-GCM encryption in Go and the importance of never persisting decryption keys to disk
- Building reusable GitHub Actions workflows for DRY CI/CD pipelines

---

## 🙏 Acknowledgments

- Course Instructor: **Mr. Jay Shankar Sharma**
- HashiCorp Vault documentation and open-source community
- Go standard library and Cobra CLI framework contributors
- Reference materials and DevOps best practice guides

---

## 📬 Contact

**Student:** Vishakha Singh
**Email:** vishakha.singh@university.edu
**GitHub:** https://github.com/vishakhasingh
**Registration No:** 23FE10CSE00134

**Course Coordinator:** Mr. Jay Shankar Sharma
**Consultation Hours:** Thursday & Friday, 5–6 PM, LHC 308F
  
