# Используем официальный образ Go в качестве базового
FROM golang:1.23-alpine AS builder

# Установим необходимые зависимости
RUN apk add --no-cache git

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum для кэширования зависимостей
COPY ../../go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем остальные файлы приложения
COPY ../.. .

# Сборка приложения
RUN go build -o main ./cmd/wallet-app

# Минимальный образ для запуска
FROM alpine:latest

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем конфигурационные файлы
COPY --from=builder /app/internal/config /app/internal/config
COPY --from=builder /app/main .
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/migrations /app/migrations
COPY ../../.env .env

# Экспонируем порт
EXPOSE 8080
