# users (internal)

Ответственность: регистрация/аутентификация пользователей и хранение пользователей в своей БД (PostgreSQL).

Входы:
- gRPC `UsersService` (порт `GRPC_PORT`, по умолчанию 8080)
- HTTP тех. эндпоинты (порт `HTTP_PORT`, по умолчанию 8081): `/health/*`, `/metrics`

Выходы:
- Postgres (env: `POSTGRES_*`)

Миграции:
- SQL миграции: `cmd/cli/migrations/*.sql` (goose)
- запуск: `go run ./cmd/cli migrate up`

