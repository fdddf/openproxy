# ---------- Frontend Build Stage ----------
FROM node:20-alpine AS ui-builder
WORKDIR /build
COPY ui/package*.json ./ui/
RUN cd ui && npm ci
COPY ui ./ui
# Vite writes into src/internal/web/dist, which the Go build embeds.
COPY src/internal/web/dist/.gitkeep ./src/internal/web/dist/.gitkeep
RUN cd ui && npm run build

# ---------- Backend Build Stage ----------
FROM golang:1.25-alpine AS go-builder
WORKDIR /app

ARG BUILD_COMMIT
ARG BUILD_DATE

# Set Go module proxy
ENV GOPROXY=https://goproxy.cn,direct

# Install dependencies
RUN apk add --no-cache ca-certificates

# Copy source code, then overlay the built UI so go:embed picks it up
COPY src/ .
COPY --from=ui-builder /build/src/internal/web/dist ./internal/web/dist

# Tidy dependencies
RUN go mod tidy

# Build binary. Migrations, static assets, and the admin UI are all embedded,
# so the result is a single self-contained file.
RUN set -eux; \
  COMMIT="${BUILD_COMMIT:-dev}"; \
  DATE="${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"; \
  CGO_ENABLED=0 GOOS=linux go build -ldflags "-X github.com/fdddf/openproxy/pkg/version.Commit=${COMMIT} -X github.com/fdddf/openproxy/pkg/version.BuildDate=${DATE}" -o /app/openproxy .

# ---------- Runtime Stage ----------
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

COPY --from=go-builder /app/openproxy .
RUN chmod +x /app/openproxy

# SQLite database location. Mount a volume here to persist it.
ENV DB_PATH=/data/openproxy.db
VOLUME /data

EXPOSE 8081

# Runs against SQLite with no configuration at all. To use Postgres instead,
# set DB_DRIVER=postgres plus DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME.
# JWT_SIGN_KEY is generated and persisted on first start when unset.
CMD ["/app/openproxy"]
