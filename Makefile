.PHONY: check test build generate run seed clean

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
