FROM node:20-alpine AS agent-builder

WORKDIR /agent

# Copy agent package files
COPY agent/package.json agent/pnpm-lock.yaml ./
RUN corepack enable pnpm && pnpm install --frozen-lockfile

# Copy agent source and build
COPY agent/ ./
RUN pnpm build

FROM golang:1.25-alpine AS api-builder

RUN apk add --no-cache git

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o signet ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from api builder
COPY --from=api-builder /build/signet .

# Copy agent build files from agent builder
COPY --from=agent-builder /agent/dist ./agent/dist

# Expose port
EXPOSE 6969

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:6969/health || exit 1

# Run
CMD ["./signet"]
