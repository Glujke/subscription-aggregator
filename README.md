# subscription-aggregator

REST-сервис для агрегации данных об онлайн-подписках пользователей.

## Структура проекта

```
subscription-aggregator/
├── cmd/
│   └── api/                 # точка входа, wiring зависимостей
├── internal/
│   ├── app/                 # сборка и запуск сервиса
│   ├── config/              # загрузка .env
│   ├── database/            # применение миграций
│   ├── domain/              # сущности, типы, валидация
│   ├── handler/             # HTTP-ручки (chi)
│   ├── repository/          # доступ к PostgreSQL (pgx)
│   └── service/             # бизнес-логика, агрегация стоимости
├── migrations/              # SQL-миграции (golang-migrate)
├── docs/                    # Swagger (генерируется swag)
├── go.mod
└── docker-compose.yml       # postgres + migrate + api
```

## Слои

```
HTTP Request
    → handler    (парсинг, статусы, swagger-аннотации)
    → service    (правила, агрегация, валидация домена)
    → repository (SQL через pgxpool)
    → PostgreSQL
```

## Планируемый стек

| Компонент   | Библиотека        |
|-------------|-------------------|
| HTTP        | go-chi/chi        |
| PostgreSQL  | jackc/pgx/v5      |
| Миграции    | golang-migrate    |
| Логи        | log/slog (stdlib) |
| Конфиг      | .env              |
| Swagger     | swaggo/swag       |

## API (черновик)

| Метод    | Путь                              | Описание        |
|----------|-----------------------------------|-----------------|
| `POST`   | `/api/v1/subscriptions`           | Create          |
| `GET`    | `/api/v1/subscriptions`           | List            |
| `GET`    | `/api/v1/subscriptions/{id}`      | Read            |
| `PATCH`  | `/api/v1/subscriptions/{id}`      | Update          |
| `DELETE` | `/api/v1/subscriptions/{id}`      | Delete          |
| `GET`    | `/api/v1/subscriptions/cost`      | Сумма за период |

## База данных

Таблица `subscriptions`:

| Колонка        | Тип           | Описание                                      |
|----------------|---------------|-----------------------------------------------|
| `id`           | UUID          | Первичный ключ, генерируется сервером         |
| `service_name` | VARCHAR(255)  | Название сервиса                              |
| `price`        | INTEGER       | Стоимость в рублях, > 0                       |
| `user_id`      | UUID          | Идентификатор пользователя                    |
| `start_date`   | DATE          | Первое число месяца начала подписки           |
| `end_date`     | DATE          | Первое число месяца окончания, nullable       |
| `created_at`   | TIMESTAMPTZ   | Время создания записи                         |
| `updated_at`   | TIMESTAMPTZ   | Время последнего обновления                   |

В API даты передаются как `MM-YYYY` (например, `07-2025`). В БД хранятся как `DATE`
с первым числом месяца (`2025-07-01`).

Индексы: `user_id`, `service_name`, `(start_date, end_date)`.

Миграции лежат в `migrations/` и применяются через [golang-migrate](https://github.com/golang-migrate/migrate).

## Конфигурация

Скопируйте пример и отредактируйте под своё окружение:

```bash
cp .env.example .env
```

| Переменная        | Обязательна | По умолчанию  | Описание                          |
|-------------------|-------------|---------------|-----------------------------------|
| `DATABASE_URL`    | да          | —             | Строка подключения к PostgreSQL   |
| `HTTP_ADDR`       | нет         | `:8080`       | Адрес HTTP-сервера                |
| `LOG_LEVEL`       | нет         | `info`        | `debug`, `info`, `warn`, `error`  |
| `MIGRATIONS_PATH` | нет         | `migrations`  | Каталог SQL-миграций              |

Текущий месяц для расчётов определяется в таймзоне `Europe/Moscow` (`config.Location()`).

## Тестирование

### Доменный слой

```bash
go test ./internal/domain/...
```

### Конфигурация

```bash
go test ./internal/config/...
```

### HTTP-слой

```bash
go test ./internal/handler/...
```

### Сервисный слой

```bash
go test ./internal/service/...
```

### Репозиторий (нужен запущенный Docker)

CRUDL, фильтры и маппинг дат — через testcontainers.

```bash
go test ./internal/repository/... -v
```

### Миграции (нужен запущенный Docker)

Поднимает PostgreSQL в testcontainers, применяет `up`, делает smoke `INSERT`,
откатывает `down` и проверяет, что таблица удалена.

```bash
go test ./internal/database/... -v
```

### Приложение

```bash
go test ./internal/app/...
```

### Все тесты

```bash
go test ./...
```

## Запуск

### Docker (рекомендуется)

```bash
make up
# или
docker compose up --build -d
```

Проверка:

```bash
make test-health
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H 'Content-Type: application/json' \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

Остановка:

```bash
make down
```

### Swagger

После `make up` документация доступна по адресу:

```
http://localhost:8080/swagger/index.html
```

Перегенерация (после изменения аннотаций):

```bash
make swagger
```

Сервисы: `postgres` → `migrate` (one-shot) → `api`. Переменные окружения заданы в `docker-compose.yml`.

### Локально (нужен PostgreSQL)

```bash
cp .env.example .env
go run ./cmd/api
```

```bash
curl http://localhost:8080/health
```
