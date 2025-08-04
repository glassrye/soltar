# Soltar VPN Client - macOS

Native macOS VPN client for the Soltar VPN system with email-based OTP authentication.

## Features

- ✅ Native macOS Swift application
- ✅ Email-based registration and OTP authentication
- ✅ Environment-aware VPN connections
- ✅ Infrastructure tracking support
- ✅ Command-line interface
- ✅ Secure JWT token management

## Prerequisites

- macOS 13.0 or later
- Xcode 15.0 or later
- Swift 5.9 or later
- Git for version control

## Building the Client

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/soltar.git
cd soltar/client
```

### 2. Build for Development

```bash
# Build in debug mode
swift build

# Run the client
.build/debug/soltar-vpn
```

### 3. Build for Release

```bash
# Build optimized release version
swift build -c release

# Run the release version
.build/release/soltar-vpn
```

### 4. Build Universal Binary

```bash
# Build for both Intel and Apple Silicon
swift build -c release --triple x86_64-apple-macosx
swift build -c release --triple arm64-apple-macosx

# Create universal binary
lipo -create \
  .build/release/soltar-vpn \
  .build/release/soltar-vpn \
  -output soltar-vpn-universal
```

## Development Workflow

### 1. Local Development

```bash
# Start development build
swift build

# Run with hot reload (if using Xcode)
open Package.swift
```

### 2. Testing

```bash
# Run tests
swift test

# Run specific test
swift test --filter TestName
```

### 3. Code Formatting

```bash
# Format code (requires swiftformat)
swiftformat .

# Check formatting without changes
swiftformat --lint .
```

## Release Process

### 1. Version Management

Update version in `Package.swift`:

```swift
let package = Package(
    name: "SoltarVPNClient",
    version: "1.0.0", // Update this
    platforms: [
        .macOS(.v13)
    ],
    // ... rest of configuration
)
```

### 2. Pre-Release Checklist

- [ ] Update version number
- [ ] Update CHANGELOG.md
- [ ] Test all functionality
- [ ] Build release version
- [ ] Test on different macOS versions
- [ ] Update documentation

### 3. Creating a Release

```bash
# 1. Update version and commit
git add .
git commit -m "Release v1.0.0"
git tag v1.0.0

# 2. Build release artifacts
./scripts/build-release.sh

# 3. Push to GitHub
git push origin main
git push origin v1.0.0
```

### 4. GitHub Release

1. Go to GitHub repository
2. Click "Releases" → "Create a new release"
3. Tag: `v1.0.0`
4. Title: `Soltar VPN Client v1.0.0`
5. Description: Include changelog and features
6. Upload artifacts:
   - `soltar-vpn-universal` (Universal binary)
   - `soltar-vpn-x86_64` (Intel binary)
   - `soltar-vpn-arm64` (Apple Silicon binary)

## Build Scripts

### Release Build Script

Create `scripts/build-release.sh`:

```bash
#!/bin/bash

set -e

VERSION=${1:-$(git describe --tags --abbrev=0)}
BUILD_DIR="builds/v${VERSION}"

echo "🚀 Building Soltar VPN Client v${VERSION}"

# Create build directory
mkdir -p ${BUILD_DIR}

# Build Intel version
echo "🔨 Building Intel version..."
swift build -c release --triple x86_64-apple-macosx
cp .build/release/soltar-vpn ${BUILD_DIR}/soltar-vpn-x86_64

# Build Apple Silicon version
echo "🔨 Building Apple Silicon version..."
swift build -c release --triple arm64-apple-macosx
cp .build/release/soltar-vpn ${BUILD_DIR}/soltar-vpn-arm64

# Create universal binary
echo "🔗 Creating universal binary..."
lipo -create \
  ${BUILD_DIR}/soltar-vpn-x86_64 \
  ${BUILD_DIR}/soltar-vpn-arm64 \
  -output ${BUILD_DIR}/soltar-vpn-universal

# Create checksums
echo "🔍 Generating checksums..."
cd ${BUILD_DIR}
shasum -a 256 soltar-vpn-universal > soltar-vpn-universal.sha256
shasum -a 256 soltar-vpn-x86_64 > soltar-vpn-x86_64.sha256
shasum -a 256 soltar-vpn-arm64 > soltar-vpn-arm64.sha256

echo "✅ Build complete! Artifacts in ${BUILD_DIR}/"
ls -la ${BUILD_DIR}/
```

### Make it executable:

```bash
chmod +x scripts/build-release.sh
```

## Installation

### From Source

```bash
# Clone and build
git clone https://github.com/your-username/soltar.git
cd soltar/client
swift build -c release

# Install to system
sudo cp .build/release/soltar-vpn /usr/local/bin/
```

### From Release

```bash
# Download latest release
curl -L -o soltar-vpn https://github.com/your-username/soltar/releases/latest/download/soltar-vpn-universal

# Make executable
chmod +x soltar-vpn

# Move to PATH
sudo mv soltar-vpn /usr/local/bin/
```

## Configuration

### Environment Variables

```bash
# Set worker URL (optional, defaults to production)
export SOLTAR_WORKER_URL="https://your-worker.workers.dev"

# Set debug mode (optional)
export SOLTAR_DEBUG="true"
```

### Configuration File

Create `~/.soltar/config.json`:

```json
{
  "worker_url": "https://your-worker.workers.dev",
  "debug": false,
  "log_level": "info"
}
```

## Usage

### Basic Usage

```bash
# Start the client
soltar-vpn

# Follow the prompts:
# 1. Enter your email
# 2. Check email for OTP
# 3. Enter OTP
# 4. VPN connects automatically
```

### Advanced Usage

```bash
# Run with debug output
SOLTAR_DEBUG=true soltar-vpn

# Specify custom worker URL
SOLTAR_WORKER_URL=https://staging-worker.workers.dev soltar-vpn
```

## Troubleshooting

### Common Issues

1. **Build fails with Swift version error**
   ```bash
   # Update Swift
   xcode-select --install
   ```

2. **Permission denied when running**
   ```bash
   # Fix permissions
   chmod +x soltar-vpn
   ```

3. **VPN connection fails**
   ```bash
   # Check network connectivity
   ping vpn.soltar.com
   
   # Check worker status
   curl https://your-worker.workers.dev/health
   ```

### Debug Mode

```bash
# Enable debug logging
export SOLTAR_DEBUG=true
soltar-vpn
```

## Development

### Project Structure

```
client/
├── Package.swift          # Swift package configuration
├── main.swift            # Main application entry point
├── Sources/              # Source code (if using package structure)
├── Tests/                # Test files
├── scripts/              # Build and deployment scripts
└── docs/                 # Documentation
```

### Adding Features

1. Create feature branch
2. Implement changes
3. Add tests
4. Update documentation
5. Create pull request

### Code Style

- Follow Swift style guidelines
- Use meaningful variable names
- Add comments for complex logic
- Include error handling

## Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature-name`
3. Commit changes: `git commit -am 'Add feature'`
4. Push branch: `git push origin feature-name`
5. Create pull request

## License

MIT License - see LICENSE file for details.

---

*Beep boop beep* 🤖 