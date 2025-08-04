# Soltar VPN Client (Linux)

A command-line VPN client for Linux systems that connects to the Soltar VPN service.

## Features

- 🔐 Email-based OTP authentication
- 🆔 Unique client ID generation
- 🔑 JWT token-based session management
- 🌐 VPN server configuration retrieval
- 📊 Connection status monitoring

## Prerequisites

- Go 1.21 or later
- Soltar VPN server running locally (or accessible via network)

## Building

```bash
# Build the client
go build -o soltar-client main.go

# Or build for release
go build -ldflags="-s -w" -o soltar-client main.go
```

## Usage

### First-time setup

1. Start the Soltar VPN server:
   ```bash
   cd ../
   sudo docker compose up -d
   ```

2. Run the client:
   ```bash
   ./soltar-client
   ```

3. Follow the prompts:
   - Enter your email address
   - Check server logs for OTP (printed to console)
   - Enter the OTP when prompted

### Using stored credentials

Set environment variables to skip registration:

```bash
export SOLTAR_CLIENT_ID="your-client-id"
export SOLTAR_TOKEN="your-jwt-token"
./soltar-client
```

## API Endpoints Tested

The client tests the following endpoints:

- `POST /register` - Register with email
- `POST /verify` - Verify OTP and get credentials
- `POST /connect` - Test connection with JWT token
- `GET /config` - Retrieve VPN configuration

## Development

### Testing

```bash
# Test against local server
go run main.go

# Test with specific server
API_BASE=http://your-server:8080 go run main.go
```

### Debugging

The client provides detailed output for each API call, making it easy to debug connection issues.

## Architecture

- **Authentication**: Email OTP → JWT token
- **Session Management**: JWT tokens for API calls
- **Configuration**: Retrieves VPN server details from API
- **Error Handling**: Graceful error reporting and recovery

## Security

- No credentials stored locally (unless explicitly set as environment variables)
- JWT tokens used for session management
- HTTPS recommended for production use
- OTP expiration after 5 minutes

## License

Same as the main Soltar project. 