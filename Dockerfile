# syntax=docker/dockerfile:1
FROM node:24.19.0-bookworm-slim AS frontend
ARG TARGETARCH
WORKDIR /src/web
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates \
    && rm -rf /var/lib/apt/lists/*
RUN case "$TARGETARCH" in \
      arm64) target=aarch64; digest=ff945fd7a68d25d87b47e45877228ad269d5a053c04522599b31c130d734f40d ;; \
      amd64) target=x86_64; digest=364b4cf201970cd5c943390dd3e736b4256b14a0ac52d304f8fb4acfd6a1fbd3 ;; \
      *) exit 1 ;; esac \
    && curl --fail --location --retry 3 "https://github.com/aubepkg/aube/releases/download/v2.6.1/aube-v2.6.1-${target}-unknown-linux-gnu.tar.gz" -o /tmp/aube.tar.gz \
    && echo "$digest  /tmp/aube.tar.gz" | sha256sum -c - \
    && mkdir /tmp/aube && tar -xzf /tmp/aube.tar.gz -C /tmp/aube \
    && find /tmp/aube -type f -name aube -exec install -m 0755 '{}' /usr/local/bin/aube ';' \
    && rm -rf /tmp/aube /tmp/aube.tar.gz
COPY web/package.json web/aube-lock.yaml ./
RUN aube install --frozen-lockfile
COPY web/ ./
RUN aube run --no-install build

FROM golang:1.27.1-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY db/ ./db/
RUN go mod download \
    && CGO_ENABLED=0 go build -p 2 -trimpath -ldflags="-s -w" -o /octopulse ./cmd/octopulse \
    && go clean -cache -modcache

FROM alpine:3.23 AS runtime
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 10001 octopulse && adduser -D -u 10001 -G octopulse octopulse \
    && mkdir -p /data /app/web/dist && chown 10001:10001 /data
WORKDIR /app
COPY --from=backend /octopulse /usr/local/bin/octopulse
COPY --from=frontend /src/web/dist/ /app/web/dist/
ENV OCTOPULSE_ADDR=0.0.0.0:8080 \
    OCTOPULSE_DB_DRIVER=sqlite \
    OCTOPULSE_DB_DSN=file:/data/octopulse.db \
    OCTOPULSE_DATA_DIR=/data \
    OCTOPULSE_STATIC_DIR=/app/web/dist
USER 10001:10001
EXPOSE 8080
VOLUME ["/data"]
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["octopulse"]
CMD ["serve"]
