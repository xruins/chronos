FROM golang:1.21 as builder

WORKDIR /app
LABEL org.opencontainers.image.source="https://github.com/xruins/chronos"

COPY . ./
RUN go build -ldflags="-w -s" -o /app/chronos

FROM debian:stable-slim
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    ca-certificates gnupg curl && \
    install -m 0755 -d /etc/apt/keyring && \
    curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg && \
    chmod a+r /etc/apt/keyrings/docker.gpg
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    docker.io \
    docker-compose && \
    docker-compose-plugin && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder --chmod=755 /app/chronos /usr/local/bin/chronos
CMD ["/usr/local/bin/chronos"]
