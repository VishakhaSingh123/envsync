#!/bin/sh
# scripts/vault-init.sh
# Initializes a local dev Vault with sample secrets for EnvSync testing.
# Run after: docker compose up vault

set -e

VAULT_ADDR="${VAULT_ADDR:-http://localhost:8200}"
VAULT_TOKEN="${VAULT_TOKEN:-root}"

export VAULT_ADDR VAULT_TOKEN

echo "→ Enabling KV secrets engine..."
vault secrets enable -path=secret kv-v2 2>/dev/null || true

echo "→ Writing dev secrets..."
vault kv put secret/myapp/dev \
  APP_ENV=development \
  DB_PASSWORD=devpassword \
  JWT_SECRET=dev-jwt-secret \
  STRIPE_SECRET_KEY=sk_test_dev

echo "→ Writing staging secrets..."
vault kv put secret/myapp/staging \
  APP_ENV=staging \
  DB_PASSWORD=stagingpassword \
  JWT_SECRET=staging-jwt-secret \
  STRIPE_SECRET_KEY=sk_test_staging

echo "→ Writing production secrets..."
vault kv put secret/myapp/production \
  APP_ENV=production \
  DB_PASSWORD=CHANGE_IN_REAL_PROD \
  JWT_SECRET=CHANGE_IN_REAL_PROD \
  STRIPE_SECRET_KEY=sk_live_CHANGE_IN_REAL_PROD

echo "✔ Vault initialized. Access UI at: $VAULT_ADDR/ui"
echo "  Token: $VAULT_TOKEN"
