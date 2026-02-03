# Go чек-лист перед PR

Используется как self-check перед созданием PR. В PR достаточно написать: “чек-лист выполнен, релевантно N пунктов, все выполнены”.

## Архитектура и структура
- Структура сервиса соответствует `docs/design-guidelines/go/services_design_requirements.md` (domain/transport/repository разделены; нет доменной логики в transport).
- HTTP (если есть): OpenAPI в `api/server/api.yaml`; реализация REST стека и правила validation/codegen/swagger — в `docs/design-guidelines/go/rest.md`.
- Async (если есть WS или RabbitMQ): AsyncAPI в `api/server/asyncapi.yaml` (YAML), описаны каналы/сообщения/версии/bindings (см. `docs/design-guidelines/go/websockets.md` и `docs/design-guidelines/go/mq.md`).

## Postgres и SQL (если есть)
- Миграции: `cmd/cli/migrations/*.sql` (goose; timestamp; `-- +goose Up/Down`); история не переписывается.
- Repo интерфейсы в `internal/domain/repository/<model>/repository.go`; реализации в `internal/repository/postgres/<model>/repository.go`.
- SQL только в `internal/repository/postgres/<model>/sql/*.sql` (по файлу на запрос) + `//go:embed`; SQL-строки в Go запрещены.
- SQL имеет имена `-- name: <model>__<operation> :one|:many|:exec`; сложные запросы допускают шаблонизацию в `.sql` с явными параметрами.

## RabbitMQ (если есть)
- Консьюмеры — входной транспорт: `internal/transport/mq/rabbit/*` (handlers), без бизнес-логики внутри handler’ов (только маппинг + вызов домена).
- Продюсеры/паблишеры — исходящий адаптер: `internal/mq/rabbit/*`, не дергаются напрямую из transport; домен зависит от интерфейса, а не от RabbitMQ SDK.
- Топики/очереди/роутинг-ключи не захардкожены строками: берутся из конфигурации/констант, отражены в AsyncAPI.
- Идемпотентность обработки, controlled retries и DLQ учтены (см. infra гайд).

## Observability
- Логи структурированные; `trace_id` обязателен; нет секретов/PII; вход логируется в middleware/interceptor; нет дублирования ошибок ниже границы.
- Трейсы: входящие операции создают/продолжают trace; внешние вызовы (DB/gRPC/HTTP/queue) в отдельных спанах.
- Метрики: `/metrics`; минимум — ops/errors/latency(hist)/runtime; лейблы низкой кардинальности.

## Ошибки
- Соблюдён `docs/design-guidelines/go/error_handling.md`.

## Protobuf/gRPC (если есть)
- Соблюдён `docs/design-guidelines/go/protobuf_grpc_contracts.md`.

## Комментарии (если затрагивается Go-код)
- Соблюдён `docs/design-guidelines/go/code_commenting_rules.md`.
