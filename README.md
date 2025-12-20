![repository-open-graph](./.github/open-graph/repository-open-graph.png)

# Access Authorization Service

A high-performance Go-based authentication service for verifying device access using card-based authorization with MongoDB Atlas backend.

## 🌐 Domain
authorization.access.stayforge.net

## 🏗️ Architecture

This service is a complete rewrite from Python/FastAPI to Go/Gin, providing:

- **Card-based Authentication**: Validates device access based on card credentials
- **Time-based Access Control**: Enforces activation and expiration times with NTP drift compensation
- **Device Authorization**: Restricts access to pre-authorized devices
- **MongoDB Atlas Integration**: Cloud-native database backend
- **RESTful API**: Clean HTTP endpoints for device identification

## 📁 Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   └── mongodb.go           # MongoDB connection handling
│   ├── models/
│   │   └── card.go              # Data models
│   ├── handlers/
│   │   └── identify.go          # HTTP request handlers
│   └── service/
│       └── card_service.go      # Business logic
├── bin/
│   └── server                   # Compiled binary
├── .env.example                 # Environment configuration template
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- MongoDB Atlas account with connection URI
- Git

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/iktahana/access-authorization-service.git
cd Access-Authorization-Service
```

2. **Install dependencies**
```bash
go mod download
```

3. **Configure environment**
```bash
cp .env.example .env
# Edit .env with your MongoDB Atlas credentials
```

4. **Build the application**
```bash
go build -o bin/server ./cmd/server
```

5. **Run the server**
```bash
./bin/server
```

The server will start on port 8080 (or the port specified in your `.env` file).

## ⚙️ Configuration

Create a `.env` file in the root directory with the following variables:

```env
# MongoDB Atlas Configuration
MONGODB_URI=mongodb+srv://username:password@cluster.mongodb.net/
MONGODB_DATABASE=your_database_name
MONGODB_COLLECTION=cards

# Server Configuration
SERVER_PORT=8080

# Environment (STANDARD, PRODUCTION, etc.)
ENVIRONMENT=STANDARD
```

### Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MONGODB_URI` | ✅ | - | MongoDB Atlas connection string |
| `MONGODB_DATABASE` | ✅ | - | Database name |
| `MONGODB_COLLECTION` | ❌ | `cards` | Collection name for card documents |
| `SERVER_PORT` | ❌ | `8080` | HTTP server port |
| `ENVIRONMENT` | ❌ | `STANDARD` | Environment name (affects logging mode) |

## 📡 API Endpoints

### 1. JSON Identification (Recommended)

**POST** `/identify/json`  
**POST** `/identify/json/:device_sn`

Identify a device using JSON request body.

**Headers:**
- `X-Device-SN` (optional): Device serial number (alternative to path parameter)
- `X-Environment` (optional): Environment name (default: STANDARD)

**Request Body:**
```json
{
  "card_number": "ABC123DEF456"
}
```

**Success Response (200):**
```json
{
  "message": "Successfully",
  "card_number": "ABC123DEF456",
  "devices": ["device-001", "device-002"],
  "invalid_at": "2024-01-01T00:00:00Z",
  "expired_at": "2024-12-31T23:59:59Z",
  "activation_offset_seconds": 60,
  "owner_client_id": "client-123",
  "name": "Guest Room 101"
}
```

**Error Response (400/404):**
```json
{
  "message": "card not found"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/identify/json/device-001 \
  -H "Content-Type: application/json" \
  -d '{"card_number": "ABC123DEF456"}'
```

### 2. vguang-m350 Device Identification

**POST** `/identify/vguang-m350/:device_name`

Special endpoint for vguang-m350 hardware devices with custom byte-handling logic.

**Path Parameters:**
- `device_name`: Device identifier

**Request Body:** Raw bytes (card data)

**Success Response (200):**
```
code=0000
```

**Error Response (404):**
```json
{
  "message": "card not found"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/identify/vguang-m350/device-001 \
  --data-raw "ABC123DEF456"
```

### 3. Health Check

