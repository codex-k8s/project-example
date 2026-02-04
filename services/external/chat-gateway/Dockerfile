# syntax=docker/dockerfile:1.4
ARG GOLANG_IMAGE_VERSION=1.25.6-bookworm

FROM golang:${GOLANG_IMAGE_VERSION} AS chat_gateway
ARG SERVICE_VERSION
ENV SERVICE_VERSION=${SERVICE_VERSION}
ENV CGO_ENABLED=0

WORKDIR /workspace
COPY libs/go/common ./libs/go/common
COPY tools ./tools
COPY services/external/chat-gateway ./services/external/chat-gateway

WORKDIR /workspace/services/external/chat-gateway
RUN go mod download
RUN GOBIN=/usr/local/bin go install github.com/githubnemo/CompileDaemon@latest

ENV HTTP_PORT=8080

CMD ["bash", "-lc", "ROOT=\"${CODE_ROOT:-/workspace}\"; SVC_DIR=\"$ROOT/services/external/chat-gateway\"; if [ ! -d \"$SVC_DIR\" ]; then echo \"chat-gateway: code dir not found: $SVC_DIR\" >&2; exit 2; fi; mkdir -p /tmp/bin; exec CompileDaemon -directory=\"$SVC_DIR\" -directory=\"$ROOT/libs/go/common\" -directory=\"$ROOT/proto\" -exclude-dir=node_modules -exclude-dir=generated -exclude-dir=.git -exclude-dir=.local -build-dir=\"$SVC_DIR\" -run-dir=\"$SVC_DIR\" -build='go build -o /tmp/bin/chat-gateway ./cmd/chat-gateway' -command='/tmp/bin/chat-gateway' -graceful-kill=true"]
