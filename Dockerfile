# syntax=docker/dockerfile:1.7
#
# Multi-arch container for kidsboard. Cross-compiles via Go (no QEMU during
# build) so building an arm64 / armv7 image on an amd64 host is fast.
#
# Local build for the host's architecture:
#   docker build -t kidsboard:dev .
#
# Build for a Raspberry Pi 4/5 (arm64) and load into the local daemon:
#   docker buildx build --platform=linux/arm64 -t kidsboard:dev --load .
#
# Multi-arch build pushed to a registry (manifest list, both arches):
#   docker buildx build --platform=linux/amd64,linux/arm64 \
#     -t your-registry/kidsboard:0.1.0 --push .
#
# Pi Zero W / Pi 1 (32-bit ARMv6) is supported via linux/arm/v6 if needed.

ARG GO_VERSION=1.26-alpine
ARG RUNTIME_IMAGE=gcr.io/distroless/static-debian12:nonroot
# Standalone Tailwind CLI version. Bundles the forms + container-queries
# plugins so kidsboard doesn't need Node or an npm install during build.
ARG TAILWIND_VERSION=v3.4.17

# --- Build stage --------------------------------------------------------------
# `--platform=$BUILDPLATFORM` pins this stage to the host architecture. The Go
# toolchain then cross-compiles to $TARGETPLATFORM via GOOS/GOARCH/GOARM. This
# avoids running the compiler under QEMU emulation, which would be 10–30×
# slower on a typical CI box.
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build

WORKDIR /src

# Cache modules in their own layer. Re-runs only when go.mod/go.sum change.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG TAILWIND_VERSION

# Build the Tailwind stylesheet BEFORE `go build` so it gets embedded into the
# binary via `//go:embed static` in internal/view/static.go. The output CSS is
# plain text and arch-independent, so we always download the BUILDPLATFORM
# variant (this stage runs on the host's arch, not the target's).
#
# alpine ships musl libc, but the standalone Tailwind binary is statically
# linked — the plain `linux-x64` / `linux-arm64` asset works on alpine.
RUN --mount=type=cache,target=/var/cache/tailwind \
    apk add --no-cache curl ca-certificates && \
    BUILD_ARCH=$(uname -m); \
    case "$BUILD_ARCH" in \
      x86_64|amd64)  TW_ARCH=x64 ;; \
      aarch64|arm64) TW_ARCH=arm64 ;; \
      *) echo "unsupported build arch: $BUILD_ARCH" >&2; exit 1 ;; \
    esac; \
    TW_URL="https://github.com/tailwindlabs/tailwindcss/releases/download/${TAILWIND_VERSION}/tailwindcss-linux-${TW_ARCH}"; \
    TW_CACHE="/var/cache/tailwind/tailwindcss-${TAILWIND_VERSION}-linux-${TW_ARCH}"; \
    if [ ! -x "$TW_CACHE" ]; then \
      echo "downloading $TW_URL"; \
      curl -fsSL -o "$TW_CACHE" "$TW_URL"; \
      chmod +x "$TW_CACHE"; \
    fi; \
    mkdir -p internal/view/static/css; \
    "$TW_CACHE" -i input.css -o internal/view/static/css/kidsboard.css --minify; \
    wc -c internal/view/static/css/kidsboard.css

# GOARM is set from the platform variant (linux/arm/v7 → "v7" → "7").
# Empty for non-ARM (amd64, arm64) — Go ignores GOARM unless GOARCH=arm.
# CGO disabled because modernc.org/sqlite is pure Go: yields a fully-static
# binary that runs on any libc-free runtime (distroless, scratch, FROM_BUSYBOX).
# -trimpath, -s, -w, -buildid= strip absolute paths and debug metadata for
# smaller, more reproducible binaries.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=$(echo "${TARGETVARIANT}" | sed 's/^v//') \
    go build -trimpath \
        -ldflags="-s -w -buildid= -X main.version=${VERSION}" \
        -o /out/kidsboard \
        .

# Quick sanity check: confirm we built for the requested platform. Helpful
# diagnostic when the buildx flags don't propagate cleanly.
RUN /usr/local/go/bin/go version /out/kidsboard

# --- Runtime stage ------------------------------------------------------------
# Distroless static-debian12:nonroot:
#   • ~2 MB base, no shell, no package manager, no apt cache to update
#   • Multi-arch (amd64, arm64, armv7) — Pi-compatible out of the box
#   • UID/GID 65532 baked in for runAsNonRoot security contexts
#   • CA certificates included for outbound TLS (not needed by kidsboard today
#     but harmless and future-proofs HTTPS calls)
FROM ${RUNTIME_IMAGE}

ARG VERSION=dev

# OCI image labels — surface metadata to `docker inspect`, registries (ghcr,
# GHCR-style indexes), and supply-chain tooling.
LABEL org.opencontainers.image.title="kidsboard" \
      org.opencontainers.image.description="RPG-style household activity tracker for kids" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.source="https://github.com/juliancantillo/kidsboard" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="cantillo.dev"

COPY --from=build /out/kidsboard /usr/local/bin/kidsboard

# Distroless doesn't have a shell, so the binary's own healthcheck subcommand
# is the only way to do an internal probe. Matches the K8s readinessProbe path.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/usr/local/bin/kidsboard", "healthcheck", "--url", "http://127.0.0.1:8080/healthz", "--timeout", "3s"]

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/kidsboard"]
# Defaults to `serve`. Per 12-factor, every config knob is also a
# KIDSBOARD_* env var — override the listen address with KIDSBOARD_ADDR,
# the DB path with KIDSBOARD_DB, log format with KIDSBOARD_LOG_FORMAT, etc.
CMD ["serve"]
