# messages (internal)

Ответственность: хранение сообщений и трансляция событий изменений через gRPC-stream.

Входы:
- gRPC `MessagesService` (порт `GRPC_PORT`, по умолчанию 8080)
- HTTP тех. эндпоинты (порт `HTTP_PORT`, по умолчанию 8081): `/health/*`, `/metrics`

Выходы:
- Postgres (env: `POSTGRES_*`)

Миграции:
- SQL миграции: `cmd/cli/migrations/*.sql` (goose)
- запуск: `go run ./cmd/cli migrate up`

