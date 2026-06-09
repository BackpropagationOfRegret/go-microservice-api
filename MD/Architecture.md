# Архитектура платформы доставки еды

Простое описание того, **зачем** проект устроен как микросервисы, **что** между ними используется и **как** проходит заказ — с событиями, транзакциями и откатами.

Связанные документы: [Idea.md](./Idea.md) (изначальная идея), [Tracing.md](./Tracing.md) (распределённая трассировка), OpenAPI и Scalar на `http://localhost:8090/docs`.

---

## Зачем микросервисы в этом домене

Платформа доставки еды — типичный пример, где один монолит быстро становится тесным:

- **Разная нагрузка.** В пятницу вечером растут заказы и платежи, а каталог ресторанов почти не меняется.
- **Разные команды.** Платежи, логистика и меню — разные зоны ответственности.
- **Разная надёжность.** Падение SMS-уведомлений не должно ломать приём заказа.
- **Много асинхронных шагов.** Оплата, кухня, курьер, письма — это не один HTTP-запрос.

Мы разбили систему на **7 backend-сервисов + gateway + фронтенд**. У каждого сервиса своя база данных и своя зона ответственности.

---

## Карта сервисов

| Сервис | Порт | За что отвечает | Своя БД |
|--------|------|-----------------|---------|
| **api-gateway** | 8090 | Единая точка входа для клиентов, JWT, агрегация (заказ + оплата) | — |
| **user-svc** | 8081 | Регистрация, логин, JWT, профиль, адреса | `user_db` |
| **restaurant-svc** | 8082 | Рестораны, меню, «кухня» (готовность заказа) | `restaurant_db` |
| **order-svc** | 8083 | Создание заказа, статусы, владелец жизненного цикла заказа | `order_db` |
| **payment-svc** | 8084 | Списание денег (симуляция), запись платежа | `payment_db` |
| **delivery-svc** | 8085 | Курьеры, назначение, доставка | `delivery_db` |
| **notification-svc** | 8086 | Push / SMS / email (пока только лог) | — |
| **frontend** | 3000 | Веб-интерфейс (Next.js + MUI) | — |

Инфраструктура в Docker Compose:

| Компонент | Зачем |
|-----------|--------|
| **PostgreSQL** | Отдельная база на сервис (database per service) |
| **Kafka** | Асинхронные события между сервисами |
| **Jaeger** | UI трассировки (OTLP gRPC `:4317`, UI `:16686`) — см. [Tracing.md](./Tracing.md) |
| **Traefik** | Reverse proxy на порт 80 (опциональный вход) |

---

## Два способа связи между сервисами

### 1. Синхронно — HTTP + JSON

Запрос → сразу ответ. Используется, когда **нужен ответ прямо сейчас**.

```
Браузер ──HTTP──► api-gateway ──HTTP──► user-svc / restaurant-svc / order-svc / …
```

**Где именно:**

| Кто | Кому | Зачем |
|-----|------|--------|
| Frontend | api-gateway | Все `/api/*` |
| api-gateway | backend-сервисы | Прокси и агрегация |
| order-svc | user-svc | `GET /internal/validate/{id}` — пользователь существует? |
| order-svc | restaurant-svc | `POST /internal/menu/validate` — блюда есть и ресторан открыт? |
| delivery-svc | order-svc | `GET /orders/{id}` — узнать `user_id` для уведомления |

**Почему не gRPC:** в идее он упоминался, в коде пока **HTTP** — проще отлаживать (curl, OpenAPI/Scalar). gRPC имел бы смысл при высокой нагрузке на внутренние вызовы.

**Авторизация клиента:** JWT (`Authorization: Bearer <token>`). Секрет `JWT_SECRET` общий у gateway и user-svc.

---

### 2. Асинхронно — Kafka (события)

Сервис **публикует факт** («заказ создан»), другие **сами решают**, что делать. Никто не ждёт ответа в том же HTTP-запросе.

Библиотека: `segmentio/kafka-go`. Общие типы событий: `pkg/events`, обёртка: `pkg/kafka`.

Формат сообщения:

```json
{
  "type": "order.created",
  "timestamp": "2026-06-09T12:00:00Z",
  "payload": { "order_id": "...", "total_amount": 42.97 }
}
```

**Топики Kafka:**

