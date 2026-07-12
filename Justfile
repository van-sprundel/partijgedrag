default:
    @just --list

# One-time setup: git hooks and Go dependencies
install:
    lefthook install
    go mod download

# Start the database, apply migrations, and serve web + API
dev:
    podman compose up -d --wait
    go run ./cmd/partijgedrag migrate
    go run ./cmd/partijgedrag serve
