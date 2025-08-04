# Soltar VPN

A stateless VPN server written in Go, designed for serverless environments with per-client infrastructure isolation.

## Architecture

- **Server:** Go-based VPN server (stateless) + Webapp registration interface
- **K/V Store:** Redis (each client gets a unique, isolated Redis instance as part of their infrastructure)
- **Client:** macOS Swift client + Linux Go client (GitHub releases only)
- **Infrastructure:** Each client gets a unique environment (VPN, Redis, etc.)
- **Authentication:** Email-based OTP with JWT tokens

## Key Features

- Stateless server design for serverless deployment
- Per-client infrastructure (VPN, Redis, etc.)
- Redis for all persistent storage (per client)
- Email OTP authentication
- JWT token-based sessions
- Zero client logging for privacy
- **Integrated webapp registration interface**

## Quickstart

### Prerequisites

- Docker and Docker Compose
- Go 1.21+
- Redis (for local development)

### Local Development

1. **Clone and setup:**
   ```bash
   git clone <repository>
   cd soltar
   ```

2. **Start Redis server:**
   ```bash
   docker run -p 6379:6379 redis:7-alpine
   ```

3. **Start the VPN server:**
   ```bash
   docker compose up -d
   ```

4. **Access the webapp:**
   ```bash
   # Open in browser
   http://localhost:8080
   ```

5. **Test the API:**
   ```bash
   curl -X POST http://localhost:8080/register \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com"}'
   ```

## Infrastructure and Redis

- **Each client receives a unique, isolated Redis instance** as part of their infrastructure deployment.
- All client data, environment, and infrastructure are stored in that client's Redis.
- This ensures complete isolation between clients.

## Security

- Zero client activity logging
- Per-client Redis for strong isolation
- JWT tokens for session management
- OTP expiration after 5 minutes

## Deployment

### Fly.io Deployment

Each client gets their own isolated deployment with both VPN server and webapp:

```bash
# Deploy a new client instance
./deploy-soltar.sh client-name
```

This creates:
- Dedicated Fly.io app
- Isolated Redis instance
- Unique VPN infrastructure
- **Webapp registration interface**
- Separate domain/subdomain

### Services Deployed

- ✅ **VPN Server API** (Go) - Handles all VPN operations
- ✅ **Webapp Registration** (HTML/JS) - Client registration interface
- ✅ **Redis KV Store** - Persistent storage
- ✅ **Health Check Endpoint** - Monitoring

## API Endpoints

The VPN server provides the following API endpoints:

- `GET /health` - Health check
- `POST /register` - Register with email
- `POST /verify` - Verify OTP
- `POST /connect` - Connect to VPN
- `GET /config` - Get VPN configuration
- `POST /infrastructure` - Update infrastructure
- `GET /infrastructure` - Get infrastructure
- `GET /debug` - Debug storage (development)
- `GET /debug/{key}` - Debug specific key (development)

All endpoints return JSON responses and support CORS.

## Client Distribution

### GitHub Releases

Clients are distributed via GitHub releases only:

- **macOS Client**: Swift-based, distributed via GitHub releases
- **Linux Client**: Go-based, distributed via GitHub releases

### Building Clients

**macOS Client:**
```bash
cd client
swift build -c release
```

**Linux Client:**
```bash
cd client-linux
go build -o soltar-client main.go
```

#### Cross-Platform macOS Client Building

**Note:** The macOS client uses macOS-specific frameworks (`Network`, `Security`) and cannot be built on Linux. The macOS client must be built on a macOS system.

**On macOS:**
```bash
cd client
swift build -c release
# Binary will be available at: .build/release/soltar-vpn
```

**Prerequisites for macOS Client:**
- macOS system with Swift 5.9+
- macOS 13+ target platform
- Network and Security frameworks (included with macOS)

## Development

### Testing

```bash
# Run Go tests
cd cmd/worker
go test -v

# Test the API
curl http://localhost:8080/health

# Test the webapp
curl http://localhost:8080/
```

### Debug Client

For easy testing of deployments, use the debug client:

```bash
# Check server health
./debug-client.sh health

# Register a new client
./debug-client.sh register

# Verify OTP (check server logs for OTP)
./debug-client.sh verify 123456

# Test connection (requires authentication)
./debug-client.sh connect

# Get VPN config (requires authentication)
./debug-client.sh config

# Test webapp interface
./debug-client.sh webapp

# Show current status
./debug-client.sh status

# Run all tests
./debug-client.sh all

# Test remote deployment
SOLTAR_SERVER_URL=https://my-app.fly.dev ./debug-client.sh all
```

The debug client stores authentication tokens between commands, making it easy to test the full workflow.

## License

MIT License 