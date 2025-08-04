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

# Create release notes
echo "📝 Creating release notes..."
cat > ${BUILD_DIR}/RELEASE_NOTES.md << EOF
# Soltar VPN Client v${VERSION}

## Downloads

- **Universal Binary** (Intel + Apple Silicon): \`soltar-vpn-universal\`
- **Intel Binary**: \`soltar-vpn-x86_64\`
- **Apple Silicon Binary**: \`soltar-vpn-arm64\`

## Installation

\`\`\`bash
# Download and install
curl -L -o soltar-vpn https://github.com/your-username/soltar/releases/download/v${VERSION}/soltar-vpn-universal
chmod +x soltar-vpn
sudo mv soltar-vpn /usr/local/bin/
\`\`\`

## Verification

\`\`\`bash
# Verify checksums
shasum -c soltar-vpn-universal.sha256
\`\`\`

## Changes

- Initial release
- Email-based OTP authentication
- Environment-aware VPN connections
- Infrastructure tracking support

## System Requirements

- macOS 13.0 or later
- Intel or Apple Silicon Mac
EOF

echo "📋 Release notes created: ${BUILD_DIR}/RELEASE_NOTES.md" 