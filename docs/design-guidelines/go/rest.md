# REST/HTTP в Go

Цель: единый стек, единый контракт (OpenAPI YAML), единая валидация и единая точка обработки ошибок.

## Стек (фиксирован)
- HTTP сервер: `github.com/labstack/echo/v5`
- OpenAPI (загрузка спеки + request/response validation): `github.com/getkin/kin-openapi`
- Codegen по OpenAPI: `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`
- Swagger UI (отдача спеки/доков): `github.com/swaggest/swgui/v5emb`

## Где хранится OpenAPI
- Спека сервиса: `api/server/api.yaml` (для `services/external/*` и `services/staff/*`).
- (Опционально) JSON schema/примеры можно держать рядом в `api/server/*`, но источником правды остаётся OpenAPI YAML.

## Как используем `kin-openapi`
Флоу (по смыслу):
1) На старте загружаем `api/server/api.yaml` (разрешая external refs при необходимости) и валидируем документ.
2) На входе запроса валидируем request по операции:
   - `openapi3.Loader` -> `doc.Validate(ctx)` -> `router.FindRoute(req)` -> `openapi3filter.ValidateRequest(...)`.
3) На выходе (опционально, минимум в dev/stage) валидируем response:
   - `openapi3filter.ValidateResponse(...)`.
4) Ошибки валидации нормализуем в единый безопасный HTTP контракт ошибок (см. `docs/design-guidelines/go/error_handling.md`).

Важно:
- Handler’ы не повторяют schema/type validation (это ответственность OpenAPI middleware).
- Сообщение ошибки валидатора не считается публичным контрактом; наружу отдаём безопасные поля (например, `message/loc/field`).

## Codegen (OpenAPI -> Go)
Генерация запускается через `go generate` / CI, а конфигурация хранится в репо.

Рекомендуемый подход:
- хранить `cfg.yaml` (oapi-codegen config) рядом со спекой или рядом с генерируемым пакетом;
- генерировать отдельно types + server/client (по необходимости), не смешивая с бизнес-логикой;
- не редактировать сгенерённый файл руками.

Пример `//go:generate`:
```go
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config cfg.yaml api/server/api.yaml
package api
```

## Swagger UI
Правило:
- В `external|staff` сервисах отдаём Swagger UI, который показывает `api/server/api.yaml` (и статически, и/или через эндпоинт).
- В prod это включается осознанно (по политике безопасности продукта), в dev/stage — обычно включено.

