<img width="2880" height="1800" alt="01-home" src="https://github.com/user-attachments/assets/53d1c0c8-c210-4414-a367-f862dbf90f8e" /># Partijgedrag

Partijgedrag is a web application that provides insight into the voting behavior of political parties in the Dutch parliament. It was originally created by Elwin Oost, later rebuilt in TypeScript, and has now become a final implementation that runs as a single Go binary.

## Project Structure

- `cmd/partijgedrag/`: The CLI entry point with `migrate`, `ingest`, `sync`, `status`, `maintenance`, `inspect`, and `serve` subcommands.
- `internal/`: Ingestion pipelines (Tweede Kamer OData), analysis queries, motion categorization, and the server-rendered web UI.
- `deploy/systemd/`: Unit files for running the server and a recurring sync on a plain Linux host.
- `docker-compose.yml`: The PostgreSQL database for local development.

## Screenshots

<details>
<summary>Home, moties, stemwijzer, en de analysepagina's</summary>

Recente stemmingen, met de totalen over de hele dataset.

<img alt="Home" src="https://github.com/user-attachments/assets/4dac2ba5-8573-4b4c-a14b-680ff48ff184" />

Alle moties, filterbaar op onderwerp en kabinetsperiode.

<img alt="Moties" src="https://github.com/user-attachments/assets/9f32ffe2-3562-4244-b402-93c249fb3f0e" />

De stemuitslag van een motie per partij, met zetels en de besluiten eromheen.

<img alt="Motie" src="https://github.com/user-attachments/assets/2eb481e8-ad26-43a6-afe8-273b14e77f3b" />

De stemwijzer begint bij een profiel: periode, onderwerpen, en partijen.

<img alt="Stemwijzer instellingen" src="https://github.com/user-attachments/assets/ef959319-228b-455d-87d4-e2dee30cc1f9" />

Daarna zijn de stellingen echte moties, met de letterlijke verzoekt-tekst.

<img alt="Stemwijzer stellingen" src="https://github.com/user-attachments/assets/c8cc20d1-ee6f-4721-b855-460bc4620711" />

De uitslag krijgt een vast adres en is deelbaar.

<img alt="Stemwijzer resultaat" src="https://github.com/user-attachments/assets/eb6643c6-68cd-42a4-ae80-dedda2044e46" />

Hoe vaak partijen hetzelfde stemmen.

<img alt="Partijgelijkenis" src="https://github.com/user-attachments/assets/1d98b9b1-77a0-48af-9b09-e9dcb4f58830" />

Eén partij uitgelicht, per onderwerp.

<img alt="Partijfocus" src="https://github.com/user-attachments/assets/ea552501-d64c-4cba-a270-5f9c641560a1" />

Stemgedrag van de coalitie tegenover de oppositie.

<img alt="Coalitie" src="https://github.com/user-attachments/assets/4bc6b612-8fe5-41a0-be96-71a3ad2e9554" />

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

The first full sync takes a while; rerunning it is incremental. See `go run ./cmd/partijgedrag` for all commands, including ingestion status and data-quality tooling.

## Deployment

The GitHub CI workflow builds a single container image; running it with `serve` (the default command) is all a server needs. On startup the server applies pending migrations and starts a built-in sync scheduler, so the data stays fresh without an external cron:

- `SYNC_INTERVAL` (default `1h`): how often `serve` runs a full `sync tweedekamer`. The first run starts one minute after boot. Set to `0` to disable, e.g. when scheduling sync externally instead (`deploy/systemd/` has a timer unit for that setup).
- `SYNC_MOTION_VOTE_LIMIT` (default `250`) and `SYNC_MOTION_DOCUMENT_LIMIT` (default `500`): how many motions get votes/documents backfilled per run. Pipeline advisory locks make concurrent syncs safe: an overlapping run fails fast.
- `SYNC_MOTION_VOTE_RESYNC_GRACE` (default `720h`): re-polls votes for motions with no terminating decision, and for decided ones until this long after that decision. A motion is normally ingested before it is voted on, so without this it keeps the zero votes it had on first sight. The window also covers late amendments; a `vergissing` is usually filed about a week after the vote. Set to `0` to sync a motion's votes only once.
- `SYNC_MOTION_DOCUMENT_RESYNC_GRACE` (default `2160h`): retries motions that still have no bullet points, until this long after they were proposed. A published document never changes, so a successful extraction is never fetched again. Past the window a motion counts as permanently without a document, which bounds the retry set. Set to `0` to disable.

## Acknowledgements

This project uses open data provided by the Tweede Kamer der Staten-Generaal (Dutch House of Representatives). For more information about the data sources and API documentation, visit https://opendata.tweedekamer.nl.

