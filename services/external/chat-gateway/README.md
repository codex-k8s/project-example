# chat-gateway (external)

Ответственность: внешний gateway/BFF для чата (REST + WebSocket), auth через cookie-сессии, интеграции внутрь по gRPC.

Входы:
- HTTP/REST (OpenAPI): `api/server/api.yaml`
- WebSocket (AsyncAPI): `api/server/asyncapi.yaml`
- Тех. эндпоинты: `/health/*`, `/metrics`, Swagger UI: `/docs`

Выходы:
- gRPC: `users:8080`, `messages:8080` (env: `USERS_GRPC_ADDR`, `MESSAGES_GRPC_ADDR`)
- Redis (сессии): env `REDIS_*`

