# OpenProxy

A flexible API proxy server for OpenAI-compatible APIs. Written in Go with a Vue.js admin interface.

## Features

- Multiple provider support (OpenAI, Anthropic, Gemini, and custom providers)
- API key management and quota tracking
- Request logging and analytics
- User authentication with JWT
- Admin dashboard for management
- Rate limiting per API key
- Health monitoring for providers

## Quick Start

The binary is self-contained: the admin UI, the SQL migrations, and the static
assets are compiled into it, and it defaults to a SQLite file in the working
directory. No config file, no database server.

```bash
cd src && make all-in-one   # builds the UI, then ../bin/openproxy
../bin/openproxy            # creates ./openproxy.db and listens on :8081
```

Open <http://localhost:8081>. On first start there is no account yet, so the UI
shows a setup page where you create the administrator.

## Prerequisites

- Go 1.24 or later
- Node.js 18+ (only to build the admin UI)
- PostgreSQL (optional — only if you choose the `postgres` driver)

## Installation

### From Source

```bash
git clone https://github.com/fdddf/openproxy.git
cd openproxy/src
make all-in-one
```

`make build` alone produces a binary without the UI; `make ui` builds the UI
into `src/internal/web/dist`, which `go:embed` picks up.

### Using Docker

```bash
docker build -t openproxy .
docker run -p 8081:8081 -v openproxy-data:/data openproxy
```

## Configuration

### Environment Variables

Environment variables override the config file. Everything has a default;
nothing below is required to start.

| Variable | Description | Default |
|----------|-------------|---------|
| `LISTEN_ADDRESS` | Address to listen on | `:8081` |
| `JWT_SIGN_KEY` | Secret key for JWT signing | generated and persisted on first start |
| `DB_DRIVER` | `sqlite` or `postgres` | `sqlite` |
| `DB_PATH` | SQLite database file | `./openproxy.db` |
| `DB_HOST` | Postgres host | — |
| `DB_PORT` | Postgres port | `5432` |
| `DB_USER` | Postgres user | — |
| `DB_PASSWORD` | Postgres password | — |
| `DB_NAME` | Postgres database name | — |
| `DB_SSL_MODE` | Postgres SSL mode | `disable` |

### Config File

Optional. See `src/config.example.yaml` for every setting; copy it to
`src/config.yaml` to override defaults, or pass `--config /path/to/config.yaml`.

## Database

Migrations are embedded in the binary and applied automatically at startup, for
both drivers. `migrations/sqlite` and `migrations/postgres` hold one set each.

**SQLite (default).** Nothing to set up. The file is created on first start,
with WAL enabled.

**Postgres.** Create the database, then point the config at it:

```yaml
database:
  driver: postgres
  host: localhost
  port: 5432
  user: gptproxy
  password: your_password
  name: gptproxy
```

Existing Postgres installs upgrade in place; their migration history is
unchanged.

## Usage

### Start the Server

```bash
./bin/openproxy
```

### API Endpoints

#### Chat Completions

```bash
curl -X POST http://localhost:8081/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer your-api-key' \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'
```

#### List Models

```bash
curl http://localhost:8081/v1/models \
  -H 'Authorization: Bearer your-api-key'
```

### Admin Interface

Access the admin dashboard at `http://localhost:8081`.

There are no default credentials. The first time you open the UI on a fresh
database it redirects to `/setup`, where you create the administrator account.

Accounts are not self-service: only a super user can create further users, and
only super users can manage providers, models, and settings.

## Development

### Backend

```bash
cd src
go run .
```

### Frontend

```bash
cd ui
npm install
npm run dev
```

### Build Frontend

```bash
cd ui
npm run build
```

### Request Logging

Proxied requests are recorded for the dashboard. Because that includes prompt and
response bodies, they are truncated (64 KB by default) and pruned after 30 days.
Tune or disable it under `request_log` in the config file.

## Security

- **Environment Variables**: Never commit sensitive data to version control
- **API Keys**: Generate secure, random API keys
- **JWT Secrets**: A key is generated on first start; set `JWT_SIGN_KEY` to pin your own
- **Database**: Use SSL in production environments
- **Exposure**: The proxy has no transport security of its own. Put it behind TLS
  before exposing it beyond localhost.

## License

[MIT](LICENSE)