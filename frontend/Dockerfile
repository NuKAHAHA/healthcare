# ── Stage 1: Build Vite app ───────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci --quiet
COPY . .
RUN npm run build

# ── Stage 2: Serve with nginx ─────────────────────────────────
FROM nginx:1.25-alpine

RUN addgroup -S appgroup && adduser -S appuser -G appgroup -u 1001

COPY --from=builder /app/dist       /usr/share/nginx/html
COPY nginx.conf                      /etc/nginx/conf.d/default.conf

# nginx needs root to bind port 80; drop privileges after binding
EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s \
  CMD wget -qO- http://localhost:80/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
