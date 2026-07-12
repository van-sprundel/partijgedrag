FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -o /partijgedrag ./cmd/partijgedrag

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /partijgedrag /usr/bin/partijgedrag
EXPOSE 3001
ENTRYPOINT ["/usr/bin/partijgedrag"]
CMD ["serve"]