| Топик | Смысл |
|-------|--------|
| `order.created` | Заказ создан, пора платить |
| `payment.processed` | Оплата прошла |
| `payment.failed` | Оплата не прошла |
| `order.paid` | Заказ оплачен, пора готовить |
| `order.ready` | Еда готова, пора звать курьера |
| `courier.assigned` | Курьер назначен |
| `order.delivered` | Доставлено |
| `order.status.changed` | Смена статуса (служебный) |

Топики создаются скриптом `deploy/kafka-init.sh` при старте.

---

## Оркестрация или хореография?

Используется **хореография** (choreography).

- **Нет** центрального «дирижёра», который по цепочке вызывает payment → restaurant → delivery.
- **Есть** цепочка **событий**: каждый сервис подписан на нужные топики и реагирует сам.

`order-svc` — не оркестратор, а **владелец сущности «заказ»**: хранит статус и публикует/слушает события, но не командует другими сервисами по HTTP.

```
                    ┌─────────────┐
  order.created ───►│ payment-svc │
                    └──────┬──────┘
                           │ payment.processed
                           ▼
                    ┌─────────────┐
                    │  order-svc  │──► order.paid
                    └─────────────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
      restaurant-svc  notification  (логи)
              │
              │ order.ready
              ▼
       delivery-svc ──► courier.assigned ──► order.delivered
```

**Плюс:** слабая связность — добавили `notification-svc`, никого не переписывали.  
**Минус:** весь сценарий размазан по коду; сложнее отлаживать «где застрял заказ».

---

## Бизнес-сценарий: заказ от начала до конца

### Шаг 0 — клиент оформляет заказ (синхронно)

```
Frontend → POST /api/orders → api-gateway → order-svc
```

**Внутри order-svc (одна «локальная» логика, но несколько шагов):**

1. HTTP → **user-svc**: пользователь валиден?
2. HTTP → **restaurant-svc**: блюда есть, ресторан открыт, цены актуальны?
3. Считает сумму, сохраняет заказ в **свою** БД со статусом `PENDING`.
4. Публикует в Kafka: **`order.created`**.

Клиент сразу получает ответ с заказом в статусе `PENDING`. Дальше процесс идёт в фоне.

---

### Шаг 1 — оплата (асинхронно)

**Слушает:** `payment-svc` на топик `order.created`.

**Делает:**

1. Симулирует платёжный шлюз (~500 мс).
2. Пишет запись в `payment_db` (`SUCCESS` или `FAILED`).
3. Публикует:
   - **`payment.processed`** — если ок,
   - **`payment.failed`** — если нет (в демо можно задать `PAYMENT_FAIL_RATE`).

**Параллельно** `notification-svc` шлёт: *«Заказ принят»*.

---

### Шаг 2 — order-svc реагирует на оплату

| Событие | Действие order-svc |
|---------|-------------------|
| `payment.processed` | Статус → `PAID`, публикует **`order.paid`** |
| `payment.failed` | Статус → `CANCELLED` (отмена заказа) |

**notification-svc:** *«Оплата прошла»* или сообщение об ошибке.

Это единственный реализованный **откат (компенсация)** в саге: неудачная оплата → заказ отменяется.

---

### Шаг 3 — кухня (асинхронно)

**Слушает:** `restaurant-svc` на `order.paid`.

**Делает:**

1. Логирует «заказ на кухне».
2. Через ~3 секунды (симуляция готовки) публикует **`order.ready`**.

**notification-svc:** *«Заказ готов, ожидайте курьера»*.

---

### Шаг 4 — доставка (асинхронно)

**Слушает:** `delivery-svc` на `order.ready`.

**Делает:**

1. Находит свободного курьера в `delivery_db`, ставит ему статус `BUSY`.
2. HTTP → order-svc (узнать `user_id`).
3. Публикует **`courier.assigned`**.
4. `order-svc` обновляет заказ: курьер, статус `DELIVERING`.
5. Через ~5 секунд — **`order.delivered`**, курьер снова `AVAILABLE`.
6. `order-svc` ставит статус `DELIVERED`.

**notification-svc:** *«Курьер в пути»*, потом *«Приятного аппетита!»*.

---

### Статусы заказа

```
PENDING → PAID → DELIVERING → DELIVERED

PENDING → CANCELLED          (оплата не прошла)
PAID → CANCELLED → REFUNDED  (кухня / доставка не смогли выполнить)
```

