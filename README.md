# subscription-aggregator

REST-сервис для агрегации данных об онлайн-подписках пользователей.

## Быстрый запуск

Нужен [Docker](https://docs.docker.com/get-docker/).

```bash
make up
```

Откройте в браузере: **http://localhost:8080/swagger/index.html**

Остановка:

```bash
make down
```

---

## Требования

| Инструмент | Версия |
|------------|--------|
| Go | 1.25+ (для локальной разработки) |
| Docker | с поддержкой Compose v2 |
| Make | опционально |

## Структура проекта

```
subscription-aggregator/
├── cmd/api/                 # точка входа
├── internal/
│   ├── app/                 # запуск, wiring, graceful shutdown
│   ├── config/              # загрузка .env
│   ├── database/            # golang-migrate
│   ├── domain/              # сущности, валидация, логика дат
│   ├── handler/             # HTTP (chi), Swagger-аннотации
│   ├── repository/          # PostgreSQL (pgx)
│   └── service/             # бизнес-логика, расчёт стоимости
├── migrations/              # SQL-миграции
├── docs/                    # сгенерированный Swagger
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Слои

```
HTTP Request
    → handler    (парсинг, статусы, swagger)
    → service    (правила, агрегация)
    → repository (SQL через pgxpool)
    → PostgreSQL
```

## Стек

| Компонент   | Библиотека        |
|-------------|-------------------|
| HTTP        | go-chi/chi        |
| PostgreSQL  | jackc/pgx/v5      |
| Миграции    | golang-migrate    |
| Логи        | log/slog (stdlib) |
| Конфиг      | .env              |
| Swagger     | swaggo/swag       |

## Соответствие ТЗ

| Требование | Реализация |
|------------|------------|
| CRUDL подписок | `POST/GET/PATCH/DELETE /api/v1/subscriptions`, `GET /api/v1/subscriptions/{id}` |
| Расчёт стоимости за период | `GET /api/v1/subscriptions/cost` |
| PostgreSQL + миграции | `migrations/`, сервис `migrate` в docker-compose |
| Логирование | `log/slog`, middleware запросов |
| Конфигурация `.env` | `internal/config`, `.env.example`, переменные в compose |
| Swagger | `/swagger/index.html` |
| Запуск через Docker Compose | `make up` |

## Бизнес-правила

### Подписка

- `price` — целое число рублей, > 0
- `start_date` — формат `MM-YYYY` (строго с ведущим нулём: `07-2025`)
- `end_date` — опционально; `null` = подписка активна
- `id` генерирует сервер (UUID)
- Дубликаты `(user_id, service_name)` разрешены
- Проверка существования пользователя не выполняется

### List

- `limit` по умолчанию 20, максимум 100
- `offset` — смещение
- опциональный фильтр `user_id`

### Расчёт стоимости (`/cost`)

| Параметр | Обязательность | Описание |
|----------|----------------|----------|
| `from` | да | `MM-YYYY`, включительно |
| `to` | нет | `MM-YYYY`; если не указан — текущий месяц |
| `user_id` | нет | точный UUID |
| `service_name` | нет | точное совпадение (`=`) |
| `strategy` | нет | `overlap` (default) или `sum` |

**Стратегии:**

- `overlap` — сумма `price × количество месяцев пересечения` подписки с периодом
- `sum` — сумма `price` всех подписок, пересекающих период (без умножения на месяцы)

**Границы:**

- Период и подписка считаются **включительно** по месяцам
- `end_date = null` у подписки → активна до **текущего месяца**
- Текущий месяц определяется в таймзоне **Europe/Moscow**

## API

| Метод    | Путь                              | Описание        |
|----------|-----------------------------------|-----------------|
| `POST`   | `/api/v1/subscriptions`           | Create          |
| `GET`    | `/api/v1/subscriptions`           | List            |
| `GET`    | `/api/v1/subscriptions/{id}`      | Read            |
| `PATCH`  | `/api/v1/subscriptions/{id}`      | Update          |
| `DELETE` | `/api/v1/subscriptions/{id}`      | Delete          |
| `GET`    | `/api/v1/subscriptions/cost`      | Сумма за период |
| `GET`    | `/health`                         | Liveness        |

### Примеры запросов

Сервис должен быть запущен (`make up`).

**Создать подписку:**

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H 'Content-Type: application/json' \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

**Список подписок:**

```bash
curl "http://localhost:8080/api/v1/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&limit=20&offset=0"
```

**Получить по ID** (подставьте `id` из ответа создания):

```bash
curl http://localhost:8080/api/v1/subscriptions/{id}
```

**Обновить (PATCH):**

```bash
curl -X PATCH http://localhost:8080/api/v1/subscriptions/{id} \
  -H 'Content-Type: application/json' \
  -d '{"price": 500}'
```

**Удалить:**

```bash
curl -X DELETE http://localhost:8080/api/v1/subscriptions/{id}
```

**Стоимость (overlap, по умолчанию):**

```bash
curl "http://localhost:8080/api/v1/subscriptions/cost?from=01-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"
```

**Стоимость (sum):**

```bash
curl "http://localhost:8080/api/v1/subscriptions/cost?from=01-2025&to=12-2025&strategy=sum&service_name=Yandex%20Plus"
```

## База данных

Таблица `subscriptions`:

| Колонка        | Тип           | Описание                                |
|----------------|---------------|-----------------------------------------|
| `id`           | UUID          | Первичный ключ                          |
| `service_name` | VARCHAR(255)  | Название сервиса                        |
| `price`        | INTEGER       | Стоимость в рублях, > 0                 |
| `user_id`      | UUID          | Идентификатор пользователя              |
| `start_date`   | DATE          | Первое число месяца начала              |
| `end_date`     | DATE          | Первое число месяца окончания, nullable |
| `created_at`   | TIMESTAMPTZ   | Время создания                          |
| `updated_at`   | TIMESTAMPTZ   | Время обновления                        |

В API даты — `MM-YYYY`, в БД — `DATE` с первым числом месяца (`2025-07-01`).

## Конфигурация

Для локального запуска без Docker:

```bash
cp .env.example .env
```

| Переменная        | Обязательна | По умолчанию  | Описание                         |
|-------------------|-------------|---------------|----------------------------------|
| `DATABASE_URL`    | да          | —             | Строка подключения к PostgreSQL  |
| `HTTP_ADDR`       | нет         | `:8080`       | Адрес HTTP-сервера               |
| `LOG_LEVEL`       | нет         | `info`        | `debug`, `info`, `warn`, `error` |
| `MIGRATIONS_PATH` | нет         | `migrations`  | Каталог SQL-миграций             |

В Docker переменные заданы в `docker-compose.yml`.

## Запуск

### Docker

```bash
make up          # postgres → migrate → api
make logs        # логи api
make down        # остановить и удалить контейнеры
```

### Локально

```bash
cp .env.example .env
go run ./cmd/api
```

### Swagger

Перегенерация после изменения аннотаций:

```bash
make swagger
```

## Тестирование

```bash
make test
```

Тесты репозитория и миграций требуют запущенный Docker (testcontainers):

```bash
go test ./internal/repository/... -v
go test ./internal/database/... -v
```

## Make-цели

| Команда | Описание |
|---------|----------|
| `make up` | Запустить compose |
| `make down` | Остановить compose |
| `make logs` | Логи API |
| `make test` | `go test ./...` |
| `make swagger` | Перегенерировать Swagger |
