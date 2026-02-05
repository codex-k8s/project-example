# Инструкции для ИИ-агентов (обязательные)

## Главные правила

- Строго следуй документации `docs/design-guidelines/**` (это обязательные требования проекта).
- Не редактируй `docs/design-guidelines/**` и `docs/design-guidelines/**/AGENTS.md`, если пользователь явно и однозначно не попросил изменить эти гайды.
- Если требования пользователя противоречат `docs/design-guidelines/**`, приостанови работу и предложи варианты (что считаем источником правды и как фиксируем решение).

## Что читать и когда (обязательно)

- Перед началом любой задачи:
  - `docs/design-guidelines/AGENTS.md` (быстрая навигация).
  - `docs/design-guidelines/common/project_architecture.md` (зоны, границы, где что лежит).
  - `docs/design-guidelines/common/design_principles.md` (принципы, базовые запреты/подходы).
- Если задача затрагивает переиспользуемый код/общие модули:
  - `docs/design-guidelines/common/libraries_reusable_code_requirements.md`
  - и профильный документ:
    - `docs/design-guidelines/go/libraries.md` (для Go)
    - `docs/design-guidelines/vue/libraries.md` (для фронта)
- Если задача про Go backend:
  - Сначала: `docs/design-guidelines/go/services_design_requirements.md`
  - Затем по необходимости:
    - HTTP/REST: `docs/design-guidelines/go/rest.md`
    - gRPC/Proto: `docs/design-guidelines/go/protobuf_grpc_contracts.md`, `docs/design-guidelines/go/grpc.md`
    - RabbitMQ/AsyncAPI: `docs/design-guidelines/go/mq.md`
    - WebSocket/AsyncAPI: `docs/design-guidelines/go/websockets.md`
    - Инфраструктура (Postgres/Redis/RabbitMQ/миграции): `docs/design-guidelines/go/infrastructure_integration_requirements.md`
    - Наблюдаемость: `docs/design-guidelines/go/observability_requirements.md`
    - Ошибки: `docs/design-guidelines/go/error_handling.md`
    - Комментарии/стиль: `docs/design-guidelines/go/code_commenting_rules.md`
    - Кодогенерация: `docs/design-guidelines/go/code_generation.md`
- Если задача про frontend (Vue 3 + TypeScript):
  - Сначала: `docs/design-guidelines/vue/frontend_architecture.md`
  - Затем по необходимости:
    - Данные/состояние/интеграции (axios/Pinia/router/i18n/cookies/PWA/WebSocket): `docs/design-guidelines/vue/frontend_data_and_state.md`
    - Ошибки: `docs/design-guidelines/vue/frontend_error_handling.md`
    - Кодстайл/практики: `docs/design-guidelines/vue/frontend_code_rules.md`
  - Когда затрагивается UI/визуальный дизайн: обязательно использовать `docs/design-guidelines/visual/AGENTS.md`
- Перед созданием PR (всегда):
  - Пройти общий чек-лист: `docs/design-guidelines/common/check_list.md`
  - Пройти профильный чек-лист:
    - `docs/design-guidelines/go/check_list.md` (если есть Go изменения)
    - `docs/design-guidelines/vue/check_list.md` (если есть frontend изменения)

## Полный индекс `docs/design-guidelines/**`

- `docs/design-guidelines/AGENTS.md` — краткая навигация по гайдам.

- `docs/design-guidelines/common/AGENTS.md` — навигация по common-докам.
- `docs/design-guidelines/common/check_list.md` — общий чек-лист перед PR (добавляется к профильным).
- `docs/design-guidelines/common/project_architecture.md` — зоны (internal/jobs/external/staff/dev), размещение приложений, границы ответственности.
- `docs/design-guidelines/common/design_principles.md` — принципы проектирования и кодирования (DDD/SOLID/DRY/KISS/Clean Architecture).
- `docs/design-guidelines/common/libraries_reusable_code_requirements.md` — общие правила выноса/повторного использования кода (libs).

- `docs/design-guidelines/go/AGENTS.md` — навигация по Go-докам.
- `docs/design-guidelines/go/check_list.md` — чек-лист перед PR для Go (включая команды линта/dupl).
- `docs/design-guidelines/go/services_design_requirements.md` — структура Go-сервиса, правила домена/репозиториев/SQL, правила по зонам.
- `docs/design-guidelines/go/infrastructure_integration_requirements.md` — Postgres/Redis/RabbitMQ/секреты/миграции (goose) и запреты.
- `docs/design-guidelines/go/observability_requirements.md` — логи/трейсы/метрики (OTel/Jaeger/Prometheus) и обязательные поля.
- `docs/design-guidelines/go/rest.md` — REST стек (Echo v5 + OpenAPI + kin-openapi + oapi-codegen + swagger UI).
- `docs/design-guidelines/go/grpc.md` — gRPC практики и требования (границы, обработка ошибок, ссылки на генерацию).
- `docs/design-guidelines/go/protobuf_grpc_contracts.md` — требования к `.proto` контрактам (как к транспортному слою).
- `docs/design-guidelines/go/mq.md` — RabbitMQ: AsyncAPI контракт, размещение producer/consumer, ограничения по зонам.
- `docs/design-guidelines/go/websockets.md` — WebSocket: AsyncAPI контракт, требования к обработчикам/безопасности.
- `docs/design-guidelines/go/code_generation.md` — правила и команды кодогенерации (grpc/openapi/asyncapi + фронт по спекам).
- `docs/design-guidelines/go/error_handling.md` — обязательная модель ошибок и требования к wrapping/классификации/логированию.
- `docs/design-guidelines/go/code_commenting_rules.md` — правила комментариев в Go.
- `docs/design-guidelines/go/libraries.md` — что и как выносить в `libs/go/*`.

- `docs/design-guidelines/vue/AGENTS.md` — навигация по Vue-докам.
- `docs/design-guidelines/vue/check_list.md` — чек-лист перед PR для frontend.
- `docs/design-guidelines/vue/frontend_architecture.md` — структура приложения, размещение (external/staff/dev), границы ответственности.
- `docs/design-guidelines/vue/frontend_data_and_state.md` — данные и состояние (axios/Pinia/router/i18n/cookies/PWA/WebSocket).
- `docs/design-guidelines/vue/frontend_error_handling.md` — правила обработки ошибок на фронте.
- `docs/design-guidelines/vue/frontend_code_rules.md` — правила кодирования (TS/Vue/организация импортов/комментарии).
- `docs/design-guidelines/vue/libraries.md` — что и как выносить в `libs/{vue,ts,js}`.
