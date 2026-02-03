# syntax=docker/dockerfile:1.4
ARG GOLANG_IMAGE_VERSION=1.25.6-bookworm

FROM golang:${GOLANG_IMAGE_VERSION} AS users
ARG SERVICE_VERSION
ENV SERVICE_VERSION=${SERVICE_VERSION}
ENV CGO_ENABLED=0

WORKDIR /workspace
COPY libs/go/common ./libs/go/common
COPY services/internal/users ./services/internal/users

WORKDIR /workspace/services/internal/users
RUN go mod download
RUN GOBIN=/usr/local/bin go install github.com/githubnemo/CompileDaemon@latest

ENV GRPC_PORT=8080
ENV HTTP_PORT=8081

CMD ["CompileDaemon", "-build=go build -o /workspace/bin/users ./cmd/users", "-command=/workspace/bin/users", "-graceful-kill=true"]

