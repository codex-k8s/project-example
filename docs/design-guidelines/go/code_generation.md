# Кодогенерация (контракты -> код)

Цель: после изменения контрактов (OpenAPI/proto/AsyncAPI) агент обязан регенерировать артефакты через `make`, закоммитить их и не править руками.

## Общие правила
- Любой сгенерированный код/артефакты живут только в директориях `**/generated/**`.
- Сгенерированное руками не правим. Любая правка = правка контракта/конфига/шаблона + регенерация.
- Источник правды транспорта:
  - REST: `api/server/api.yaml` (OpenAPI YAML)
  - gRPC: `proto/**/*.proto`
  - async: `api/server/asyncapi.yaml` (AsyncAPI YAML для RabbitMQ + WebSocket)
- Команды генерации — через `Makefile` в корне репозитория.

## OpenAPI (REST) -> Go (Echo v5)
Фиксированный стек описан в: `docs/design-guidelines/go/rest.md`.

Инструмент:
- `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`

Проектные шаблоны/конфиги:
- config: `tools/codegen/openapi/configs/go_echo_server.yaml`
- templates: `tools/codegen/openapi/templates/` (включает override для `echo/v5`)

Выход:
- `internal/transport/http/generated/openapi.gen.go`

Запуск:
```bash
make gen-openapi-go SVC=services/<zone>/<service>
```

Если нужно указать иной путь вывода (обязательно под `**/generated/**`):
```bash
make gen-openapi-go SVC=services/<zone>/<service> OUT=internal/transport/http/generated/openapi.gen.go
```

## Protobuf/gRPC -> Go (+ опционально grpc-gateway + OpenAPIv2 для dev)
Инструменты (protoc плагины):
- `google.golang.org/protobuf/cmd/protoc-gen-go@latest`
- `google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`
- `github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest`
- `github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest`

Установка плагинов (в контейнере агента уже установлено):
```bash
make install-proto-tools
```

Выход:
- Go: `internal/transport/grpc/generated/**` (в т.ч. `*.pb.go`, `*_grpc.pb.go`, `*.pb.gw.go` если включён gateway)
- OpenAPIv2 (dev): `api/server/generated/grpc/openapi/**`

Запуск (все `.proto` сервиса):
```bash
make gen-proto-go SVC=services/<zone>/<service>
```

Запуск с grpc-gateway + OpenAPIv2:
```bash
make gen-proto-go SVC=services/<zone>/<service> WITH_GATEWAY=1 WITH_OPENAPIV2=1
```

Если нужно указать иные директории вывода (обязательно под `**/generated/**`):
```bash
make gen-proto-go SVC=services/<zone>/<service> GO_OUT=internal/transport/grpc/generated OPENAPIV2_OUT=api/server/generated/grpc/openapi WITH_OPENAPIV2=1
```

Запуск для одного файла:
```bash
make gen-proto-go SVC=services/<zone>/<service> PROTO=proto/<path>/<file>.proto WITH_OPENAPIV2=1
```

## AsyncAPI (RabbitMQ + WebSocket)
Контракт: `api/server/asyncapi.yaml`.

Правило: контракт должен быть валидным и соответствовать реальной реализации (версии сообщений, bindings, schema).
Важно по зонам:
- `services/external|staff/*`: AsyncAPI описывает только WebSocket (ws/http bindings), прямое AMQP запрещено.
- `services/internal|jobs/*`: AsyncAPI может описывать RabbitMQ (amqp bindings) и/или WebSocket, если сервис их реально использует.

Инструмент:
- `@asyncapi/cli` (AsyncAPI CLI)

Валидация (обязательно, перед PR при изменениях AsyncAPI):
```bash
make validate-asyncapi SVC=services/<zone>/<service>
```

Генерация моделей (AsyncAPI -> Go):
- выход: `internal/transport/async/generated/**`
```bash
make gen-asyncapi-go SVC=services/<zone>/<service>
```

Генерация моделей (AsyncAPI -> TypeScript) для frontend:
- выход (по умолчанию): `src/shared/ws/generated/**`
```bash
make gen-asyncapi-ts APP=services/<zone>/<app> SPEC=services/<zone>/<service>/api/server/asyncapi.yaml
```

## Frontend codegen по OpenAPI (TypeScript + Axios)
Рекомендуемый инструмент:
- `@hey-api/openapi-ts` + `@hey-api/client-axios`

Рекомендованный выход (внутри конкретного фронта):
- `src/shared/api/generated/**`

Запуск (через `make`, рекомендуется):
```bash
make gen-openapi-ts APP=services/<zone>/<app> SPEC=services/<zone>/<service>/api/server/api.yaml
```

Запуск (пример напрямую через `npx`):
```bash
npx @hey-api/openapi-ts -i api/server/api.yaml -o src/shared/api/generated -c @hey-api/client-axios
```
