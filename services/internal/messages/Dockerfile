# syntax=docker/dockerfile:1.4
ARG GOLANG_IMAGE_VERSION=1.25.6-bookworm

FROM golang:${GOLANG_IMAGE_VERSION} AS messages
ARG SERVICE_VERSION
ENV SERVICE_VERSION=${SERVICE_VERSION}
ENV CGO_ENABLED=0

WORKDIR /workspace
COPY libs/go/common ./libs/go/common
COPY services/internal/messages ./services/internal/messages

WORKDIR /workspace/services/internal/messages
RUN go mod download
RUN GOBIN=/usr/local/bin go install github.com/githubnemo/CompileDaemon@latest

ENV GRPC_PORT=8080
ENV HTTP_PORT=8081

CMD ["bash", "-lc", "ROOT=\"${CODE_ROOT:-/workspace}\"; SVC_DIR=\"$ROOT/services/internal/messages\"; if [ ! -d \"$SVC_DIR\" ]; then echo \"messages: code dir not found: $SVC_DIR\" >&2; exit 2; fi; mkdir -p /tmp/bin; exec CompileDaemon -directory=\"$SVC_DIR\" -directory=\"$ROOT/libs/go/common\" -directory=\"$ROOT/proto\" -exclude-dir=node_modules -exclude-dir=generated -exclude-dir=.git -exclude-dir=.local -build-dir=\"$SVC_DIR\" -run-dir=\"$SVC_DIR\" -build='go build -o /tmp/bin/messages ./cmd/messages' -command='/tmp/bin/messages' -graceful-kill=true"]
