# syntax=docker/dockerfile:1.7

# --- Build stage --------------------------------------------------------------
# Pinned to the Go version recorded in go.mod. CGO is disabled so the binary
# is fully static (works because modernc.org/sqlite is pure Go).
FROM golang:1.26-alpine AS build

WORKDIR /src

# Cache go modules in their own layer for fast incremental rebuilds.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Source + build.
COPY . .

ARG VERSION=dev
ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath \
        -ldflags="-s -w -X main.version=${VERSION}" \
        -o /out/kidsboard \
        .

# --- Runtime stage ------------------------------------------------------------
# Distroless static-nonroot: ~2MB base, no shell, no package manager, UID 65532.
# Pair with PSP/Pod securityContext.runAsNonRoot=true for defense in depth.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/kidsboard /usr/local/bin/kidsboard

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/kidsboard"]
CMD ["serve", "--addr", ":8080", "--db", "/data/kidsboard.db"]
