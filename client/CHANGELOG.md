# Changelog

All notable changes to the Soltar VPN Client will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release
- Email-based OTP authentication
- Environment-aware VPN connections
- Infrastructure tracking support
- Command-line interface
- Universal binary support (Intel + Apple Silicon)

### Changed

### Deprecated

### Removed

### Fixed

### Security

## [1.0.0] - 2024-01-01

### Added
- Initial release of Soltar VPN Client
- Email-based registration and OTP authentication
- Environment-aware VPN connections with unique server per client
- Infrastructure tracking and management
- Native macOS Swift application
- Universal binary support for Intel and Apple Silicon Macs
- Command-line interface with interactive prompts
- JWT token-based authentication
- Debug mode and logging support
- Configuration file support
- Checksum verification for downloads

### Features
- **Authentication**: Email-based OTP system with 5-minute expiry
- **Environment Isolation**: Each client gets unique VPN server and infrastructure
- **Infrastructure Tracking**: Monitor VPN instances, load balancers, databases, storage
- **Cross-Platform**: Universal binary works on Intel and Apple Silicon Macs
- **Security**: No client logging, secure JWT tokens, isolated environments
- **Usability**: Simple command-line interface with clear prompts

### System Requirements
- macOS 13.0 or later
- Intel or Apple Silicon Mac
- Network connectivity for OTP delivery and VPN connection

### Installation
```bash
# Download latest release
curl -L -o soltar-vpn https://github.com/your-username/soltar/releases/latest/download/soltar-vpn-universal
chmod +x soltar-vpn
sudo mv soltar-vpn /usr/local/bin/
```

### Usage
```bash
# Start the client
soltar-vpn

# Follow prompts:
# 1. Enter email address
# 2. Check email for OTP
# 3. Enter OTP
# 4. VPN connects automatically
```

---

## Release Process

### Creating a Release

1. **Update Version**
   ```bash
   # Update version in Package.swift
   # Update this CHANGELOG.md
   # Update any other version references
   ```

2. **Commit and Tag**
   ```bash
   git add .
   git commit -m "Release v1.0.1"
   git tag v1.0.1
   ```

3. **Push to GitHub**
   ```bash
   git push origin main
   git push origin v1.0.1
   ```

4. **Automated Release**
   - GitHub Actions will automatically build and create release
   - Universal binary and checksums will be uploaded
   - Release notes will be generated

### Version Format

- **Major.Minor.Patch** (e.g., 1.0.0)
- **Major**: Breaking changes
- **Minor**: New features, backward compatible
- **Patch**: Bug fixes, backward compatible

### Tagging Convention

- Use semantic versioning: `v1.0.0`
- Always prefix with `v`
- Use lowercase letters
- No spaces in tag names

---

*Beep boop beep* 🤖 