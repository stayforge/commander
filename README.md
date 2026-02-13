# Stayforge Commander

[![Go Version](https://img.shields.io/github/go-mod/go-version/stayforge/commander?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-008ECF?style=for-the-badge&logo=gin&logoColor=white)](https://gin-gonic.com/)
[![GitHub License](https://img.shields.io/github/license/stayforge/commander?style=for-the-badge)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/stayforge/commander?style=for-the-badge)](https://github.com/stayforge/commander/stargazers)
[![Codecov](https://img.shields.io/codecov/c/github/stayforge/commander?style=for-the-badge&logo=codecov)](https://codecov.io/gh/stayforge/commander)

A high-performance Go KV storage abstraction service for Stayforge edge devices. Supports **BBolt**, **MongoDB**, and **Redis** backends behind a unified interface. Designed for maximum portability, it can be deployed in any Docker-compatible environment -- from public clouds to local edge servers.

## Architecture

Commander provides a pluggable storage layer through the `KV` interface:

```text
                  +-----------+
                  | Commander |
                  |  (Gin)    |
                  +-----+-----+
                        |
              +---------+---------+
              |    KV Interface   |
              | Get/Set/Delete/   |
              | Exists/Ping       |
              +---------+---------+
              |         |         |
         +----+--+ +----+---+ +--+----+
         | BBolt | | MongoDB| | Redis |
         +-------+ +--------+ +-------+
```

| Backend | Best For | Data Model |
|---------|----------|------------|
| **BBolt** (default) | Edge devices, single-node, zero config | Namespace = DB file, Collection = bucket |
| **MongoDB** | Cloud, distributed, complex queries | Namespace = database, Collection = collection |
| **Redis** | High-performance caching, clustering | Key = `namespace:collection:key` |

## Project Structure

```text
.
├── cmd/
│   ├── server/
│   │   └── main.go                 # Application entry point
│   ├── fix_device/                 # Utility: fix device records
│   └── query_card/                 # Utility: query card records
├── internal/
│   ├── config/                     # Configuration management
│   ├── kv/                         # KV interface definition
│   ├── database/
│   │   ├── factory.go              # Backend factory (NewKV)
│   │   ├── bbolt/                  # BBolt implementation
│   │   ├── mongodb/                # MongoDB implementation
│   │   └── redis/                  # Redis implementation
│   ├── models/                     # Data models
│   ├── handlers/                   # HTTP handlers
│   └── services/                   # Business logic
├── Dockerfile                      # Multi-stage build (distroless)
├── docker-compose.yml              # Production deployment
├── docker-compose.dev.yml          # Development (with optional Redis/MongoDB)
├── .env.example                    # Environment configuration template
└── go.mod
```

## Quick Start

### Prerequisites

- Go 1.25.5+
- (Optional) MongoDB or Redis if using those backends

### Installation

```bash
git clone https://github.com/stayforge/commander.git
cd commander
go mod download
```

### Configure

```bash
cp .env.example .env
# Edit .env -- choose your database backend (bbolt/mongodb/redis)
```

### Build & Run

```bash
go build -o bin/server ./cmd/server
./bin/server
```

Or run directly:

```bash
go run cmd/server/main.go
```

The server starts on port `8080` by default. Verify with:

```bash
curl http://localhost:8080/health
```

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE` | No | `bbolt` | Storage backend: `bbolt`, `mongodb`, `redis` |
| `SERVER_PORT` | No | `8080` | HTTP server port |
| `ENVIRONMENT` | No | `STANDARD` | `STANDARD` or `PRODUCTION` (enables Gin release mode) |
| `DATA_PATH` | For bbolt | `/var/lib/stayforge/commander` | BBolt data directory |
| `MONGODB_URI` | For mongodb | - | MongoDB connection string |
| `REDIS_URI` | For redis | - | Redis connection URI |

## API Endpoints

### Health Check

**GET** `/health`

```json
{
  "status": "healthy",
  "environment": "STANDARD",
  "message": "Commander service is running",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

### Root

**GET** `/`

```json
{
  "message": "Welcome to Commander API",
  "version": "dev"
}
```

### Card Verification (MongoDB backend only)

**POST** `/api/v1/namespace/:namespace`

Standard card verification endpoint.

| Source | Parameter | Description |
|--------|-----------|-------------|
| Header | `X-Device-SN` | Device serial number |
| Body | plain text | Card number |

- **204** -- Card is valid, device is authorized
- **400** -- Bad request or card not active/expired
- **403** -- Device not authorized
- **404** -- Card not found

```bash
curl -X POST http://localhost:8080/api/v1/namespace/default \
  -H "X-Device-SN: device-001" \
  -d "ABC123DEF456"
```

**POST** `/api/v1/namespace/:namespace/device/:device_name/vguang`

Legacy vguang-m350 device compatibility endpoint.

- **200** `code=0000` -- Success
- **404** -- Not found

## Docker

### Build & Run

```bash
docker build -t commander .
docker run -p 8080:8080 --env-file .env commander
```

### Docker Compose (Production)

```bash
docker compose up -d
```

### Docker Compose (Development)

```bash
# BBolt only (default)
docker compose -f docker-compose.dev.yml up -d

# With Redis
docker compose -f docker-compose.dev.yml --profile redis up -d

# With MongoDB
docker compose -f docker-compose.dev.yml --profile mongodb up -d
```

## Development

### Run Tests

```bash
go test ./...
```

### Lint

```bash
golangci-lint run
```

### Build for Production

```bash
go build -ldflags="-s -w" -o bin/server ./cmd/server
```

## License

[Business Source License 1.1](LICENSE) -- Converts to Apache-2.0 on 2035-01-01.
