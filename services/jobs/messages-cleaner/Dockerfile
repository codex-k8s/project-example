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

CMD ["go", "run", "./cmd/messages-cleaner"]

