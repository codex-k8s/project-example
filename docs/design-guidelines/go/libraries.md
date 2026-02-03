# Go: что выносить в `libs/go/*`

Цель: уменьшать дублирование между сервисами без “god-lib” и без протечки бизнес-логики конкретного домена.

## Когда выносить
- код нужен >= 2 сервисам;
- требуется единый стандарт поведения (логирование/метрики/otel, middlewares, клиенты);
- API библиотеки можно сделать минимальным и стабильным.

## Что обычно выносим
- `libs/go/observability/*` — логгер, метрики, OTel helpers.
- `libs/go/transport/*` — HTTP/gRPC middleware/interceptors общего назначения.
- `libs/go/storage/*` — postgres helpers (пулы, транзакции), tooling для миграций.
- `libs/go/mq/*` — общие утилиты для RabbitMQ (подключение, retries, ack/nack helpers), без доменных сообщений.
- `libs/go/cache/*` — redis helpers (TTL, key naming helpers), без доменной логики.

## Что запрещено выносить
- доменные правила конкретного сервиса;
- “общие DTO” продукта, если они являются транспортными моделями (для этого есть `proto/` или OpenAPI/AsyncAPI контракты);
- тяжёлые зависимости ради одной функции.

## Контракты транспорта
- gRPC правила см. `docs/design-guidelines/go/protobuf_grpc_contracts.md`.
- Ошибки см. `docs/design-guidelines/go/error_handling.md`.
