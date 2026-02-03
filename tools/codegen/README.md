# Codegen (project-wide)

Цель: единые конфиги/шаблоны и единый запуск кодогенерации через `make` в корне репозитория.

Точки входа:
- OpenAPI -> Go: `tools/codegen/openapi/gen-openapi-go.sh` (конфиг `tools/codegen/openapi/configs/go_echo_server.yaml`, templates `tools/codegen/openapi/templates/`)
- OpenAPI -> TS: `tools/codegen/openapi/gen-openapi-ts.sh`
- AsyncAPI validate/models: `tools/codegen/asyncapi/*`
- Proto -> Go: `tools/codegen/proto/gen-proto.sh`

См. также: `docs/design-guidelines/go/code_generation.md`.
