# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

ARG VERSION=1.0.0
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
    -o /bin/order-service ./cmd/server

# Runtime stage (minimal & secure)
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /bin/order-service /app/order-service

USER appuser:appgroup

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/order-service"]
