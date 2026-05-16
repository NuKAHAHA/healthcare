# Healthcare API

A secure, production-ready Healthcare Management REST API built with Go, PostgreSQL, and Redis.

---

## Quick Start (3 commands)

```bash
cp .env.example .env          # edit with your secrets
make deps                     # download Go dependencies
make up                       # docker compose up (builds + starts everything)
```

Swagger UI: http://localhost:8080/swagger/index.html
Frontend:   http://localhost:3000

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `APP_ENV` | No | `development` or `production` (default: `production`) |
| `DB_HOST` | Yes | PostgreSQL host |
| `DB_PORT` | No | PostgreSQL port (default: 5432) |
| `DB_USER` | Yes | PostgreSQL username |
| `DB_PASSWORD` | **Yes** | PostgreSQL password — **use strong random value** |
| `DB_NAME` | Yes | Database name |
| `DB_SSL_MODE` | No | `require` (production) or `disable` (local dev) |
| `REDIS_ADDR` | Yes | Redis address (host:port) |
| `REDIS_PASSWORD` | Yes | Redis password |
| `JWT_SECRET` | **Yes** | HS256 signing secret — **min 32 chars, use 64-byte random** |
| `JWT_ACCESS_EXPIRE_MIN` | No | Access token TTL in minutes (default: 5) |
| `JWT_REFRESH_EXPIRE_DAY` | No | Refresh token TTL in days (default: 7) |
| `RATE_LIMIT_GLOBAL_RPM` | No | Global rate limit requests/minute per IP (default: 100) |
| `LOGIN_MAX_ATTEMPTS` | No | Failed logins before block (default: 5) |
| `ALLOWED_ORIGINS` | No | Comma-separated CORS origins |
| `LOG_LEVEL` | No | `debug`, `info`, `warn`, `error` |

Generate secrets:
```bash
openssl rand -hex 64   # JWT_SECRET
openssl rand -hex 32   # DB_PASSWORD, REDIS_PASSWORD
```

---

## API Endpoints

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/auth/login` | No | — | Authenticate, get access token |
| POST | `/auth/register` | JWT | admin | Create new user |
| POST | `/auth/refresh` | Cookie | — | Rotate refresh token |
| POST | `/auth/logout` | JWT | any | Revoke tokens |
| GET | `/patients` | JWT | registrar, admin | List patients |
| POST | `/patients` | JWT | registrar, admin | Register patient |
| GET | `/patients/:id` | JWT | any (doctor = own patients) | Get patient |
| GET | `/appointments` | JWT | any (doctor = own) | List appointments |
| POST | `/appointments` | JWT | registrar, admin | Create appointment |
| POST | `/treatments` | JWT | doctor | Add treatment |
| GET | `/reports/:id` | JWT | doctor, admin | Get appointment report |
| GET | `/audit-logs/:id` | JWT | admin | Get patient audit logs |
| GET | `/health` | No | — | Service health check |
| GET | `/swagger/*` | No | — | Swagger UI |

---

## Default Credentials (after seed)

> **WARNING**: Change all credentials immediately. Never use seed data in production.

| Role | Email | Password |
|---|---|---|
| Admin | `admin@healthcare.local` | _Random — printed during seed_ |
| Registrar | `registrar1@healthcare.local` | `Registrar@seed2024` |
| Doctor | `doctor1@healthcare.local` | `Doctor@seed2024` |

---

## Development Setup

```bash
# 1. Start only infrastructure (DB + Redis)
docker compose up postgres redis -d

# 2. Set environment
cp .env.example .env
# Edit .env: set DB_SSL_MODE=disable, APP_ENV=development

# 3. Apply migrations
make db-migrate

# 4. Run seed (creates 25 users, 200 patients, 300 appointments, 250 treatments)
make db-seed

# 5. Start API
make run
```

---

## Running the Seed

```bash
APP_ENV=development go run ./seed/...
```

The seed:
- Is **blocked** unless `APP_ENV=development`
- Generates a **random admin password** and prints it once to stdout
- Creates 25 users, 200 patients, 300 appointments, 250 treatments
- Uses bulk inserts for performance

---

## Security Features

- **bcrypt** (cost 12) password hashing
- **JWT** with `iss`, `aud`, `jti` claims; 5-minute access token TTL
- **Token blacklist** — revoked JTIs stored in Redis on logout
- **Refresh token rotation** — reuse detection revokes all sessions
- **Redis-backed rate limiting** — global (100 req/min/IP) + login brute-force (5 attempts/15 min)
- **httpOnly cookies** for refresh tokens
- **Security headers** on all responses (HSTS, X-Frame-Options, CSP, etc.)
- **CORS** with configurable allowed origins
- **Parameterized queries** throughout — no SQL injection possible
- **Audit logs** for all critical actions (login, registration, patient access, treatments, etc.)
- **RBAC + object-level authorization** for all protected endpoints
- **No secrets in code or git** — `.env` is gitignored

---

## Make Targets

```bash
make up           # Start with Docker Compose
make down         # Stop containers
make build        # Compile binary
make run          # Build and run
make deps         # Download dependencies
make swagger      # Regenerate Swagger docs
make db-seed      # Run development seed
make test         # Run unit tests
make test-race    # Tests with race detector
make security     # govulncheck + gosec
make fmt          # Format code
make lint         # golangci-lint
```
