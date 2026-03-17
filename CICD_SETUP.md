# CI/CD Setup Guide — EnvSync

This guide walks you through integrating EnvSync into GitHub Actions from zero to a fully automated,
drift-gated deployment pipeline.

---

## Architecture Overview

```
Developer Push
      │
      ▼
┌─────────────────────────────────────────────────────────┐
│                  GitHub Actions Pipeline                  │
│                                                           │
│  ① Build & Test ──────────────────────────────────────  │
│     go build + go test ./...                              │
│                                                           │
│  ② Preflight Audit (all environments) ────────────────  │
│     envsync audit --env dev --fail-on-missing             │
│     envsync audit --env staging --threshold 5             │
│                                                           │
│  ③ Drift Detection ────────────────────────────────────  │
│     envsync diff dev staging --output json                │
│     Posts diff as PR comment                              │
│     Exit code 2 = drift (non-blocking on PR)              │
│                                                           │
│  ④ Deploy → Staging (auto, on push to staging branch)   │
│     envsync snapshot create staging                       │
│     envsync sync dev staging --overwrite                  │
│     envsync validate --env staging                        │
│                                                           │
│  ⑤ Deploy → Production (REQUIRES APPROVAL)              │
│     [GitHub Environment: production + Required Reviewers] │
│     envsync snapshot create production                    │
│     envsync audit --env production --threshold 0          │
│     envsync sync staging production --strict              │
│                                                           │
│  ⑥ Emergency Rollback (manual dispatch only)            │
│     envsync rollback [staging|production]                 │
└─────────────────────────────────────────────────────────┘
```

---

## Step 1 — Fork / Clone the Repo

```bash
git clone https://github.com/your-org/envsync.git
cd your-project
cp -r envsync/.github .
cp envsync/envsync.yaml .
cp envsync/.env.example .
```

---

## Step 2 — Configure envsync.yaml

Edit `envsync.yaml` to point at your actual environment files or remote sources:

```yaml
version: "1"
source_of_truth: .env.example

environments:
  dev:
    type: file
    path: .env.dev
  staging:
    type: file
    path: .env.staging
  production:
    type: file
    path: .env.production
```

---

## Step 3 — Add GitHub Repository Secrets

Navigate to: **GitHub Repo → Settings → Secrets and Variables → Actions → New Repository Secret**

| Secret Name       | Value                                        | Notes                              |
|-------------------|----------------------------------------------|------------------------------------|
| `ENV_DEV`         | Full contents of `.env.dev`                  | Paste raw file content             |
| `ENV_STAGING`     | Full contents of `.env.staging`              | Paste raw file content             |
| `ENV_PRODUCTION`  | Full contents of `.env.production`           | Paste raw file content             |
| `ENVSYNC_KEY`     | `$(openssl rand -base64 32)`                 | AES-256 encryption key             |

**Generate the encryption key locally:**
```bash
openssl rand -base64 32
# Copy the output and paste as the ENVSYNC_KEY secret value
```

---

## Step 4 — Create GitHub Environments

Navigate to: **GitHub Repo → Settings → Environments → New Environment**

### `staging` environment
- Name: `staging`
- No protection rules (auto-deploys from `staging` branch)
- Optional: Add deployment branch rule: `staging`

### `production` environment
- Name: `production`
- ✅ Enable **Required Reviewers** — add 1–2 team leads
- ✅ Enable **Prevent self-review** (recommended)
- Set deployment branch rule: `main` only

This means every production sync requires an explicit approval click in GitHub — this IS your "Strict Mode" gate.

---

## Step 5 — Branch Strategy

```
feature/my-change
      │
      ▼  PR → develop
  develop ──────────────────────────────────────────▶ staging branch
                                                           │
                                                           │ auto-deploy to staging
                                                           │ envsync sync dev → staging
                                                           ▼
                                                      staging environment
                                                           │
                                                           │ PR merge → main
                                                           ▼
                                                       main branch
                                                           │
                                                           │ deploy to production
                                                           │ REQUIRES APPROVAL
                                                           │ envsync sync staging → prod
                                                           ▼
                                                     production environment
```

