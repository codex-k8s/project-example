# syntax=docker/dockerfile:1.4
ARG GOLANG_IMAGE_VERSION=1.25.6-bookworm

FROM golang:${GOLANG_IMAGE_VERSION} AS messages_cleaner
ARG SERVICE_VERSION
ENV SERVICE_VERSION=${SERVICE_VERSION}
ENV CGO_ENABLED=0

WORKDIR /workspace
COPY libs/go/common ./libs/go/common
COPY services/jobs/messages-cleaner ./services/jobs/messages-cleaner

WORKDIR /workspace/services/jobs/messages-cleaner
RUN go mod download

CMD ["bash", "-lc", "ROOT=\"${CODE_ROOT:-/workspace}\"; SVC_DIR=\"$ROOT/services/jobs/messages-cleaner\"; if [ ! -d \"$SVC_DIR\" ]; then echo \"messages-cleaner: code dir not found: $SVC_DIR\" >&2; exit 2; fi; cd \"$SVC_DIR\"; exec go run ./cmd/messages-cleaner"]
