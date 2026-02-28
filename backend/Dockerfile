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

CMD ["./api"]
