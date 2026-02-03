# Сервисы: требования к проектированию

Цель: все сервисы одинаково устроены “снаружи”, имеют ясные доменные границы, стандартизированное взаимодействие и observability “из коробки”.

## Сервис: ответственность, имя и размещение

Имена:
- `kebab-case`.
- Имя отражает домен/роль; технологии в имени — только если это и есть домен (напр. “pdf”).

Размещение:
- `services/<zone>/<service-name>/`, где `<zone>` ∈ `internal|external|staff|jobs|dev`.

Правила ответственности:
- `internal`: бизнес-логика + владение данными.
- `external`/`staff`: thin-edge (валидация, authn/authz, маршрутизация, агрегация; без доменных правил).
- `jobs`: cron/worker/consumer, без публичного API.
- `dev`: только dev, отключено в prod окружениях.
  - UI/приложения (frontend) размещаются в `external` или `staff` в зависимости от аудитории.

Мини-документация сервиса:
- В каждом сервисе обязателен краткий `README.md`: назначение, входы/выходы (HTTP/gRPC/queue), ключевые env.

## Выбор протокола: HTTP/REST, gRPC, WebSocket, очередь

### gRPC (внутренний синхронный протокол)
Правила:
- Внутренние sync-вызовы сервис-сервис.
- Контракты в `proto/`; изменения версионируемы и обратно совместимы.
- Ошибки маппятся в gRPC status codes (единым правилом) и документируются.

### HTTP/REST (внешний интерфейс)
Правила:
- Внешние клиенты/браузеры/интеграции.
- Версионирование (`/api/v1/...`), корректные HTTP status codes, единый формат ошибок.
- Для public endpoints обязательны auth и rate limiting.
- HTTP описываем через OpenAPI; для Go-сервисов используем `github.com/getkin/kin-openapi` для загрузки/валидации спеки и запросов.
- Для `external|staff` OpenAPI-спека хранится в `api/server/api.yaml` сервиса.

### WebSocket (real-time)
Правила:
- Real-time push/двунаправленность.
- Фиксированный формат сообщений (тип + payload), ping/pong, таймауты, лимиты.
- Масштабирование: sticky sessions или broker.
- Для описания WS протокола/сообщений используем AsyncAPI: `api/server/asyncapi.yaml` (единый контракт сообщений; без привязки к реализации).

### RabbitMQ (асинхронно)
Правила:
- Async задачи/события/интеграции; слабая связность; устойчивость к сбоям.
- Идемпотентность, контролируемые ретраи (лимит+DLQ), `message_id`/correlation-id для важных цепочек.

## Внутренняя структура сервиса и слои

Слои (идея): `transport` (вход/выход) -> `domain` (правила/модели/порты); инфраструктура (БД/кеш/очереди/внешние клиенты) подключается снаружи через реализации интерфейсов.

Ключевое: `proto/` и `api/server/api.yaml` — источники правды **транспорта**. Домен имеет свои модели и “кастеры” (маппинг) между транспортом/хранилищем и доменом.

### Рекомендуемый каркас Go-сервиса (пример)
Внутри `services/<zone>/<service-name>/`:

- `cmd/<service-name>/main.go` — тонкий вход.
- `internal/app/` — composition root + graceful shutdown.
- `internal/transport/{http,grpc,ws}/` — handlers/registration без бизнес-логики.
- `internal/transport/{http,grpc,ws}/middleware/` — middleware/interceptors, специфичные для сервиса (не из `libs/*`).
- `internal/transport/helpers/` — неэкспортируемые helpers для транспорта.

- `internal/domain/` — доменная область (правила, модели, порты).
- `internal/domain/service/` — доменная бизнес-логика.
- `internal/domain/errs/` — доменные typed errors (если нужны).
- `internal/domain/casters/` — маппинг (transport DTO <-> domain, persistence <-> domain); в домене нет зависимостей от transport/pgx.
- `internal/domain/helpers/` — неэкспортируемые helpers домена (валидация, нормализация, конвертеры и т.п.).
- `internal/domain/types/` — доменные модели, разнесённые по категориям:
  - `internal/domain/types/entity/<model_name>.go` — сущности домена (допустимы `db`-теги для маппинга в repo-слое на `pgx`, напр. поле `ID int64` с тегом `db:"id"`; без зависимостей от pgx).
  - `internal/domain/types/value/<model_name>.go` — value objects.
  - `internal/domain/types/enum/<model_name>.go` — enum-подобные типы.
  - `internal/domain/types/query/<model_name>.go` — входы/фильтры поиска, параметры запросов use-case.
  - `internal/domain/types/mixin/<model_name>.go` — общие “встраиваемые” фрагменты (time ranges, paging и т.п.).
- `internal/domain/repository/<model>/repository.go` — интерфейсы репозиториев (порты домена).

- `internal/repository/postgres/` — реализации репозиториев на `pgx`.
- `internal/repository/postgres/<model>/repository.go` — реализация интерфейса домена.
- `internal/repository/postgres/<model>/sql/*.sql` — SQL-запросы (строго из файлов, через `//go:embed`).
  - каждый запрос в отдельном `.sql`;
  - запросы именуются комментариями в стиле `-- name: <model>__<operation> :one|:many|:exec` (для стабильной привязки к коду);
  - сложные запросы допускают шаблонизацию (`text/template`) в `.sql`, с явными параметрами.
- `internal/repository/postgres/helpers/` — неэкспортируемые helpers репозитория (скан, построители параметров и т.п.).
- `internal/cache/redis/` и `internal/cache/redis/helpers/` — cache адаптер и его helpers.
- `internal/mq/rabbit/` и `internal/mq/rabbit/helpers/` — MQ адаптер и его helpers.
- `internal/observability/` — подключение логов/метрик/трейсов (или через `libs/*`).
- `cmd/cli/` — сервисная CLI (в т.ч. миграции).
- `cmd/cli/migrations/*.sql` — миграции БД (goose; timestamp-предфиксированные файлы).
- `api/server/api.yaml` — OpenAPI (для `external|staff`).
- `api/server/asyncapi.yaml` — AsyncAPI (для WebSocket/async сообщений, если используется).

Запрещено:
- доменная логика в `transport/*`,
- прямые импорты драйверов БД из домена,
- смешивать доменные модели с транспортными DTO без явного маппинга (casters).
- SQL-запросы строками в Go-коде (все запросы — в отдельных `.sql` файлах, через `//go:embed`).

## Нефункциональные требования к каждому сервису

Обязательное:
- Health: `/health/livez` (process alive), `/health/readyz` (готовность + критичные зависимости).
- Metrics: `/metrics` (Prometheus).
- Логи: структурированные в stdout/stderr; без секретов/PII.
- Трейсинг: OpenTelemetry + корректная пропагация контекста.
- Ошибки: оборачивать контекстом, не “глотать”, корректно маппить в HTTP/gRPC, не логировать секреты.
- Graceful shutdown (серверы/коннекты закрываются корректно).
- Конфигурация: только env/конфиг; никаких хардкодов секретов/адресов.
- Безопасность: `external|staff` = authn/authz; staff = аудит; external = rate limiting; `internal|jobs` = сетевые ограничения.
- Производительность: stateless по умолчанию, ресурсы/лимиты, таймауты на внешние вызовы.
