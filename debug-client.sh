#!/bin/bash

# Soltar VPN Debug Client
# A simple client for testing deployments without building full clients

set -e

# Default values
SERVER_URL="${SOLTAR_SERVER_URL:-http://localhost:8080}"
EMAIL="${SOLTAR_EMAIL:-test@example.com}"
TOKEN_FILE="/tmp/soltar_debug_token.txt"
CLIENT_ID_FILE="/tmp/soltar_debug_client_id.txt"

echo "🔒 Soltar VPN Debug Client"
echo "=========================="
echo "Server: $SERVER_URL"
echo "Email: $EMAIL"
echo ""

# Function to make API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ -n "$data" ]; then
        curl -s -X "$method" "$SERVER_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data"
    else
        curl -s -X "$method" "$SERVER_URL$endpoint"
    fi
}

# Function to check server health
check_health() {
    echo "🏥 Checking server health..."
    response=$(api_call "GET" "/health")
    echo "Health response: $response"
    echo ""
}

# Function to register
register() {
    echo "📧 Registering with email: $EMAIL"
    response=$(api_call "POST" "/register" "{\"email\":\"$EMAIL\"}")
    echo "Registration response: $response"
    echo ""
}

# Function to verify OTP
verify_otp() {
    local otp=$1
    echo "🔐 Verifying OTP: $otp"
    response=$(api_call "POST" "/verify" "{\"email\":\"$EMAIL\",\"otp\":\"$otp\"}")
    echo "Verification response: $response"
    echo ""
    
    # Extract client ID and token if successful
    if echo "$response" | grep -q "client_id"; then
        CLIENT_ID=$(echo "$response" | grep -o '"client_id":"[^"]*"' | cut -d'"' -f4)
        TOKEN=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo "✅ Success! Client ID: $CLIENT_ID"
        echo "🔑 Token: ${TOKEN:0:20}..."
        echo ""
        
        # Store for later use
        echo "$TOKEN" > "$TOKEN_FILE"
        echo "$CLIENT_ID" > "$CLIENT_ID_FILE"
        echo "💾 Token and Client ID saved for future commands"
        echo ""
    fi
}

# Function to get stored token
get_stored_token() {
    if [ -f "$TOKEN_FILE" ]; then
        cat "$TOKEN_FILE"
    fi
}

# Function to get stored client ID
get_stored_client_id() {
    if [ -f "$CLIENT_ID_FILE" ]; then
        cat "$CLIENT_ID_FILE"
    fi
}

# Function to test connection
test_connect() {
    TOKEN=$(get_stored_token)
    if [ -z "$TOKEN" ]; then
        echo "❌ No token available. Please verify OTP first."
        return
    fi
    
    echo "🔗 Testing connection..."
    response=$(curl -s -X "POST" "$SERVER_URL/connect" \
        -H "Authorization: Bearer $TOKEN")
    echo "Connection response: $response"
    echo ""
}

# Function to get config
get_config() {
    TOKEN=$(get_stored_token)
    if [ -z "$TOKEN" ]; then
        echo "❌ No token available. Please verify OTP first."
        return
    fi
    
    echo "⚙️  Getting VPN config..."
    response=$(curl -s -X "GET" "$SERVER_URL/config" \
        -H "Authorization: Bearer $TOKEN")
    echo "Config response: $response"
    echo ""
}

# Function to test webapp
test_webapp() {
    echo "🌐 Testing webapp..."
    response=$(curl -s -I "$SERVER_URL/" | head -1)
    echo "Webapp response: $response"
    echo ""
}

# Function to debug storage
debug_storage() {
    echo "🔍 Debugging storage..."
    response=$(api_call "GET" "/debug")
    echo "Debug response: $response"
    echo ""
}

# Function to show status
show_status() {
    echo "📊 Current Status:"
    echo "  Server: $SERVER_URL"
    echo "  Email: $EMAIL"
    
    CLIENT_ID=$(get_stored_client_id)
    if [ -n "$CLIENT_ID" ]; then
        echo "  Client ID: $CLIENT_ID"
    else
        echo "  Client ID: Not authenticated"
    fi
    
    TOKEN=$(get_stored_token)
    if [ -n "$TOKEN" ]; then
        echo "  Token: ${TOKEN:0:20}..."
    else
        echo "  Token: Not available"
    fi
    echo ""
}

# Function to clear stored data
clear_data() {
    rm -f "$TOKEN_FILE" "$CLIENT_ID_FILE"
    echo "🗑️  Cleared stored token and client ID"
    echo ""
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  health     - Check server health"
    echo "  register   - Register with email"
    echo "  verify OTP - Verify OTP code"
    echo "  connect    - Test connection (requires token)"
    echo "  config     - Get VPN config (requires token)"
    echo "  webapp     - Test webapp interface"
    echo "  debug      - Debug storage"
    echo "  status     - Show current status"
    echo "  clear      - Clear stored token/client ID"
    echo "  all        - Run all tests"
    echo ""
    echo "Environment variables:"
    echo "  SOLTAR_SERVER_URL - Server URL (default: http://localhost:8080)"
    echo "  SOLTAR_EMAIL      - Email for testing (default: test@example.com)"
    echo ""
    echo "Examples:"
    echo "  $0 health"
    echo "  $0 register"
    echo "  $0 verify 123456"
    echo "  $0 all"
    echo "  SOLTAR_SERVER_URL=https://my-app.fly.dev $0 all"
}

# Main logic
case "${1:-help}" in
    "health")
        check_health
        ;;
    "register")
        register
        ;;
    "verify")
        if [ -z "$2" ]; then
            echo "❌ Please provide OTP code"
            echo "Usage: $0 verify <OTP>"
            exit 1
        fi
        verify_otp "$2"
        ;;
    "connect")
        test_connect
        ;;
    "config")
        get_config
        ;;
    "webapp")
        test_webapp
        ;;
    "debug")
        debug_storage
        ;;
    "status")
        show_status
        ;;
    "clear")
        clear_data
        ;;
    "all")
        echo "🧪 Running all tests..."
        echo ""
        check_health
        register
        echo "📝 Please check server logs for OTP, then run:"
        echo "   $0 verify <OTP>"
        echo ""
        test_webapp
        debug_storage
        ;;
    "help"|*)
        show_usage
        ;;
esac 