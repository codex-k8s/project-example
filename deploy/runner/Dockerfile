# syntax=docker/dockerfile:1.4

ARG RUNNER_IMAGE=ghcr.io/actions/actions-runner:latest

FROM gcr.io/kaniko-project/executor:debug AS kaniko

FROM ${RUNNER_IMAGE}

USER root

ENV DEBIAN_FRONTEND=noninteractive \
    TZ=UTC \
    GOPATH=/root/go \
    GOBIN=/usr/local/bin \
    PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin

RUN apt-get update -y \
  && apt-get install -y --no-install-recommends \
      ca-certificates curl wget gnupg lsb-release \
      git jq bash unzip zip make build-essential \
  && rm -rf /var/lib/apt/lists/*

RUN install -m 0755 -d /etc/apt/keyrings \
 && curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
      -o /etc/apt/keyrings/githubcli-archive-keyring.gpg \
 && chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
 && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
 && apt-get update -y \
 && apt-get install -y gh \
 && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://go.dev/dl/go1.25.1.linux-amd64.tar.gz -o /tmp/go.tgz \
 && rm -rf /usr/local/go \
 && tar -C /usr/local -xzf /tmp/go.tgz \
 && ln -sf /usr/local/go/bin/go /usr/local/bin/go \
 && rm /tmp/go.tgz

ARG CODEXCTL_VERSION=latest
RUN GOBIN="$(go env GOPATH)/bin" go install github.com/codex-k8s/codexctl/cmd/codexctl@${CODEXCTL_VERSION} \
 && cp "$(go env GOPATH)/bin/codexctl" /usr/local/bin/codexctl \
 && chmod +x /usr/local/bin/codexctl

ARG KUBECTL_VERSION=v1.34.1
RUN curl -fsSL -o /usr/local/bin/kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \
  && chmod +x /usr/local/bin/kubectl \
  && kubectl version --client --output=yaml || true

COPY --from=kaniko /kaniko/executor /kaniko/executor
RUN chmod +x /kaniko/executor

USER runner
