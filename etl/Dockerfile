FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o etl ./cmd/etl

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/etl .

COPY --from=builder /app/configs ./configs

RUN echo "0 1 * * * /root/etl -after yesterday" > /etc/crontabs/root

CMD ["crond", "-f"]
