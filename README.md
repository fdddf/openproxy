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

## Prerequisites

- Go 1.24 or later
- PostgreSQL database
- Node.js 18+ (for frontend development)

## Installation

### From Source

```bash
git clone https://github.com/fdddf/openproxy.git
cd openproxy
cd src && go build -o ../bin/openproxy .
```

### Using Docker

```bash
docker build -t openproxy .
```

## Configuration

### Environment Variables

Sensitive configuration should be set via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_SIGN_KEY` | Secret key for JWT signing | (required) |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `gptproxy` |
| `DB_PASSWORD` | Database password | (required) |
| `DB_NAME` | Database name | `gptproxy` |
| `DB_SSL_MODE` | Database SSL mode | `disable` |

### Config File

Create `src/config.yaml`:

```yaml
proxy:
  listen_address: ":8081"
  default_provider: "openai"

database:
  host: "localhost"
  port: 5432
  user: "gptproxy"
  password: "your_password"
  name: "gptproxy"
  ssl_mode: "disable"
  migrations_path: "./migrations"
```

## Database Setup

1. Create a PostgreSQL database:
   ```bash
   createdb gptproxy
   ```

2. Run migrations (automatic on server start, or manually):
   ```bash
   cd src
   migrate -path ./migrations -database "postgres://user:pass@localhost:5432/gptproxy?sslmode=disable" up
   ```

## Usage

### Start the Server

```bash
# Set environment variables
export JWT_SIGN_KEY="your-secret-key"
export DB_PASSWORD="your-db-password"

# Run the server
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

Access the admin dashboard at `http://localhost:8081/admin`. 

Default credentials: `admin` / `admin123`

**Important:** Change the default password immediately after first login!

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

## Security

- **Environment Variables**: Never commit sensitive data to version control
- **API Keys**: Generate secure, random API keys
- **JWT Secrets**: Use strong, unique secrets for JWT signing
- **Database**: Use SSL in production environments

## License

[MIT](LICENSE)