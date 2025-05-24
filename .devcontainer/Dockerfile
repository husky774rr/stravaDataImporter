ARG VARIANT=1.24
FROM mcr.microsoft.com/devcontainers/go:${VARIANT}-bookworm

ARG GO_VERSION=1.24.1

RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o go.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

ENV PATH="/usr/local/go/bin:$PATH"
