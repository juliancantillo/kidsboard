.PHONY: check test build generate run seed clean docker helm-lint helm-template

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

# Build the container image. Pass VERSION=x.y.z to stamp the binary.
docker:
	docker build --build-arg VERSION=$${VERSION:-dev} -t kidsboard:$${VERSION:-dev} .

# Lint the Helm chart for syntax + best-practice issues.
helm-lint:
	helm lint deploy/helm/kidsboard

# Render the chart to stdout. Useful for sanity-checking value overrides.
helm-template:
	helm template kidsboard deploy/helm/kidsboard
