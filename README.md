# Partijgedrag

Partijgedrag is a web application that provides insight into the voting behavior of political parties in the Dutch parliament. It was originally created by Elwin Oost, later rebuilt in TypeScript, and now runs as a single Go binary.

## Project Structure

The whole application — data ingestion, web pages, and JSON API — is one Go program:

- `cmd/partijgedrag/`: The CLI entry point with `migrate`, `ingest`, `sync`, `status`, `maintenance`, `inspect`, and `serve` subcommands.
- `internal/`: Ingestion pipelines (Tweede Kamer OData), analysis queries, motion categorization, and the server-rendered web UI.
- `deploy/systemd/`: Unit files for running the server and a recurring sync on a plain Linux host.
- `docker-compose.yml`: The PostgreSQL database for local development.

## Screenshots

<details>

<img width="2560" height="1330" alt="image" src="https://github.com/user-attachments/assets/acd42a7a-7705-47e3-9b50-2918aeabb3d1" />
<img width="2545" height="1330" alt="image" src="https://github.com/user-attachments/assets/bc2f1df7-dbe0-41a8-b8c1-5f784fe1aa66" />
<img width="2560" height="1330" alt="image" src="https://github.com/user-attachments/assets/360702fd-eaed-4c7e-aed9-51e78e547e19" />
<img width="2545" height="1330" alt="image" src="https://github.com/user-attachments/assets/85a47ef2-f668-4652-ae36-e11560bda8af" />
</details>

## Development Setup

### Prerequisites

- [Go](https://go.dev/) (see `go.mod` for the version)
- [Podman](https://podman.io/) or Docker, with compose
- [just](https://github.com/casey/just) and [lefthook](https://lefthook.dev/) for the development workflow

### Quick start

```bash
just install   # git hooks + Go dependencies
just dev       # start the database, apply migrations, serve web + API
```

The server runs on `http://localhost:3001`. Configuration is read from environment variables; see `.env.example` for the defaults.

### Loading data

The database starts empty. Fetch parties, motions, and votes from the Tweede Kamer open data API, and categorize motions, with:

```bash
go run ./cmd/partijgedrag sync tweedekamer
```

The first full sync takes a while; rerunning it is incremental. `deploy/systemd/` contains a timer unit that keeps the data fresh on a server. See `go run ./cmd/partijgedrag` for all commands, including ingestion status and data-quality tooling.

## Acknowledgements

This project uses open data provided by the Tweede Kamer der Staten-Generaal (Dutch House of Representatives). For more information about the data sources and API documentation, visit https://opendata.tweedekamer.nl.
