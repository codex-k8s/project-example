# WebSocket в Go

Цель: единый подход к real-time каналам, контрактам сообщений и обработке ошибок.

## Контракт (AsyncAPI)
- Контракт WebSocket сообщений описывается в AsyncAPI YAML: `api/server/asyncapi.yaml`.
- В AsyncAPI фиксируем: каналы, типы сообщений, payload schemas, обязательные поля корреляции (если есть), версии сообщений.

## Типы сообщений
- Типы сообщений генерируются из AsyncAPI в `internal/transport/async/generated/*` (см. `docs/design-guidelines/go/code_generation.md`).
- Запрещено вручную править сгенерированные типы; маппинг в домен — через casters.

## Серверный слой
Правила:
- WS handlers тонкие: handshake/auth, чтение/парсинг сообщений, маппинг в домен, вызов доменных use-case’ов.
- Нельзя держать “бизнес-правила” в обработчиках сообщений; только orchestration/dispatch.
- Heartbeat обязателен (ping/pong или app-level), таймауты и лимиты соединений обязательны.

## Observability
- Логировать/трейсить ключевые события: connect/disconnect, ошибки парсинга/валидации, отправка сообщений (без PII).
- Корреляция (trace_id/request_id/message_id) — если поддерживается протоколом.

## Ссылки
- Ошибки: `docs/design-guidelines/go/error_handling.md`.