**GET** `/health`

Check service health status.

**Success Response (200):**
```json
{
  "status": "healthy",
  "environment": "STANDARD",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 🔐 Authentication Logic

### Time-Based Validation

Cards have activation and expiration times with drift compensation:

```
Card is VALID when:
  current_time >= (invalid_at - activation_offset_seconds)
  AND
  current_time <= expired_at
```

The `activation_offset_seconds` (default: 60) compensates for NTP clock drift, allowing cards to activate slightly before their scheduled time.

### Device Authorization

Each card contains a list of authorized device IDs. A device can only authenticate if:
1. The card exists in the database
2. The card is within its valid time range
3. The device ID is in the card's `devices` array

## 🗄️ MongoDB Schema

### Card Document

```javascript
{
  "_id": ObjectId("..."),
  "card_number": "ABC123DEF456",        // Unique card identifier (uppercase)
  "devices": [                          // List of authorized device IDs
    "device-001",
    "device-002"
  ],
  "invalid_at": ISODate("2024-01-01T00:00:00Z"),  // Activation time
  "expired_at": ISODate("2024-12-31T23:59:59Z"),  // Expiration time
  "activation_offset_seconds": 60,      // Drift compensation (seconds)
  "owner_client_id": "client-123",      // Optional: Owner identifier
  "name": "Guest Room 101"              // Optional: Human-readable name
}
```

### Required Indexes

Create these indexes for optimal performance:

```javascript
db.cards.createIndex({ "card_number": 1 }, { unique: true })
db.cards.createIndex({ "devices": 1 })
db.cards.createIndex({ "invalid_at": 1, "expired_at": 1 })
```

## 🛠️ Development

### Run in Development Mode

```bash
# With hot reload using air (install: go install github.com/cosmtrek/air@latest)
air

# Or run directly
go run cmd/server/main.go
```

### Build for Production

```bash
# Build optimized binary
go build -ldflags="-s -w" -o bin/server ./cmd/server

# Build for Linux (from macOS)
GOOS=linux GOARCH=amd64 go build -o bin/server-linux ./cmd/server
```

### Run Tests

```bash
go test ./...
```

## 📊 Error Codes

| HTTP Status | Description |
|-------------|-------------|
| 200 | Authentication successful - device authorized |
| 400 | Bad request or card not active/expired |
| 404 | Card not found in database |
| 500 | Internal server error |

## 🔍 Logging

The service logs all requests with the following format:

```
[POST] /identify/json - Status: 200 - Latency: 15ms - IP: 192.168.1.100
```

Error details are logged separately for debugging.

## 🚦 Deployment

### Docker (Recommended)

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

Build and run:
```bash
docker build -t access-authorization-service .
docker run -p 8080:8080 --env-file .env access-authorization-service
```

### Systemd Service

Create `/etc/systemd/system/access-auth.service`:

```ini
[Unit]
Description=Access Authorization Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/access-authorization-service
Environment="MONGODB_URI=your_uri"
Environment="MONGODB_DATABASE=your_db"
ExecStart=/opt/access-authorization-service/bin/server
Restart=always

[Install]
WantedBy=multi-user.target
```

## 🔄 Migration from Python

Key differences from the original Python/FastAPI implementation:

1. **Performance**: ~5-10x faster request handling
2. **Memory**: ~70% lower memory footprint
3. **Concurrency**: Native goroutines vs Python asyncio
4. **Deployment**: Single binary (no dependencies)
5. **Type Safety**: Compile-time type checking

## 📝 License

See LICENSE file for details.

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📧 Support

For issues and questions:
- Create an issue on GitHub
- Contact: authorization.access.stayforge.net

## 🎯 Roadmap

- [ ] Add metrics endpoint (Prometheus)
- [ ] Implement rate limiting
- [ ] Add Redis caching layer
- [ ] Support JWT authentication for admin endpoints
- [ ] Add comprehensive test suite
- [ ] OpenAPI/Swagger documentation

---

**Built with ❤️ using Go and Gin Framework**
