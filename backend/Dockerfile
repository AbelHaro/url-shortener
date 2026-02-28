FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /api ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /app

COPY --from=builder /api .

RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

LABEL "traefik.enable=true" \
      "traefik.http.routers.backend.rule=Host('api.url-shortener.abelharo.me')" \
      "traefik.http.routers.backend.entrypoints=websecure" \
      "traefik.http.routers.backend.tls.certresolver=le" \
      "traefik.http.services.backend.loadbalancer.server.port=8080"

CMD ["./api"]
