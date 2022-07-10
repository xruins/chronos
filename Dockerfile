FROM golang:1.18 as builder

WORKDIR /app

COPY go.* ./
COPY . ./
RUN go build -ldflags="-w -s" -o /app/chronos

FROM debian:stable-slim
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

ENV DOCKERVERSION=20.10.17
RUN curl -fsSL https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKERVERSION}.tgz | \
    tar xzv --strip 1 -C /usr/local/bin docker/docker && \
    chmod u+x /usr/local/bin/docker
COPY --from=builder --chmod=755 /app/chronos /usr/local/bin/chronos
CMD ["/usr/local/bin/chronos"]
