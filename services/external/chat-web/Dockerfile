# syntax=docker/dockerfile:1.4
ARG NODE_IMAGE_VERSION=22-bookworm

FROM node:${NODE_IMAGE_VERSION} AS chat_web
ARG SERVICE_VERSION
ENV SERVICE_VERSION=${SERVICE_VERSION}

WORKDIR /app
COPY . .

ENV NODE_ENV=development

CMD ["sh", "-c", "if [ ! -d /app/node_modules ]; then npm install --no-package-lock; fi; exec npm run dev -- --host 0.0.0.0 --port 5173"]

