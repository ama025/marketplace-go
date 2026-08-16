# marketplace

Учебный e-commerce бэкенд на Go, построенный как набор независимых микросервисов по принципам Clean Architecture / CQRS. Каждый сервис — отдельное приложение со своей базой данных, которое общается с остальными через REST, gRPC и очередь сообщений.

## Архитектура

| Сервис | Протокол | Хранилище | Назначение |
|---|---|---|---|
| **catalog** | REST (Gin) | PostgreSQL | Товары, бренды, категории |
| **basket** | REST (Gin) | Redis (кэш) + PostgreSQL | Корзина покупателя, оформление корзины |
| **checkout** | REST (Gin) | PostgreSQL | Заказы; создаются асинхронно по событию из RabbitMQ |
| **promotion** | gRPC | MySQL | Скидки/акции на товары |

Basket и Checkout связаны через **RabbitMQ**: при оформлении корзины (`checkout`) basket публикует событие `order confirmed`, которое забирает checkout и создаёт заказ.

Каждый сервис внутри разложен по слоям:

```
internal/<service>/
├── api/             # HTTP-хендлеры и роуты (или grpc/ для promotion)
├── application/     # command/query-обработчики (CQRS)
├── domain/          # сущности, репозитории (интерфейсы), value-объекты
└── infrastructure/  # реализация репозиториев, кэш, брокер сообщений
```

Общий код (middleware, метрики, шина сообщений) вынесен в `internal/shared`.

## Технологии

- Go 1.26, [Gin](https://github.com/gin-gonic/gin) — REST-сервисы
- gRPC + Protocol Buffers — сервис promotion
- PostgreSQL, MySQL, Redis
- RabbitMQ — асинхронная интеграция basket → checkout
- golang-migrate — миграции БД
- Prometheus + Grafana — метрики и дашборды
- k6 — нагрузочное тестирование
- Docker / Docker Compose

## Структура репозитория

```
cmd/<service>/        # main.go и Dockerfile каждого сервиса
internal/              # исходный код сервисов (см. выше)
migrations/<service>/  # SQL-миграции по сервисам
infrastructure/
├── prometheus/         # конфиг Prometheus
├── grafana/             # дашборды и provisioning
└── k6/                  # скрипты нагрузочного тестирования
compose-dev.yaml       # инфраструктура для локальной разработки
compose-prod.yaml      # сборка и запуск всех сервисов в проде
```

## Быстрый старт (dev)

Сервисы БД/кэша/очереди/мониторинга поднимаются в Docker, сами Go-сервисы (catalog/basket/checkout) запускаются локально через `go run`; promotion собирается и запускается в контейнере.

1. Поднять инфраструктуру:

   ```bash
   docker compose -p marketplace-dev -f compose-dev.yaml up -d
   ```

2. Создать `.env` в корне проекта со своими переменными окружения (хосты/порты берутся из `compose-dev.yaml`, пароли — `12345678`). Пример для одного сервиса:

   ```env
   CATALOG_APP_PORT=9001
   CATALOG_PG_HOST=localhost
   CATALOG_PG_PORT=9101
   CATALOG_PG_DATABASE=catalog-db-dev
   CATALOG_PG_USER=postgres
   CATALOG_PG_PASSWORD=12345678
   CATALOG_PG_SSLMODE=disable
   ```

   По аналогии добавьте переменные `BASKET_*`, `CHECKOUT_*` (см. `cmd/*/main.go` — там перечислены все читаемые переменные окружения).

3. Запустить нужный сервис:

   ```bash
   go run ./cmd/catalog
   go run ./cmd/basket
   go run ./cmd/checkout
   ```

   При старте каждый сервис сам накатывает миграции из `migrations/<service>`.

4. Promotion (gRPC) уже поднимается вместе с инфраструктурой (`promotion_grpc_dev` в compose-dev.yaml).

## Продакшн

```bash
docker compose -f compose-prod.yaml up -d --build
```

Все сервисы собираются из своих `Dockerfile` и запускаются вместе с базами данных, RabbitMQ и мониторингом.

## Мониторинг

- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (admin/admin)

Каждый сервис отдаёт метрики на `/metrics`.

## Нагрузочное тестирование

```bash
docker compose -p marketplace-dev -f compose-dev.yaml run k6_catalog
docker compose -p marketplace-dev -f compose-dev.yaml run k6_basket
docker compose -p marketplace-dev -f compose-dev.yaml run k6_checkout
```

## API

### Catalog — `/api/v1`
- `GET /brands`, `GET /categories`
- `GET /catalog-items`, `GET /catalog-items/:id`, `GET /catalog-items/title/:title`, `GET /catalog-items/brand/:brand`
- `POST /catalog-items`, `PUT /catalog-items`, `DELETE /catalog-items/:id`

### Basket — `/api/v1`
- `POST /cart`, `GET /cart/:accountName`, `DELETE /cart/:accountName`
- `POST /cart/:accountName/checkout`

### Checkout — `/api/v1/orders`
- `POST /orders`, `GET /orders/:id`, `GET /orders`

### Promotion — gRPC (`cmd/proto/greet.proto`)
- `GetDiscount`, `GetDiscounts`, `AddDiscount`, `DeactivateDiscount`