Промежуточные `PREPARING` / `READY` в order-svc не выставляются — они живут в событиях кухни и доставки. Фронтенд опрашивает `GET /api/orders/{id}` каждые 3 секунды.

---

## Транзакции: что есть и чего нет

### Локальные транзакции (есть)

Каждый сервис пишет **только в свою** PostgreSQL-базу. Операции внутри одного сервиса атомарны на уровне БД (один `INSERT` / `UPDATE`).

Примеры:

- `order-svc`: создать заказ — одна запись в `order_db`.
- `payment-svc`: создать платёж — одна запись в `payment_db`.
- `delivery-svc`: занять курьера — `UPDATE` в `delivery_db`.

### Распределённых транзакций нет (и это нормально)

Нет **2PC**, нет единой транзакции «заказ + платёж + курьер» на все базы сразу.

В микросервисах так обычно и делают: вместо одной большой транзакции — **сага** из шагов и событий.

### Транзакция «запись в БД + публикация в Kafka»

Сейчас это **не** атомарно:

```text
order-svc: INSERT в БД  →  Publish order.created
```

Если Kafka упала после INSERT — заказ в БД есть, события нет (зависший `PENDING`).

В продакшене решают паттернами:

- **Transactional Outbox** — событие в ту же БД, отдельный процесс шлёт в Kafka;
- **Inbox** на стороне consumer — идемпотентная обработка;
- брокер с гарантиями доставки + retry.

В учебном проекте это упрощено намеренно.

---

## Rollback и компенсации

В сагах «откат» — это не `ROLLBACK` across databases, а **компенсирующие события**.

### Реализованные сценарии

| Ситуация | События | Результат |
|----------|---------|-----------|
| Оплата не прошла | `payment.failed` | `order-svc` → `CANCELLED` |
| Кухня не может готовить (закрыт / перегруз) | `order.preparation.failed` | `order-svc` → `CANCELLED` → `payment.refund.requested` → `payment.refunded` → `REFUNDED` |
| Нет свободного курьера | `delivery.failed` | то же: отмена + возврат |
| Успешный возврат | `payment.refunded` | `order-svc` → статус `REFUNDED` |

Цепочка возврата:

```text
order.preparation.failed / delivery.failed
  → order-svc: CANCELLED + order.cancelled
  → order-svc: payment.refund.requested
  → payment-svc: refund в БД + payment.refunded
  → order-svc: REFUNDED
  → notification-svc: письма пользователю
```

### Идемпотентность (базовая)

- `payment-svc` не создаёт второй платёж, если для `order_id` уже есть запись.
- `order-svc` игнорирует повторные события, если заказ уже в терминальном статусе (`CANCELLED`, `REFUNDED`, `DELIVERED`) или уже `PAID`.

### Что ещё нужно для hard prod

| Задача | Подход |
|--------|--------|
| Атомарность БД + Kafka | Transactional Outbox |
| Повторная доставка событий | Inbox + `processed_events` |
| Dead letter queue | Отдельный Kafka topic / retry policy |
| Ручная сверка платежей | Admin API + reconciliation job |

---

## Кто какие данные владеет

Принцип: **у каждого сервиса свои данные**, чужую БД напрямую не читаем.

| Данные | Владелец | Как другие узнают |
|--------|----------|-------------------|
| Пользователи, JWT | user-svc | HTTP API / gateway |
| Меню, рестораны | restaurant-svc | HTTP API |
| Заказы, статусы | order-svc | HTTP API + события |
| Платежи | payment-svc | HTTP + события; gateway подмешивает в заказ |
| Курьеры | delivery-svc | HTTP + события |
| Уведомления | notification-svc | Только из Kafka (своей БД нет) |

Дублирование намеренно минимальное: в событиях передаётся только нужное (`order_id`, `user_id`, суммы).

---

## api-gateway как BFF

Gateway — не просто прокси:

- **Проксирует** публичные маршруты (`/api/restaurants`, auth).
- **Агрегирует** `GET /api/orders/{id}`: заказ из order-svc + платёж из payment-svc.
- **Проверяет JWT** до вызова защищённых ручек.
- **Отдаёт** OpenAPI и Scalar (`/docs`, `/openapi.yaml`).

Клиент (фронт) знает только gateway — не ходит напрямую в 6 микросервисов.

---

## Фронтенд

