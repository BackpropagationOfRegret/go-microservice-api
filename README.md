# Food Delivery Platform

Учебная платформа заказа и доставки еды в стиле Delivery Hero / Uber Eats. Система построена как набор Go-микросервисов с **database per service**, **Kafka-хореографией** саги заказа и веб-клиентом на **Next.js**.

> Это MVP для изучения микросервисной архитектуры, а не production-ready решение.

## Что внутри

| Слой | Технологии |
|------|------------|
| Backend | Go 1.22, HTTP/JSON, `segmentio/kafka-go` |
| Frontend | Next.js, MUI, TypeScript |
| Данные | PostgreSQL 16 (отдельная БД на сервис) |
| События | Apache Kafka 3.8 |
| Наблюдаемость | OpenTelemetry → Jaeger |
| Инфраструктура | Docker Compose, Traefik |

**7 backend-сервисов:** `user`, `restaurant`, `order`, `payment`, `delivery`, `notification`, `gateway` (BFF).

Клиент ходит только в **API Gateway** (`:8090`). Сервисы общаются синхронно по HTTP и асинхронно через Kafka. Жизненный цикл заказа — сага на событиях: `order.created` → оплата → кухня → доставка.

Подробнее: [MD/Architecture.md](./MD/Architecture.md) (диаграммы, поток заказа, откаты, ограничения).

## Быстрый старт

**Требования:** Docker, Docker Compose, Make.

```bash
git clone <repo-url>
cd go-microservice

make up          # поднимает всё (ENV=dev по умолчанию)
make docs        # список URL
make demo        # curl: регистрация + список ресторанов
```

Откройте [http://localhost:3000](http://localhost:3000) — фронтенд. API и документация: [http://localhost:8090/docs](http://localhost:8090/docs).

Остановка:

```bash
make down        # контейнеры остановлены, данные PostgreSQL сохранены
make down-clean  # полный сброс volumes
```

## Полезные URL

| URL | Назначение |
|-----|------------|
| http://localhost:3000 | Frontend |
| http://localhost:8090 | API Gateway |
| http://localhost:8090/docs | Scalar (OpenAPI) |
| http://localhost:8088 | Traefik dashboard |
| http://localhost:16686 | Jaeger UI |

## Make-команды

| Команда | Описание |
|---------|----------|
| `make up ENV=dev\|stage\|prod` | Запуск стека (`deploy/env/$ENV.env`) |
| `make down` | Остановить контейнеры |
| `make down-clean` | Остановить и удалить volumes |
| `make build` | Собрать образы |
| `make logs` | Логи всех сервисов |
| `make demo` | Демо через curl |
| `make docs` | URL сервисов |
| `make test` | `go test` во всех модулях |
| `make tidy` | `go mod tidy` во всех модулях |
| `make frontend-dev` | Next.js локально (API на `:8090`) |

## Структура репозитория

```text
pkg/                  # общие пакеты: config, events, kafka, auth, telemetry
services/
  user/               # пользователи, JWT
  restaurant/         # рестораны, меню, кухня
  order/              # заказы, владелец саги
  payment/            # платежи и возвраты
  delivery/           # курьеры, доставка
  notification/       # уведомления (пока лог)
  gateway/            # BFF, прокси, OpenAPI
frontend/             # Next.js клиент
deploy/               # init-db, kafka-init, traefik, env-файлы
docker-compose.yml
MD/                   # архитектура, идея, трассировка
```

Go workspace: `go.work` связывает все модули.

## Локальная разработка

### Backend (отдельный сервис)

```bash
# PostgreSQL и Kafka должны быть доступны (например, через make up)
cd services/order
go run ./cmd
```

Переменные окружения — см. `deploy/env/dev.env` и `.env.example`.

### Frontend

```bash
make up              # gateway на :8090
make frontend-dev    # Next.js на :3000, API_URL=http://localhost:8090
```

### Тесты

```bash
make test
make tidy
```

## Конфигурация

| Файл | Назначение |
|------|------------|
| `deploy/env/dev.env` | Локальная разработка |
| `deploy/env/stage.env.example` | Шаблон staging |
| `deploy/env/prod.env.example` | Шаблон production |
| `.env.example` | Быстрый старт |

Ключевые переменные: `JWT_SECRET`, `POSTGRES_*`, `KAFKA_BROKERS`, `*_SVC_URL`, `PAYMENT_FAIL_RATE`, `KITCHEN_FAIL_RATE`, `SEED_DATA`.

Симуляции задержек: `KITCHEN_COOK_DURATION`, `DELIVERY_DURATION`, `PAYMENT_PROCESS_LATENCY`.

## Документация

| Документ | Содержание |
|----------|------------|
| [MD/Architecture.md](./MD/Architecture.md) | Архитектура, диаграммы, сага, транзакции |
| [MD/Idea.md](./MD/Idea.md) | Изначальная идея и доменная модель |
| [MD/Tracing.md](./MD/Tracing.md) | Распределённая трассировка |
| [http://localhost:8090/docs](http://localhost:8090/docs) | OpenAPI / Scalar (после `make up`) |

## Карта сервисов

| Сервис | Порт | БД |
|--------|------|-----|
| gateway | 8090 | — |
| user-svc | 8081 | `user_db` |
| restaurant-svc | 8082 | `restaurant_db` |
| order-svc | 8083 | `order_db` |
| payment-svc | 8084 | `payment_db` |
| delivery-svc | 8085 | `delivery_db` |
| notification-svc | 8086 | — |
| frontend | 3000 | — |

## Ограничения MVP

- Хореография без центрального оркестратора
- Нет transactional outbox / inbox
- HTTP между сервисами (не gRPC)
- `notification-svc` только логирует события
- Время готовки и доставки — симуляция через env

Полный список и направления развития — в [MD/Architecture.md](./MD/Architecture.md).
