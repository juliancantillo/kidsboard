.PHONY: check test build css css-clean fmt generate run seed clean docker docker-pi docker-multiarch buildx-bootstrap helm-lint helm-template

# Override on the command line: `make docker-pi REGISTRY=ghcr.io/me VERSION=0.1.0`
REGISTRY  ?= kidsboard
VERSION   ?= dev
IMAGE     ?= $(REGISTRY):$(VERSION)
PLATFORMS ?= linux/amd64,linux/arm64

# Tailwind CSS standalone binary. Pinned for reproducibility; bump together.
TAILWIND_VERSION ?= v3.4.17
TAILWIND_BIN     ?= bin/tailwindcss
CSS_INPUT        := input.css
CSS_OUTPUT       := internal/view/static/css/kidsboard.css

# Default target: build CSS, then run tests + build. Anything green here is
# shippable.
check: css test build

test:
	go test ./...

# Build embeds the CSS via the existing //go:embed in internal/view/static.go,
# so the stylesheet must exist on disk before `go build`.
build: css
	go build ./...

# Compile the Tailwind stylesheet from input.css to the embedded static path.
# Falls back to downloading the standalone CLI when `tailwindcss` isn't on the
# PATH — no Node required.
css: $(CSS_OUTPUT)

$(CSS_OUTPUT): $(CSS_INPUT) tailwind.config.js $(shell find internal/view/templates -name '*.html')
	@mkdir -p $(dir $(CSS_OUTPUT))
	@if command -v tailwindcss >/dev/null 2>&1; then \
		echo "tailwindcss -> $(CSS_OUTPUT)"; \
		tailwindcss -i $(CSS_INPUT) -o $(CSS_OUTPUT) --minify; \
	else \
		$(MAKE) $(TAILWIND_BIN); \
		echo "$(TAILWIND_BIN) -> $(CSS_OUTPUT)"; \
		$(TAILWIND_BIN) -i $(CSS_INPUT) -o $(CSS_OUTPUT) --minify; \
	fi

# Download the standalone Tailwind CLI for the host's OS+arch on demand.
# Cached at bin/tailwindcss; delete it with `make css-clean`.
$(TAILWIND_BIN):
	@mkdir -p $(dir $(TAILWIND_BIN))
	@OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
	ARCH=$$(uname -m); \
	case "$$OS" in \
	  darwin)  TWOS=macos ;; \
	  linux)   TWOS=linux ;; \
	  *) echo "unsupported OS: $$OS" >&2; exit 1 ;; \
	esac; \
	case "$$ARCH" in \
	  x86_64|amd64) TWARCH=x64 ;; \
	  aarch64|arm64) TWARCH=arm64 ;; \
	  *) echo "unsupported arch: $$ARCH" >&2; exit 1 ;; \
	esac; \
	URL="https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-$$TWOS-$$TWARCH"; \
	echo "downloading $$URL"; \
	curl -fsSL -o $(TAILWIND_BIN) "$$URL"; \
	chmod +x $(TAILWIND_BIN)

css-clean:
	rm -f $(CSS_OUTPUT) $(TAILWIND_BIN)

fmt:
	gofmt -w .

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
