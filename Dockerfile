# Используем официальный образ Go для сборки
FROM golang:1.24.1-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем весь проект в контейнер
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o bookshop ./cmd/server

# Используем минимальный образ Alpine для финального контейнера
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем собранное приложение из builder
COPY --from=builder /app/bookshop .

# Копируем миграции
COPY migrations ./migrations

# Указываем порт, который будет использовать приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./bookshop"]