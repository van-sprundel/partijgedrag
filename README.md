# Install

## ETL

### Sync

This will also setup the GORM DB schema

```sh
# Use any period you want
go run cmd/etl/main.go -after this-week
```

### Seed Categories

```sh
go run cmd/manage_categories/main.go --action=seed
```

### Enrich motions with categories

```sh
go run cmd/manage_categories/main.go --action=enrich
```

## Web App

Setup .env both in Frontend and Backend

### Backend

```sh
npm run dev
```

### Frontend

```sh
npm run dev
```
