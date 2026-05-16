# ── Stage 1: Build ───────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s -extldflags '-static'" \
      -o /app/healthcare-api \
      ./cmd

# ── Stage 2: Final image ─────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S appgroup && \
    adduser  -S appuser -G appgroup -u 1001

WORKDIR /app

COPY --from=builder /app/healthcare-api .
COPY --chown=appuser:appgroup migrations/ ./migrations/

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./healthcare-api"]
