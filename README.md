# EnvSync — Environment Synchronization Tool

> Detect drift • Sync secrets • Validate runtimes • Never break prod again.

EnvSync is a CLI tool and dashboard that audits and synchronizes environment variables, infrastructure state, and configuration files across **Dev → Staging → Production**. It solves the classic *"it works on my machine"* problem caused by configuration drift.

---

## Table of Contents
1. [Architecture](#architecture)
2. [Phase 1 — Discovery & Architecture](#phase-1--discovery--architecture)
3. [Phase 2 — Core Engine (Diff Logic)](#phase-2--core-engine-diff-logic)
4. [Phase 3 — Integration & Execution (Sync Logic)](#phase-3--integration--execution-sync-logic)
5. [Phase 4 — CI/CD Automation & Safety](#phase-4--cicd-automation--safety)
6. [Installation](#installation)
7. [Configuration Reference](#configuration-reference)
8. [CLI Command Reference](#cli-command-reference)
9. [Security](#security)
10. [Dashboard](#dashboard)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        EnvSync Architecture                      │
├───────────────┬──────────────────┬──────────────────────────────┤
│     Dev       │     Staging      │        Production            │
│  .env.dev     │  .env.staging    │    .env.production           │
│  (SSH/file)   │  (SSH/AWS SSM)   │    (Vault/AWS SSM)           │
└───────┬───────┴────────┬─────────┴──────────────┬──────────────┘
        │                │                         │
        └────────────────┼─────────────────────────┘
                         │
              ┌──────────▼──────────┐
              │    EnvSync Engine   │
              │  ┌───────────────┐  │
              │  │    Parser     │  │  Reads .env, YAML, JSON
              │  ├───────────────┤  │
              │  │  Comparator   │  │  Produces Drift Reports
              │  ├───────────────┤  │
              │  │  Crypto Layer │  │  AES-256-GCM encryption
              │  ├───────────────┤  │
              │  │  Sync Engine  │  │  Conflict resolution + apply
              │  ├───────────────┤  │
              │  │  Snapshot Mgr │  │  Auto-backup before every sync
              │  └───────────────┘  │
              └──────────┬──────────┘
                         │
              ┌──────────▼──────────┐
              │   CI/CD Pipeline    │
              │  GitHub Actions     │
              │  Pre-flight Audits  │
              │  Drift Gate         │
              │  Approval Flows     │
              └─────────────────────┘
```

---

## Phase 1 — Discovery & Architecture

### Source of Truth
The `.env.example` file defines the **canonical set of keys** every environment must have. Any key present here but absent in an environment is flagged as `MISSING`.

### State File Format
EnvSync uses `envsync.yaml` as the state descriptor:
```yaml
version: "1"
source_of_truth: .env.example
environments:
  dev:
    type: file
    path: .env.dev
  staging:
    type: file        # or: ssh | aws_ssm | vault
    path: .env.staging
  production:
    type: file
    path: .env.production
```

### Supported Environment Sources
| Type | Description |
|------|-------------|
| `file` | Local `.env`, YAML, or JSON file |
| `ssh` | Remote server over SSH using `~/.ssh/deploy_key` |
| `aws_ssm` | AWS SSM Parameter Store (requires aws-sdk-go-v2) |
| `vault` | HashiCorp Vault KV (v2 engine) |

---

## Phase 2 — Core Engine (Diff Logic)

### How the Diff Works
The comparator takes two `map[string]string` environments and returns a `DriftReport` with entries classified as:

| Status | Meaning |
|--------|---------|
| `MISSING` | Key exists in source but **not** in target |
| `MISMATCH` | Key exists in both but values **differ** |
| `EXTRA` | Key exists in target but **not** in source |
| `MATCH` | Key and value are **identical** |

### Sensitive Value Masking
Any key containing `password`, `secret`, `token`, `key`, `private`, `credential`, or `auth` has its value automatically masked in all output (e.g. `ab**********yz`). **Plaintext secrets are never written to disk or logs.**

### Output Formats
```bash
envsync diff dev staging              # Human-readable table
envsync diff dev staging --output json   # Machine-readable JSON
envsync diff dev staging --output yaml   # YAML format
```

---

## Phase 3 — Integration & Execution (Sync Logic)

### Push (Dev → Staging)
```bash
envsync sync dev staging
```
1. Runs a diff and shows the drift report
2. Auto-snapshots `staging` before any changes
3. For conflicts, interactively prompts: **[S]ource / [T]arget / [K]eep**
4. Applies changes and writes to target

### Pull (Staging → Dev)
```bash
envsync sync staging dev
```
Pulls the non-sensitive architecture config from staging into dev.

### Specific Keys Only
```bash
envsync sync dev staging --keys DB_HOST,REDIS_URL,APP_PORT
```

### Dry Run (Preview Without Applying)
```bash
envsync sync dev staging --dry-run
```

### Conflict Resolution
When a key exists in both environments but has different values, EnvSync prompts:
```
⚡ CONFLICT: DB_HOST
  [S] Source value: localhost
  [T] Target value: staging-db.internal
  [K] Keep target (skip)
Choose [S/T/K]:
```

---

## Phase 4 — CI/CD Automation & Safety

### GitHub Actions Setup

#### Step 1: Add Repository Secrets
Go to **Settings → Secrets and Variables → Actions** and add:

| Secret | Description |
|--------|-------------|
| `ENV_DEV` | Full contents of your `.env.dev` file |
| `ENV_STAGING` | Full contents of your `.env.staging` file |
| `ENV_PRODUCTION` | Full contents of your `.env.production` file |
| `ENVSYNC_KEY` | Encryption key: `openssl rand -base64 32` |

#### Step 2: Create GitHub Environments
Go to **Settings → Environments** and create:

- **`staging`** — no protection rules (auto-deploy from `staging` branch)
- **`production`** — enable **Required Reviewers** (add your team leads)

This enforces the PR-style approval flow for production syncs.

#### Step 3: Branch Strategy
```
feature/* → develop → staging branch → main branch
                          │                  │
                          ▼                  ▼
                    Deploy Staging    Deploy Production
                    (auto)            (requires approval)
```

#### Step 4: Pipeline Stages
The `.github/workflows/envsync.yml` workflow runs:

```
On every PR:
  ┌─ build ─────────────────────────────────┐
  │  Compiles binary, runs unit tests        │
  └──────────────────────────────────────────┘
       │
       ├─ preflight-audit (dev + staging) ───┐
       │  Checks all required keys exist      │
       │  Fails build if keys are missing     │
       └──────────────────────────────────────┘
       │
       └─ drift-detection ───────────────────┐
          Runs diff dev → staging             │
          Posts report as PR comment          │
          (Non-blocking — informs, not fails) │
          └──────────────────────────────────┘

On push to `staging` branch:
  ┌─ deploy-staging ────────────────────────┐
  │  1. Snapshot staging                     │
  │  2. Dry-run sync                         │
  │  3. Apply sync                           │
  │  4. Validate runtimes                    │
  └──────────────────────────────────────────┘

On push to `main` branch (after staging):
  ┌─ deploy-production (APPROVAL REQUIRED) ─┐
  │  1. Snapshot production                  │
  │  2. Zero-tolerance audit                 │
  │  3. Dry-run sync                         │
  │  4. Apply sync (strict mode)             │
  └──────────────────────────────────────────┘

Manual dispatch:
  ┌─ rollback ──────────────────────────────┐
  │  Restores staging OR production          │
  │  to most recent snapshot                 │
  └──────────────────────────────────────────┘
```

#### Step 5: Drift Gate (Blocking)
The audit runs with `--threshold 5` — if more than 5 keys are drifted, the build **fails** before deployment. For production, it uses `--threshold 0` (zero tolerance).

---

## Installation

### Build from Source
```bash
git clone https://github.com/your-org/envsync.git
cd envsync
go build -o bin/envsync .
sudo mv bin/envsync /usr/local/bin/envsync
```

### Quick Start
```bash
# 1. Scaffold config
envsync init

# 2. Edit envsync.yaml to point at your env files

# 3. Generate encryption key
export ENVSYNC_KEY=$(openssl rand -base64 32)

# 4. Run your first audit
envsync audit --env dev

# 5. Compare dev vs staging
envsync diff dev staging

# 6. Sync (with dry-run first)
envsync sync dev staging --dry-run
envsync sync dev staging
```

---

## Configuration Reference

```yaml
version: "1"

source_of_truth: .env.example   # Keys every env must have

environments:
  dev:
    type: file                   # file | ssh | aws_ssm | vault
    path: .env.dev
    format: dotenv               # dotenv | yaml | json

  staging:
    type: ssh
    remote:
      host: staging.example.com
      user: deploy
      key_file: ~/.ssh/id_rsa

  production:
    type: aws_ssm
    remote:
      region: us-east-1
      profile: myapp-prod
      vault_path: /myapp/production/

runtimes:
  node: "20"                     # Prefix match: "20" matches "20.11.1"
  python: "3.11"

secrets:
  encryption_key_env: ENVSYNC_KEY   # env var holding AES key
  redacted_keys:                    # always mask these in output
    - PASSWORD
    - SECRET
    - TOKEN

snapshots:
  directory: .envsync/snapshots  # where snapshots are stored
  max_keep: 10                   # prune older than 10 snapshots
  encrypted: true                # encrypt snapshots at rest
```

---

## CLI Command Reference

```
envsync diff <env1> <env2>          Compare two environments
  --output table|json|yaml          Output format (default: table)

envsync sync <source> <target>      Sync source into target
  --dry-run                         Preview without applying
  --overwrite                       Overwrite conflicts without prompting
  --keys KEY1,KEY2                  Sync specific keys only
  --strict                          Require confirmation for production

envsync audit                       Audit against .env.example
  --env <name>                      Environment to audit (default: dev)
  --fail-on-missing                 Exit 1 if any keys missing
  --threshold <n>                   Max allowed drift count

envsync snapshot create <env>       Take a snapshot
envsync snapshot list <env>         List all snapshots

envsync rollback <env>              Rollback to latest snapshot
  --id <snap_id>                    Rollback to specific snapshot

envsync validate                    Check runtime versions
  --env <name>                      Environment to validate

envsync init                        Scaffold envsync.yaml
```

---

## Security

### Encryption
- All snapshots are encrypted at rest using **AES-256-GCM**
- The encryption key is derived from `ENVSYNC_KEY` env var — **never stored on disk**
- Sensitive values are never logged, even in verbose mode

### Strict Mode for Production
```bash
envsync sync staging production --strict
```
In strict mode, syncing to production requires an explicit confirmation prompt. In CI/CD, use GitHub's **Required Reviewers** on the `production` environment as the approval gate.

### Key Never Written to Disk
The encryption passphrase is read from an environment variable at runtime and immediately used to derive a key — it is **never persisted** to the filesystem in any form.

---

## Dashboard

A web dashboard is available in `web/dashboard/`. Open `index.html` in a browser or serve it:

```bash
cd web/dashboard
python3 -m http.server 8080
# Open: http://localhost:8080
```

The dashboard provides:
- Visual drift heatmap across all environments
- Side-by-side key comparison
- Snapshot history timeline
- One-click sync and rollback actions

---

## Project Structure

```
envsync/
├── main.go                        Entry point
├── go.mod                         Go module
├── envsync.yaml                   Config (edit this)
├── .env.example                   Source of truth
├── .env.dev                       Dev environment
├── .env.staging                   Staging environment
├── cmd/
│   ├── root.go                    Root cobra command
│   ├── diff.go                    envsync diff
│   ├── sync.go                    envsync sync
│   ├── audit.go                   envsync audit
│   ├── snapshot.go                envsync snapshot
│   ├── rollback.go                envsync rollback
│   └── validate.go                envsync validate + init
├── internal/
│   ├── parser/parser.go           Config & env file parser
│   ├── comparator/comparator.go   Diff engine
│   ├── crypto/crypto.go           AES-256-GCM encryption
│   ├── sync/sync.go               Sync plan + apply
│   └── snapshot/snapshot.go       Snapshot manager
├── web/dashboard/
│   └── index.html                 Web dashboard
├── scripts/
│   └── install.sh                 Install script
└── .github/
    └── workflows/
        └── envsync.yml            CI/CD pipeline
```
Development branch 
Version 1.0.0 
