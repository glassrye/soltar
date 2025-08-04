# Soltar VPN Worker

This directory contains the Go-based VPN server implementation for the Soltar VPN system. The server uses Redis for storage and serves both the VPN API and webapp registration interface. **Each client receives a unique, isolated Redis instance as part of their infrastructure deployment.**

## Architecture

### Serverless Architecture with Redis

The worker is designed to be:
- **Stateless**: No local storage, all data in Redis
- **Serverless-ready**: Can run on AWS Lambda, Google Cloud Functions, etc.
- **Scalable**: Each client gets dedicated resources
- **Redis for storage** (one per client in production)
- **Dual service**: VPN API + Webapp registration interface

### Per-Client Infrastructure

- **Each client gets a dedicated Redis instance**
- **Isolated VPN infrastructure** per client
- **Unique environment** for each client
- **Integrated webapp** for client registration

## Why Go + Redis?

- **Go**: Fast, efficient, stateless server
- **Redis**: Fast, reliable key-value storage (one per client)
- **Unique Infrastructure**: Each client gets dedicated resources, including their own Redis
- **Single deployment**: VPN API and webapp in one service

## Architecture Comparison

| Component | Development | Production |
|-----------|-------------|------------|
| **Server** | Go | Go (serverless) |
| **Storage** | Redis | Redis (per client) |
| **Webapp** | Static files | Static files |
| **Infrastructure** | Shared | Unique per client (Redis per client) |

## File Structure

```
cmd/worker/
├── main.go              # Go server with Redis + webapp
├── main_test.go         # Test suite
└── README.md           # This file

webapp/
├── index.html          # Registration interface
└── README.md          # Webapp documentation
```

## Development

### Prerequisites

- Go 1.21+
- Redis (for local development)

### Local Development

1. **Start Redis server:**
   ```bash
   docker run -p 6379:6379 redis:7-alpine
   ```

2. **Run the server:**
   ```bash
   go run main.go
   ```

3. **Test the API:**
   ```bash
   curl http://localhost:8080/health
   ```

4. **Access the webapp:**
   ```bash
   # Open in browser
   http://localhost:8080
   ```

### Environment Variables

- `REDIS_URL`: Redis connection URL (default: `redis://localhost:6379`)
- `JWT_SECRET`: Secret for JWT signing (default: development key)
- `PORT`: HTTP server port (default: `8080`)

## API Endpoints

The worker serves both the VPN API and webapp registration interface:

### VPN API Endpoints
- `GET /health` - Health check
- `POST /register` - Register with email
- `POST /verify` - Verify OTP
- `POST /connect` - Connect to VPN
- `GET /config` - Get VPN configuration
- `POST /infrastructure` - Update infrastructure
- `GET /infrastructure` - Get infrastructure
- `GET /debug` - Debug storage (development)
- `GET /debug/{key}` - Debug specific key (development)

### Webapp
- `GET /` - Registration interface (HTML/JS)

All API endpoints return JSON responses and support CORS.

## Testing

### Run Tests

```bash
go test -v
```

### Test Coverage

```bash
go test -cover
```

### Integration Tests

The test suite includes:
- OTP generation and verification
- Client creation and retrieval
- JWT token validation
- API endpoint testing
- Error handling
- Concurrent access testing

## Production Deployment

### Serverless Platforms

The worker can be deployed to:
- **AWS Lambda** with Redis ElastiCache
- **Google Cloud Functions** with Redis Cloud
- **Azure Functions** with Azure Cache for Redis
- **Fly.io** with Redis (recommended)

### Fly.io Deployment

```bash
# Deploy to Fly.io (includes both VPN API and webapp)
fly deploy

# Or use the deployment script
../deploy-soltar.sh client-name

# Each client gets a dedicated Redis instance as part of their infrastructure
```

## Infrastructure and Redis

### Per-Client Isolation

- **Each client receives a unique, isolated Redis instance** as part of their infrastructure deployment.
- All client data, environment, and infrastructure are stored in that client's Redis.
- This ensures complete isolation and security.

### Data Storage

| Data Type | Key Pattern | TTL |
|-----------|-------------|-----|
| **OTP** | `otp:{email}` | 5 minutes |
| **Client** | `client:{id}` | None |
| **Environment** | `env:{client_id}` | None |
| **Infrastructure** | `infra:{client_id}` | None |

## Security

### Privacy

- **Zero client logging**: No client activity is logged
- **JWT tokens**: Secure session management
- **OTP expiration**: 5-minute TTL for OTP codes

### Isolation

- **Per-client Redis** for strong isolation
- **Unique infrastructure** per client
- **No shared state** between clients

## Performance

### Development
- **Storage**: Redis (local)
- **Latency**: < 1ms
- **Throughput**: High

### Production
- **Storage**: Redis (distributed, per client)
- **Latency**: < 10ms
- **Throughput**: High

## Monitoring

### Health Checks

```bash
curl http://localhost:8080/health
```

### Debug Endpoints

```bash
# List available keys
curl http://localhost:8080/debug

# Get specific key
curl http://localhost:8080/debug/client_123
```

## Development Workflow

1. **Local Development**: Redis + Go server + webapp
2. **Testing**: Comprehensive test suite
3. **Local Validation**: Redis testing
4. **Deploy**: Serverless platform deployment (Redis per client)

## Troubleshooting

### Common Issues

1. **Redis Connection Failed**
   - Check Redis server is running
   - Verify `REDIS_URL` environment variable

2. **OTP Verification Fails**
   - Check OTP expiration (5 minutes)
   - Verify email format in key

3. **JWT Token Invalid**
   - Check `JWT_SECRET` environment variable
   - Verify token expiration

4. **Webapp Not Loading**
   - Check static file serving
   - Verify webapp directory is copied to container

### Debug Mode

Enable debug logging by checking server logs for detailed Redis operations.

## Future Enhancements

- **Redis Cluster**: For high availability
- **Redis Sentinel**: For failover
- **Redis Modules**: For advanced features
- **Monitoring**: Redis metrics and alerts
- **Webapp Enhancements**: Better UI/UX

## Summary

The Soltar VPN worker provides:
- **Stateless design** for serverless deployment
- **Redis storage** for persistence and performance
- **Per-client isolation** for security
- **Scalable architecture** for growth
- **Integrated webapp** for client registration 