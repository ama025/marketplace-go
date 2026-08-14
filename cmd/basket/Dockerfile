# ─── Stage 1: Build ───────────────────────────────────────────────────────────
# Используем официальный образ Go для сборки бинарника.
# alpine — минимальный образ, уменьшает размер слоя сборки.
FROM golang:1.26-alpine AS build

WORKDIR /src

# Копируем файлы зависимостей отдельно — Docker кэширует этот слой.
# go mod download запускается заново только если go.mod или go.sum изменились.
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код проекта
COPY . .

# Собираем статический бинарник basket-сервиса.
# CGO_ENABLED=0 — отключаем CGO для полностью статической сборки (нет зависимостей от libc).
RUN CGO_ENABLED=0 go build -o /app/basket ./cmd/basket

# ─── Stage 2: Final ───────────────────────────────────────────────────────────
# Минимальный образ для запуска — только alpine без Go-тулчейна.
# Итоговый образ будет ~15MB вместо ~800MB с golang:alpine.
FROM alpine:3.20 AS final

WORKDIR /app

# Копируем только скомпилированный бинарник из стадии build
COPY --from=build /app/basket .

# Копируем SQL-миграции (они нужны в рантайме — migrate.Up() читает их из папки)
COPY migrations/basket ./migrations/basket

# Внутренний порт сервиса (совпадает с BASKET_APP_PORT в compose-prod.yaml)
EXPOSE 7002

ENTRYPOINT ["./basket"]
