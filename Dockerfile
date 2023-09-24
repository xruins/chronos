FROM golang:1.21 as builder

WORKDIR /app
LABEL org.opencontainers.image.source="https://github.com/xruins/chronos"

COPY . ./
RUN go build -ldflags="-w -s" -o /app/chronos

FROM debian:stable-slim
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    ca-certificates \
    docker.io \
    docker-compose && \
    docker-compose-plugin && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder --chmod=755 /app/chronos /usr/local/bin/chronos
CMD ["/usr/local/bin/chronos"]
