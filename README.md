# Environment Synchronization Tool (EnvSync)

**Student Name:** Vishakha Singh  
**Registration No:** 23FE10CSE00134  
**Course:** CSE3253 DevOps [PE6]  
**Semester:** VI (2025–2026)  
**Project Type:** DevOps Fundamentals & Ecosystem  

---

## Problem Statement

In software development, teams maintain multiple environments — **Dev**, **Staging**, and **Production**. Over time, manual changes cause **configuration drift** — environments behave differently from each other. This leads to the classic *"it works on my machine"* problem, failed deployments, and production bugs.

**EnvSync** solves this by comparing and synchronizing environment variables across all environments.

---

## Features

| Feature | Description |
|---------|-------------|
| **Drift Detection** | Compare two environments and get a report showing MISSING, MISMATCH, and EXTRA keys |
| **Environment Audit** | Check any environment against a source of truth (`.env.example`) |
| **Sync** | Copy missing/mismatched keys from one environment to another |
| **Snapshot & Rollback** | Auto-save environment state before sync; restore if something goes wrong |
| **Runtime Validation** | Verify Node.js, Python versions match the required spec |
| **Secret Masking** | Sensitive keys (passwords, tokens) are never displayed in plain text |
| **AES-256 Encryption** | Snapshots are encrypted using AES-256-GCM |

---

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Python 3.11+ |
| CLI Framework | Click |
| Encryption | cryptography (AES-256-GCM) |
| Config Format | YAML (PyYAML) |
| Output Formatting | tabulate, colorama |

---

## Project Structure

```
envsync/
│
├── main.py                    # Entry point — registers all CLI commands
├── envsync.yaml               # Master config — points to all environments
├── requirements.txt           # Python dependencies
├── .gitignore                 # Files to exclude from Git
│
├── .env.example               # Source of truth (all required keys)
├── .env.dev                   # Dev environment variables
├── .env.staging               # Staging environment variables
├── .env.production            # Production environment variables
│
├── core/                      # Core logic modules
│   ├── parser/parser.py       # Reads & writes .env, .yaml, .json files
│   ├── comparator/comparator.py  # Compares two envs, detects drift
│   ├── crypto/crypto.py       # AES-256-GCM encryption/decryption
│   ├── sync/sync.py           # Builds sync plan & applies changes
│   └── snapshot/snapshot.py   # Create, list, restore, prune snapshots
│
├── commands/                  # CLI commands (user-facing)
│   ├── root.py                # Base CLI group + helper functions
│   ├── diff.py                # `diff` command — compare two envs
│   ├── sync.py                # `sync` command — sync source → target
│   ├── audit.py               # `audit` command — check env health
│   ├── snapshot.py            # `snapshot` command — save env state
│   ├── rollback.py            # `rollback` command — restore snapshot
│   └── validate.py            # `validate` command — check runtimes
│
└── web/dashboard/index.html   # Web dashboard for visual monitoring
```

---

## Installation & Setup

### Method A: Local Python (Traditional)
```bash
# 1. Install Python dependencies
pip install -r requirements.txt

# 2. Set encryption key (PowerShell)
$env:ENVSYNC_KEY = [Convert]::ToBase64String((1..32 | ForEach-Object { [byte](Get-Random -Max 256) }))

# 3. Run the tool
python main.py --help
```

### Method B: Docker (Recommended)
No local Python installation required! We use a Docker Volume Mount to sync your local `.env` files.
```bash
# 1. Build the image
docker-compose build

# 2. Run the tool (Linux/macOS)
export ENVSYNC_KEY="your_32_character_secret_key_here_"
docker-compose run --rm envsync --help

# Or on Windows PowerShell:
$env:ENVSYNC_KEY="your_32_character_secret_key_here_"
docker-compose run --rm envsync --help
```

---

## Usage Examples

> **Note:** If using Docker, simply replace `python main.py` in the examples below with `docker-compose run --rm envsync`.

```bash
# Compare dev and staging environments
python main.py diff dev staging

# Audit dev environment against source of truth
python main.py audit --env dev

# Sync dev → staging (with preview)
python main.py sync dev staging --dry-run

# Sync dev → staging (apply changes)
python main.py sync dev staging --overwrite

# Take a snapshot of staging
python main.py snapshot staging

# Rollback staging to last snapshot
python main.py rollback staging

# Validate runtime versions
python main.py validate --env dev
```

---

## How It Works

1. **Config Loading** — `envsync.yaml` defines where each environment file lives
2. **Parsing** — The parser reads `.env` files into key-value dictionaries
3. **Comparison** — The comparator finds MISSING, MISMATCH, EXTRA, and MATCH keys
4. **Sync Plan** — For mismatches, user chooses Source/Target/Keep (or `--overwrite`)
5. **Snapshot** — Before any sync, the current state is saved (encrypted with AES-256)
6. **Apply** — Changes are written to the target environment file
7. **Rollback** — If something breaks, restore from the last snapshot

---

## CI/CD Pipeline

This project includes a **GitHub Actions** CI/CD pipeline integrated via `.github/workflows/ci.yml`.  
Every time code is pushed or a Pull Request is opened, the pipeline automatically:
1. Provisions an Ubuntu environment.
2. Sets up **Python 3.11** with dependency caching.
3. Installs requirements and verifies the CLI runs without import or syntax errors.
4. Builds the **Docker Image** to ensure the `Dockerfile` integrity is maintained.

---

## Acknowledgments

- **Course Instructor:** Mr. Jay Shankar Sharma
- Built with Python, Click CLI framework, and the cryptography library

---

## License

This project is licensed under the MIT License.
