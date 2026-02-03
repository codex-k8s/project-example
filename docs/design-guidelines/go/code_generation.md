# Code Generation (Go/Vue): TODO

Этот документ — точка сборки требований по кодогенерации (server/client) из транспортных контрактов:
- gRPC/protobuf
- OpenAPI (REST)
- AsyncAPI (RabbitMQ/WebSocket)

TODO: расписать единый стандарт “где лежат спеки”, “что генерим”, “куда кладём”, “как запускать в CI”, “как не коммитить мусор”.

## OpenAPI (REST) -> Go
Инструменты:
- `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`

TODO:
- структура пакетов для generated code (types/client/server);
- `cfg.yaml` шаблоны (models-only, client-only, server-only);
- правила обновления спеки и регенерации;
- где храним примеры/fixtures для контракт-тестов.

## Protobuf/gRPC -> Go
Инструменты (устанавливаемые плагины):
- `google.golang.org/protobuf/cmd/protoc-gen-go@latest`
- `google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`
- `github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest`
- `github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest`

Установка (пример):
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

TODO:
- единый способ запуска (Makefile/buf/protoc) и директории output;
- правила для аннотаций grpc-gateway;
- генерация OpenAPI из proto для dev/dev-ai окружений (grpc-gateway openapiv2), и как/где её отдаём;
- генерация dev-grpc HTTP сервера/прокси для отладки (swagger/openapi).

## AsyncAPI -> (Go + Vue)
Контракт:
- `api/server/asyncapi.yaml` (YAML) — единый источник правды для async сообщений RabbitMQ/WebSocket.

TODO:
- выбрать стандарт генерации:
  - вариант A: генерировать типы сообщений (Go + TS) и вручную писать wiring (producer/consumer/ws);
  - вариант B: использовать генератор SDK (если выбираем конкретный инструмент);
- шаблоны:
  - Go: producer/consumer (RabbitMQ), ws-client/ws-server message dispatch;
  - Vue: типы сообщений + безопасный парсер/диспетчер + Pinia интеграция;
- схема версионирования сообщений и миграций.

## Frontend codegen по OpenAPI
TODO: выбрать инструмент для генерации TS типов/клиента из OpenAPI и зафиксировать паттерн интеграции с axios.

Варианты (кандидаты):
- `openapi-generator` (TS client + axios), если нужна “тяжёлая” генерация с множеством языков.
- `openapi-ts` / генераторы TS типов, если хотим компактный типизированный слой поверх собственного axios клиента.

