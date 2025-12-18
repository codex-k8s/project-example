# syntax=docker/dockerfile:1.4
ARG GOLANG_IMAGE_VERSION=1.25.1-bookworm

FROM golang:${GOLANG_IMAGE_VERSION} AS project_example_chat_backend
ARG SERVICE_VERSION
ENV SERVICE_VERSION=${SERVICE_VERSION}
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN GOBIN=/usr/local/bin go install github.com/githubnemo/CompileDaemon@latest

RUN chmod +x /app/main.go || true

ENV HTTP_PORT=8080

CMD ["CompileDaemon", "-build=go build -o /app/chat-backend ./main.go", "-command=/app/chat-backend", "-graceful-kill=true"]

