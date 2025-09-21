# Partijgedrag

Partijgedrag is a web application that provides insight into the voting behavior of political parties in the Dutch parliament. It was originally created by Elwin Oost and later rebuilt in TypeScript.

## Project Structure

This repository is a monorepo containing three main components:

- `app/`: The web application, split into a frontend and a backend.
  - `frontend/`: A React/TypeScript frontend built with Vite.
  - `backend/`: A Node.js/TypeScript backend using Express and Prisma.
- `etl/`: A Go application responsible for extracting, transforming, and loading the voting data into the database.
- `docker-compose.yml`: Defines the services for the application, including the PostgreSQL database.

## Development Setup

### Prerequisites

- [Docker](https://www.docker.com/) and Docker Compose
- [Go](https://go.dev/) (version 1.21 or higher)
- [Node.js](https://nodejs.org/) (version 18 or higher)

### 1. Start the Database

The application requires a PostgreSQL database. You can start one using Docker Compose:

```bash
docker-compose up -d
```

This will start a PostgreSQL server and expose it on port 5432.

### 2. Load the Data (ETL)

The `etl` service fetches and processes parliamentary data.

1.  Navigate to the ETL directory:
    ```bash
    cd etl
    ```
2.  Run the ETL process:
    ```bash
    go run cmd/etl/main.go
    ```
    This will populate the database with the necessary data. The database connection is configured in `etl/configs/config.yaml`.

### 3. Run the Application

The `app` is split into a backend and a frontend, which must be run separately for development.

#### Backend

1.  Navigate to the backend directory:
    ```bash
    cd app/backend
    ```
2.  Install dependencies:
    ```bash
    npm install
    ```
3.  Set up your environment variables:
    ```bash
    cp .env.example .env
    ```
    Ensure the `DATABASE_URL` in the new `.env` file is correctly configured for your environment.
4.  Start the backend development server:
    ```bash
    npm run dev
    ```
    The backend will be running on `http://localhost:3001`.

#### Frontend

1.  Navigate to the frontend directory:
    ```bash
    cd app/frontend
    ```
2.  Install dependencies:
    ```bash
    npm install
    ```
3.  Set up your environment variables:
    ```bash
    cp .env.example .env
    ```
    The default `VITE_API_URL` should point to the backend at `http://localhost:3001`.
4.  Start the frontend development server:
    ```bash
    npm run dev
    ```
    The frontend will be accessible at `http://localhost:3000`.
