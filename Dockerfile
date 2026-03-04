FROM golang:1.25-bookworm AS builder

RUN apt-get update && apt-get install -y gcc musl-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /api cmd/api/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

RUN groupadd -r appuser && useradd -r -g appuser appuser

WORKDIR /app

COPY --from=builder /api .
COPY migrations/ ./migrations/

RUN mkdir -p /app/data && chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/app/api", "-health"] || exit 1

ENTRYPOINT ["/app/api"]
