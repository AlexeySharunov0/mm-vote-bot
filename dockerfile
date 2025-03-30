# Используем официальный образ Go
FROM golang:1.21-bookworm AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем зависимости
COPY app/go.mod app/go.sum ./
RUN go mod download

# Копируем исходный код
COPY app/ .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Используем минимальный образ для запуска приложения
FROM alpine:latest

# Копируем собранный бинарник
WORKDIR /root/
COPY --from=builder /app/main .

# Запускаем приложение
CMD ["./main"]