---

## Step 6 — Understand the Workflow Files

### `.github/workflows/envsync.yml` — Main Pipeline

| Job | Trigger | What it does |
|-----|---------|-------------|
| `build` | All pushes + PRs | Compiles binary, runs tests, uploads artifact |
| `preflight-audit` | After build | Audits dev + staging for missing keys |
| `drift-detection` | After build, PRs | Diffs dev→staging, comments on PR |
| `deploy-staging` | Push to `staging` branch | Snapshot + sync + validate |
| `deploy-production` | Push to `main` | Snapshot + zero-tolerance audit + sync (approval required) |
| `rollback` | Manual dispatch only | Restore staging or production to latest snapshot |

### `.github/workflows/drift-check.yml` — PR Drift Guard

Runs on every PR. Posts a drift table as a PR comment:

```
### 🔍 EnvSync Drift Report

⚠️ 2 drift(s) detected:

| Key               | Status   |
|-------------------|----------|
| MAINTENANCE_MODE  | 🔴 MISSING  |
| ENABLE_PAYMENTS   | 🟡 MISMATCH |
```

### `.github/workflows/pre-deploy.yml` — Reusable Preflight

A reusable workflow (`workflow_call`) that can be called from any pipeline:

```yaml
# In your own workflow:
jobs:
  preflight:
    uses: ./.github/workflows/pre-deploy.yml
    with:
      target_env: staging
      drift_threshold: 5
    secrets:
      ENVSYNC_KEY: ${{ secrets.ENVSYNC_KEY }}
```

---

## Step 7 — Drift Thresholds

| Environment | Threshold | Meaning |
|-------------|-----------|---------|
| `dev`       | `--threshold 10` | 10 drift keys allowed (dev is loose) |
| `staging`   | `--threshold 5`  | 5 drift keys allowed before CI fails |
| `production`| `--threshold 0`  | Zero tolerance — any drift blocks deploy |

---

## Step 8 — Testing the Pipeline Locally

```bash
# Build
go build -o envsync .

# Export encryption key
export ENVSYNC_KEY=$(openssl rand -base64 32)

# Run audit (mimics CI)
./envsync audit --env staging --fail-on-missing --threshold 5

# Run diff (exit code 2 = drift)
./envsync diff dev staging; echo "Exit: $?"

# Full sync dry-run
./envsync sync dev staging --dry-run

# Snapshot before sync (CI does this automatically)
./envsync snapshot create staging

# Apply sync
./envsync sync dev staging --overwrite
```

---

## Step 9 — Emergency Rollback via GitHub Actions

1. Go to **Actions → EnvSync CI/CD Pipeline**
2. Click **Run workflow**
3. Select the environment: `staging` or `production`
4. Click **Run workflow**

This triggers the `rollback` job which runs:
```bash
envsync rollback [environment]
```

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `encryption key not found` | Add `ENVSYNC_KEY` to GitHub secrets |
| `environment 'staging' not defined` | Check `envsync.yaml` has a `staging:` entry |
| `could not read .env.staging` | Ensure `ENV_STAGING` secret is set and written to file |
| Drift blocks deploy | Run `envsync diff` locally and fix missing keys |
| Production approval stuck | Check GitHub Environment → Required Reviewers config |

---

## Security Checklist

- [ ] `ENVSYNC_KEY` is set as a GitHub Secret (never hardcoded)
- [ ] `.env.*` files are in `.gitignore`  
- [ ] Production environment has Required Reviewers enabled
- [ ] Snapshots directory (`.envsync/snapshots/`) is in `.gitignore`
- [ ] `envsync.yaml` does NOT contain actual secret values (only paths/types)
- [ ] Encryption is enabled in `envsync.yaml` (`snapshots.encrypted: true`)
