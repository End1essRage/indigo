# Сборка приложения
FROM golang:alpine AS builder

WORKDIR /app

# Копируем исходники
COPY core ./core

# Компилируем приложение
RUN cd core/cmd && \
    CGO_ENABLED=0 GOOS=linux go build -o main main.go

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /app/core/cmd/main .

RUN mkdir -p /app/config /app/scripts

# Указываем точки монтирования
VOLUME [ "/app/config", "/app/scripts" ]

# Запускаем приложение
CMD [ "./main" ]