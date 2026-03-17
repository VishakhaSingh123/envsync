# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache module downloads
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /envsync ./...

# ── Stage 2: Minimal runtime image ────────────────────────────────────────────
FROM scratch

# Copy CA certs for HTTPS (needed for Vault/AWS API calls)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /envsync /envsync

# The config and env files are mounted at runtime
VOLUME ["/workspace"]
WORKDIR /workspace

ENTRYPOINT ["/envsync"]
CMD ["--help"]

# Usage:
#   docker build -t envsync .
#   docker run --rm \
#     -v $(pwd):/workspace \
#     -e ENVSYNC_KEY=$ENVSYNC_KEY \
#     envsync diff dev staging
