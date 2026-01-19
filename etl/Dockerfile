FROM golang:alpine AS builder
WORKDIR /app

ARG DATABASE_URL
ENV DATABASE_URL=${DATABASE_URL}

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY configs/ ./configs/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/etl ./cmd/etl/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/manage_categories ./cmd/manage_categories/main.go

FROM alpine:latest
LABEL org.opencontainers.image.source=https://github.com/van-sprundel/partijgedrag
WORKDIR /app

# Install cron
RUN apk add --no-cache busybox-suid

COPY --from=builder /app/etl .
COPY --from=builder /app/manage_categories .
COPY --from=builder /app/configs ./configs
COPY crontab /etc/crontabs/root

# Give execution rights on the cron job
RUN chmod 0644 /etc/crontabs/root

# Create log file for cron jobs
RUN touch /var/log/etl.log

# Run the command on container startup
CMD ["crond", "-f"]