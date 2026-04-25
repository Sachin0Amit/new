# --- Stage 1: Build Polyglot Core ---
FROM golang:1.23-alpine AS builder

# Install C++ build toolchain and dependencies
RUN apk add --no-cache build-base cmake make g++ gcc musl-dev

WORKDIR /app

# Copy Go dependencies first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy full source
COPY . .

# Build C++ Titan Engine Core
RUN mkdir -p build && cd build && \
    cmake ../cpp && \
    make -j$(nproc) && \
    cp libtitan.a ../internal/titan/

# Build Go Sovereign Orchestrator with CGO
# Note: We link against musl on Alpine; for Debian runtime, we might need a glibc builder
# but following user instruction to use golang:1.23-alpine.
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o sovereign ./cmd/sovereign/main.go

# --- Stage 2: Final Production Image ---
FROM debian:bookworm-slim

# Create non-root user
RUN groupadd -g 1001 sovereign && \
    useradd -m -u 1001 -g sovereign sovereign

WORKDIR /app

# Copy binary and web assets
COPY --from=builder /app/sovereign /app/sovereign
COPY --from=builder /app/web/src /app/web/src
COPY --from=builder /app/web/dist /app/web/dist

# Set permissions
RUN chown -R sovereign:sovereign /app

# Switch to non-root user
USER 1001

# Expose the orchestrator port
EXPOSE 8081

# Healthcheck
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8081/health || exit 1

# Runtime settings
ENV ENV=prod
ENV SOVEREIGN_PORT=8081

ENTRYPOINT ["./sovereign"]
