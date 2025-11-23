# ---------- Build stage ----------
FROM golang:1.25 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем бинарник
RUN go build -o server ./cmd/server

# ---------- Runtime stage ----------
FROM debian:12-slim

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

# Копируем бинарник
COPY --from=build /app/server .

# ❗ Копируем миграции (ОБЯЗАТЕЛЬНО!)
COPY --from=build /app/internal/repository/migrations ./internal/repository/migrations

EXPOSE 8080

CMD ["./server"]
