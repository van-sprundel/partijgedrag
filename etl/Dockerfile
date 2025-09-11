FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/etl ./cmd/etl/main.go

FROM alpine:latest
WORKDIR /app

# Install cron
RUN apk add --no-cache busybox-suid

COPY --from=builder /app/etl .
COPY configs/config.yaml ./configs/config.yaml
COPY crontab /etc/crontabs/root

# Give execution rights on the cron job
RUN chmod 0644 /etc/crontabs/root

# Run the command on container startup
CMD ["crond", "-f"]
