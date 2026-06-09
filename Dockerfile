# 1. Этап сборки
FROM golang:1.25.0-bookworm AS builder

WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Компилируем приложение (на выходе будет бинарник main)
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 2. Этап запуска
FROM alpine:latest

WORKDIR /app

# Копируем скомпилированный файл из первого этапа
COPY --from=builder /app/main .

# Открываем порт, на котором слушает ваше Go web-приложение
EXPOSE 8080

# Команда при запуске контейнера
CMD ["./main"]