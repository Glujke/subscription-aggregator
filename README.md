# subscription-aggregator

REST-сервис для агрегации данных об онлайн-подписках пользователей.

## Структура проекта

```
subscription-aggregator/
├── cmd/
│   └── api/                 # точка входа, wiring зависимостей
├── internal/
│   ├── config/              # загрузка .env / yaml
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
| `PUT`    | `/api/v1/subscriptions/{id}`      | Update          |
| `DELETE` | `/api/v1/subscriptions/{id}`      | Delete          |
| `GET`    | `/api/v1/subscriptions/cost`      | Сумма за период |

## Запуск

> Скелет проекта. Реализация — в следующих шагах.

```bash
go run ./cmd/api
```
