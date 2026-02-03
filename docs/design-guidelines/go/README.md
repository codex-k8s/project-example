# Go Design Guidelines

Документы для Go backend.

- `docs/design-guidelines/go/check_list.md` — чек-лист перед PR для Go изменений.
- `docs/design-guidelines/go/services_design_requirements.md` — структура сервиса, домен/кастеры, repo+SQL правила, OpenAPI/AsyncAPI.
- `docs/design-guidelines/go/infrastructure_integration_requirements.md` — Postgres/Redis/RabbitMQ/секреты/миграции (goose) и запреты.
- `docs/design-guidelines/go/observability_requirements.md` — логи/трейсы/метрики (OTel/Jaeger/Prometheus).
- `docs/design-guidelines/go/protobuf_grpc_contracts.md` — правила gRPC `.proto` как транспортного контракта.
- `docs/design-guidelines/go/rest.md` — REST стек (echo + OpenAPI validation + codegen + swagger UI).
- `docs/design-guidelines/go/code_generation.md` — TODO по кодогенерации (grpc/openapi/asyncapi + frontend).
- `docs/design-guidelines/go/code_commenting_rules.md` — правила комментариев в Go.
- `docs/design-guidelines/go/error_handling.md` — обязательные правила обработки ошибок в Go.
- `docs/design-guidelines/go/libraries.md` — что выносить в `libs/go/*` и как.
