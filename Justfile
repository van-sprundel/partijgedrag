default:
    @just --list

# One-time setup: git hooks and Go dependencies
install:
    lefthook install
    go mod download

# Start the database, apply migrations, and serve web + API with hot reload:
# templates and static files reload from disk per request (DEV=1), and the
# server restarts when Go files change (wgo).
dev:
    podman compose up -d --wait
    go run ./cmd/partijgedrag migrate
    DEV=1 go run github.com/bokwoon95/wgo@latest run ./cmd/partijgedrag serve
