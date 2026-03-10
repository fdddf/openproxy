# ---------- Frontend Build Stage ----------
FROM node:20-alpine AS ui-builder
WORKDIR /ui
COPY ui/package*.json ./
RUN npm ci
COPY ui .
RUN npm run build

# ---------- Backend Build Stage ----------
FROM golang:1.25-alpine AS go-builder
WORKDIR /app

ARG BUILD_COMMIT
ARG BUILD_DATE

# Set Go module proxy
ENV GOPROXY=https://goproxy.cn,direct

# Install dependencies
RUN apk add --no-cache ca-certificates

# Copy source code
COPY src/ .

# Tidy dependencies
RUN go mod tidy

# Build binary
RUN set -eux; \
  COMMIT="${BUILD_COMMIT:-dev}"; \
  DATE="${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"; \
  CGO_ENABLED=0 GOOS=linux go build -ldflags "-X github.com/fdddf/openproxy/pkg/version.Commit=${COMMIT} -X github.com/fdddf/openproxy/pkg/version.BuildDate=${DATE}" -o /app/openproxy .

# ---------- Runtime Stage ----------
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy executable and frontend assets
COPY --from=go-builder /app/openproxy .
COPY --from=go-builder /app/migrations ./migrations
COPY --from=go-builder /app/statics ./statics
COPY --from=ui-builder /ui/dist ./ui

# Grant execute permission
RUN chmod +x /app/openproxy

EXPOSE 8081

# Start service
# Environment variables can be passed at runtime:
# JWT_SIGN_KEY, DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
CMD ["/app/openproxy"]