#!/bin/bash

# Soltar VPN System Deployment Script
# Deploys both VPN server API and webapp registration to Fly.io

set -e

if [ -z "$1" ]; then
    echo "❌ Usage: $0 <client-name>"
    echo "   Example: $0 myclient"
    exit 1
fi

CLIENT_NAME="$1"
APP_NAME="soltar-vpn-${CLIENT_NAME}"

echo "🚀 Deploying Soltar VPN System: $CLIENT_NAME"
echo "📱 App name: $APP_NAME"
echo "🌐 Services: VPN Server API + Webapp Registration"
echo ""

# Create Fly.io app
echo "📋 Creating Fly.io app..."
fly apps create "$APP_NAME" --org personal

# Set secrets
echo "🔐 Setting secrets..."
fly secrets set \
    JWT_SECRET="$(openssl rand -hex 32)" \
    REDIS_URL="redis://localhost:6379" \
    --app "$APP_NAME"

# Deploy the application
echo "🚀 Deploying application..."
fly deploy --app "$APP_NAME"

echo ""
echo "✅ Deployment complete!"
echo "🌐 Webapp: https://$APP_NAME.fly.dev"
echo "🔌 API Endpoints:"
echo "   Health: https://$APP_NAME.fly.dev/health"
echo "   Register: https://$APP_NAME.fly.dev/register"
echo "   Verify: https://$APP_NAME.fly.dev/verify"
echo "   Connect: https://$APP_NAME.fly.dev/connect"
echo "   Config: https://$APP_NAME.fly.dev/config"
echo ""
echo "🧪 Test with debug client:"
echo "   SOLTAR_SERVER_URL=https://$APP_NAME.fly.dev ./debug-client.sh all" 