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

CMD ["CompileDaemon", "-build=go build -o /workspace/bin/chat-gateway ./cmd/chat-gateway", "-command=/workspace/bin/chat-gateway", "-graceful-kill=true"]

