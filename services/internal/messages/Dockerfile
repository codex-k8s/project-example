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

CMD ["CompileDaemon", "-build=go build -o /workspace/bin/messages ./cmd/messages", "-command=/workspace/bin/messages", "-graceful-kill=true"]

