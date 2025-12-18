# syntax=docker/dockerfile:1.4
ARG NODE_IMAGE_VERSION=22-bookworm

FROM node:${NODE_IMAGE_VERSION} AS web_frontend
ARG SERVICE_VERSION
ENV SERVICE_VERSION=${SERVICE_VERSION}

WORKDIR /app

COPY package.json ./
RUN npm install

COPY . .

EXPOSE 5173

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]
