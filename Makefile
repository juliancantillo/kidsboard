.PHONY: check test build generate run seed clean docker docker-pi docker-multiarch buildx-bootstrap helm-lint helm-template

# Override on the command line: `make docker-pi REGISTRY=ghcr.io/me VERSION=0.1.0`
REGISTRY  ?= kidsboard
VERSION   ?= dev
IMAGE     ?= $(REGISTRY):$(VERSION)
PLATFORMS ?= linux/amd64,linux/arm64

# Default target: run tests, then build. Used as the single CI/local
# verification step — anything green here is shippable.
check: test build

test:
	go test ./...

build:
	go build ./...

# Regenerate sqlc code from queries/ + migrations/. Run after schema or
# query changes; commit the generated output under internal/storage/sqldb/.
generate:
	sqlc generate

# Run the HTTP server (cobra subcommand to be wired).
run:
	go run . serve

# Apply the curated seed data (categories, activity types, achievements).
# Idempotent — re-running updates by slug.
seed:
	go run . seed

clean:
	rm -f kidsboard kidsboard.db kidsboard.db-shm kidsboard.db-wal

# Build the container image for the host's architecture. Fast loop for local
# dev. Override VERSION on the command line: `make docker VERSION=0.1.0`.
docker:
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE) .

# Bootstrap a buildx builder with QEMU emulators registered. One-time setup;
# safe to re-run. Required before `make docker-pi` on an amd64 host.
buildx-bootstrap:
	docker run --privileged --rm tonistiigi/binfmt --install all
	docker buildx create --name kidsboard-builder --use 2>/dev/null || docker buildx use kidsboard-builder

# Build for Raspberry Pi 4/5 (arm64 64-bit Raspberry Pi OS) and load the image
# into the local Docker daemon. Use this when you want to `docker save` the
# image and `scp` it to the Pi, or push to a private registry the Pi can pull
# from. Requires `make buildx-bootstrap` once on amd64 hosts.
docker-pi:
	docker buildx build \
		--platform=linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		-t $(IMAGE) \
		--load \
		.

# Build a multi-arch image (amd64 + arm64 by default) and push to a registry.
# Set REGISTRY and VERSION on the command line. PLATFORMS can be overridden
# to add linux/arm/v7 for Pi Zero 2 / Pi 3 on 32-bit OS.
docker-multiarch:
	docker buildx build \
		--platform=$(PLATFORMS) \
		--build-arg VERSION=$(VERSION) \
		-t $(IMAGE) \
		--push \
		.

# Lint the Helm chart for syntax + best-practice issues.
helm-lint:
	helm lint deploy/helm/kidsboard

# Render the chart to stdout. Useful for sanity-checking value overrides.
helm-template:
	helm template kidsboard deploy/helm/kidsboard
