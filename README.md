# subscription-aggregator

REST-сервис для агрегации данных об онлайн-подписках пользователей.

## Структура проекта

```
subscription-aggregator/
├── cmd/
│   └── api/                 # точка входа, wiring зависимостей
├── internal/
│   ├── config/              # загрузка .env / yaml
│   ├── database/            # применение миграций
│   ├── domain/              # сущности, типы, валидация
│   ├── handler/             # HTTP-ручки (chi)
│   ├── repository/          # доступ к PostgreSQL (pgx)
│   └── service/             # бизнес-логика, агрегация стоимости
├── migrations/              # SQL-миграции (golang-migrate)
├── docs/                    # Swagger (генерируется swag)
├── go.mod
└── docker-compose.yml       # (позже) postgres + migrate + api
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
| Конфиг      | .env / yaml       |
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

## Тестирование

### Доменный слой

```bash
go test ./internal/domain/...
```

### Миграции (нужен запущенный Docker)

Поднимает PostgreSQL в testcontainers, применяет `up`, делает smoke `INSERT`,
откатывает `down` и проверяет, что таблица удалена.

```bash
go test ./internal/database/... -v
```

### Все тесты

```bash
go test ./...
```

## Запуск

> Скелет проекта. Реализация — в следующих шагах.

```bash
go run ./cmd/api
```
