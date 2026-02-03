SHELL := bash

.PHONY: help
help:
	@printf '%s\n' \
		'Цели кодогенерации:' \
		'  make gen-openapi-go SVC=services/<zone>/<service> [SPEC=api/server/api.yaml] [OUT=internal/transport/http/generated/openapi.gen.go]' \
		'  make gen-openapi-ts APP=services/<zone>/<app> SPEC=services/<zone>/<service>/api/server/api.yaml [OUT=src/shared/api/generated]' \
		'  make gen-proto-go   SVC=services/<zone>/<service> [PROTO_ROOT=proto] [PROTO=<file.proto>] [GO_OUT=internal/transport/grpc/generated] [OPENAPIV2_OUT=api/server/generated/grpc/openapi] [WITH_GATEWAY=1] [WITH_OPENAPIV2=1]' \
		'' \
		'Цели проверок (перед PR):' \
		'  make lint-go   [SVC=services/<zone>/<service>]' \
		'  make dupl-go   [SVC=services/<zone>/<service>] [DUPL_THRESHOLD=120]' \
		'  make lint      (lint-go + dupl-go)' \
		'' \
		'Установка инструментов:' \
		'  make install-proto-tools' \
		'  make install-lint-tools' \
		'  make install-codegen-tools' \
		'' \
		'Примечания:' \
		'  - Сгенерированный код/артефакты живут только под **/generated/** и не правятся руками.' \
		'  - После правок OpenAPI/AsyncAPI/proto контрактов обязательна регенерация.'

.PHONY: gen-openapi-go
gen-openapi-go:
	@if [ -z "$$SVC" ]; then echo "SVC обязателен (пример: SVC=services/external/chat_backend)"; exit 2; fi
	@tools/codegen/openapi/gen-openapi-go.sh \
		--service "$$SVC" \
		$${SPEC:+--spec "$$SPEC"} \
		$${OUT:+--out "$$OUT"}

.PHONY: gen-proto-go
gen-proto-go:
	@if [ -z "$$SVC" ]; then echo "SVC обязателен (пример: SVC=services/external/chat_backend)"; exit 2; fi
	@tools/codegen/proto/gen-proto.sh \
		--service "$$SVC" \
		$${PROTO_ROOT:+--proto-root "$$PROTO_ROOT"} \
		$${PROTO:+--proto "$$PROTO"} \
		$${GO_OUT:+--go-out "$$GO_OUT"} \
		$${OPENAPIV2_OUT:+--openapiv2-out "$$OPENAPIV2_OUT"} \
		$${WITH_GATEWAY:+--with-gateway} \
		$${WITH_OPENAPIV2:+--with-openapiv2}

.PHONY: gen-openapi-ts
gen-openapi-ts:
	@if [ -z "$$APP" ]; then echo "APP обязателен (пример: APP=services/external/web_frontend)"; exit 2; fi
	@if [ -z "$$SPEC" ]; then echo "SPEC обязателен (пример: SPEC=services/external/chat_backend/api/server/api.yaml)"; exit 2; fi
	@tools/codegen/openapi/gen-openapi-ts.sh \
		--app "$$APP" \
		--spec "$$SPEC" \
		$${OUT:+--out "$$OUT"}

.PHONY: install-proto-tools
install-proto-tools:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "Готово: protoc плагины установлены (убедись, что путь установки в PATH)."

.PHONY: install-lint-tools
install-lint-tools:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/mibk/dupl@latest
	@echo "Готово: golangci-lint и dupl установлены."

.PHONY: install-codegen-tools
install-codegen-tools:
	@go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	@go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	@if command -v npm >/dev/null 2>&1; then npm install -g @hey-api/openapi-ts @hey-api/client-axios @asyncapi/cli; else echo "npm не найден: пропускаю установку openapi-ts/asyncapi (нужно для фронта)."; fi
	@echo "Готово: codegen/contract CLI инструменты установлены (oapi-codegen, grpcurl, openapi-ts, asyncapi)."

.PHONY: lint-go
lint-go:
	@tools/lint/golangci.sh $${SVC:+--service "$$SVC"}

.PHONY: dupl-go
dupl-go:
	@tools/lint/dupl-go.sh $${SVC:+--service "$$SVC"} $${DUPL_THRESHOLD:+--threshold "$$DUPL_THRESHOLD"}

.PHONY: lint
lint: lint-go dupl-go