Next.js + MUI на порту **3000**. Ходит на `NEXT_PUBLIC_API_URL` (по умолчанию `http://localhost:8090`).

Основные экраны: рестораны → меню → корзина → заказ → трекинг статуса (polling).

---

## Окружения: dev / stage / prod

Конфигурация через переменные окружения и пакет `pkg/config`.

| `APP_ENV` | Поведение |
|-----------|-----------|
| `dev` | Дефолты в коде, `SEED_DATA=true`, CORS `*` |
| `stage` | Обязательные env без дефолтов для секретов |
| `prod` | То же + `SEED_DATA=false` |

Файлы:

```text
deploy/env/dev.env              # локальная разработка (можно коммитить)
deploy/env/stage.env.example    # шаблон staging
deploy/env/prod.env.example     # шаблон production
.env.example                    # быстрый старт
```

Запуск:

```bash
make up ENV=dev      # использует deploy/env/dev.env
make down            # остановить контейнеры, данные PostgreSQL сохраняются
make down-clean      # остановить и удалить volumes (полный сброс БД)
make up ENV=stage    # нужен deploy/env/stage.env (из .example)
make up ENV=prod     # нужен deploy/env/prod.env (из .example)
```

Ключевые переменные: `JWT_SECRET`, `POSTGRES_*`, `KAFKA_BROKERS`, `*_SVC_URL`, `PAYMENT_FAIL_RATE`, `KITCHEN_FAIL_RATE`, `SEED_DATA`, `CORS_ALLOWED_ORIGINS`, `NEXT_PUBLIC_API_URL`.

Симуляции (не хардкод в коде): `KITCHEN_COOK_DURATION`, `DELIVERY_DURATION`, `PAYMENT_PROCESS_LATENCY`, `PAYMENT_REFUND_LATENCY`.

## Как запустить

```bash
cp .env.example .env   # или make up ENV=dev
make up
make docs
make demo
```

Полезные URL:

| URL | Что |
|-----|-----|
| http://localhost:3000 | Фронтенд |
| http://localhost:8090 | API Gateway |
| http://localhost:8090/docs | Scalar (документация API) |
| http://localhost:8088 | Traefik dashboard |

---

## Структура репозитория (архитектурно)

```text
pkg/
  config/     — загрузка env (dev/stage/prod)
  events/     — контракты событий (топики + payload)
  kafka/      — producer / consumer
  auth/       — JWT

services/
  user/ restaurant/ order/ payment/ delivery/ notification/ gateway/

frontend/     — Next.js клиент

deploy/       — init-db, kafka-init, traefik
docker-compose.yml
```

---

## Ограничения текущей реализации (честно)

Это **учебный MVP**, не продакшен:

1. **Хореография без оркестратора** — сложно видеть целую сагу целиком.
2. **Нет outbox / inbox** — риск рассинхрона БД и Kafka.
3. **Идемпотентность базовая** — не полноценный inbox pattern.
4. **HTTP вместо gRPC** между сервисами — ок для демо.
5. **notification-svc** только логирует, реально не шлёт SMS/push.
6. **Время готовки/доставки** — симуляция через env (`KITCHEN_COOK_DURATION`, `DELIVERY_DURATION`).

---

## Куда развивать архитектуру

| Задача | Подход |
|--------|--------|
| Надёжная доставка событий | Transactional Outbox |
| Сложные откаты | Оркестрация (Temporal) или компенсирующие события |
| Нагрузка на создание заказа | gRPC между order ↔ user/restaurant |
| Наблюдаемость саги | Correlation ID в событиях + tracing (Jaeger) |
| Идемпотентность | Таблица `processed_events` в каждом consumer |

---

## Краткий итог

| Вопрос | Ответ |
|--------|--------|
| Почему микросервисы? | Разные домены, нагрузка, независимый деплой |
| Как общаются синхронно? | HTTP/JSON (+ JWT для клиента) |
| Как общаются асинхронно? | Kafka, события в `pkg/events` |
| Сага | Хореография через топики |
| Транзакции | Локальные в PostgreSQL, распределённых нет |
| Rollback | Только отмена заказа при failed payment |
| Точка входа | api-gateway + frontend |

Поток заказа в одном предложении: **клиент создаёт заказ по HTTP → Kafka несёт события через оплату, кухню и доставку → каждый сервис обновляет своё → order-svc хранит итоговый статус заказа.**
