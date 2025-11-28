# Partijgedrag

Partijgedrag is a web application that provides insight into the voting behavior of political parties in the Dutch parliament. It was originally created by Elwin Oost and later rebuilt in TypeScript.

## Project Structure

This repository is a monorepo containing three main components:

- `app/`: The web application, split into a frontend and a backend.
  - `frontend/`: A React/TypeScript frontend built with Vite.
  - `backend/`: A Node.js/TypeScript backend using Express, node-pg and SafeQL.
- `etl/`: A Go application responsible for extracting, transforming, and loading the voting data into the database.
- `docker-compose.yml`: Defines the services for the application, which for now is only the PostgreSQL database.

## Screenshots

<details>

<img width="2560" height="1330" alt="image" src="https://github.com/user-attachments/assets/acd42a7a-7705-47e3-9b50-2918aeabb3d1" />
<img width="2545" height="1330" alt="image" src="https://github.com/user-attachments/assets/bc2f1df7-dbe0-41a8-b8c1-5f784fe1aa66" />
<img width="2560" height="1330" alt="image" src="https://github.com/user-attachments/assets/360702fd-eaed-4c7e-aed9-51e78e547e19" />
<img width="2545" height="1330" alt="image" src="https://github.com/user-attachments/assets/85a47ef2-f668-4652-ae36-e11560bda8af" />
</details>

## Development Setup

### Prerequisites

- [Docker](https://www.docker.com/) and Docker Compose
- [Go](https://go.dev/) (version 1.21 or higher)
- [Node.js](https://nodejs.org/) (version 18 or higher)

### 1. Start the Database

The application requires a PostgreSQL database. You can start one using Docker Compose:

```bash
docker compose up -d
```

or Podman

```bash
podman compose up -d
```

This will start a PostgreSQL server and expose it on port 5432.

### 2. Load the Data (ETL)

The `etl` service fetches and processes parliamentary data.

1.  Navigate to the ETL directory:
    ```bash
    cd etl
    ```
2.  On the first run, seed the database:
    ```bash
    go run cmd/manage_categories/main.go --action=seed
    ```
3.  Run the ETL process:
    ```bash
    go run cmd/etl/main.go
    ```

### 3. Run the Application

The `app` is split into a backend and a frontend, which are separately for development.

### App

1.  Navigate to the app directory:
    ```bash
    cd app
    ```
2.  Install dependencies:
    ```bash
    npm install
    ```
3.  Set up your environment variables:
    ```bash
    cd backend; cp .env.example .env; cd ../frontend; cp .env.example .env; cd ..
    ```

> Ensure the `DATABASE_URL` in the new `.env` file is correctly configured for your environment.

4.  Start the app server:

    ```bash
    cd -; npm run dev
    ```

    The backend will be running on `http://localhost:3001`.

    The frontend will be accessible at `http://localhost:3000`.

## Acknowledgements

This project uses open data provided by the Tweede Kamer der Staten-Generaal (Dutch House of Representatives). For more information about the data sources and API documentation, visit https://opendata.tweedekamer.nl.
