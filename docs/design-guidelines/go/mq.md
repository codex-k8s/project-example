# RabbitMQ (MQ) в Go

Цель: единый подход к async сообщениям (паблишинг/консьюминг), контрактам, обработке ошибок и observability.

## Контракт (AsyncAPI)
- Контракт RabbitMQ сообщений/каналов описывается в AsyncAPI YAML: `api/server/asyncapi.yaml`.
- В AsyncAPI фиксируем: версии сообщений, exchange/queue/routingKey, headers (correlation/message id), payload schemas.

## Ограничения по зонам
- RabbitMQ (AMQP) разрешён только в `services/internal/*` и `services/jobs/*`.
- `services/external/*` и `services/staff/*` не подключаются к RabbitMQ напрямую. Если edge/WS должен инициировать async-операцию, он вызывает внутренний сервис по gRPC/HTTP, а уже внутренний сервис публикует сообщение.

## Где живёт код
Consumer (входной транспорт):
- `internal/transport/mq/rabbit/*` — handlers (decode + ack/nack + вызов домена), без бизнес-логики.
- `internal/transport/mq/rabbit/middleware/*` — tracing/logging/metrics/correlation (если не в `libs/*`).
- `internal/transport/async/generated/*` — сгенерированные типы сообщений по AsyncAPI (source of truth транспорта).
- `internal/transport/mq/rabbit/messages/*` (опционально) — ручные транспортные обёртки/адаптеры, если генерации недостаточно; маппинг в домен через `internal/domain/casters/*`.

Publisher (исходящий адаптер):
- `internal/mq/rabbit/*` — публикация сообщений (publisher), не вызывается напрямую из transport.
- Доменные use-case’ы зависят от интерфейсов (портов), а не от RabbitMQ SDK.

## Обработка и надёжность
Правила:
- Идемпотентность обязательна (повторная доставка не должна ломать систему).
- Controlled retries: временные ошибки -> retry; необратимые -> DLQ.
- Для важных сообщений: `message_id` и correlation-id обязательны; publisher confirms включать осознанно.

## Observability
- В consumer’ах: связывать лог/трейс с `message_id`/correlation-id.
- Метрики: длина очереди/скорость обработки/ошибки; latency обработки.

## Ссылки
- Инфра-правила (QoS/prefetch, ack/nack, DLQ): `docs/design-guidelines/go/infrastructure_integration_requirements.md`.
- Ошибки: `docs/design-guidelines/go/error_handling.md`.
