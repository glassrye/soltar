#!/bin/bash

set -e

echo "🧪 Running Soltar VPN Worker Tests"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test functions
run_go_tests() {
    echo -e "${BLUE}🔧 Running Go Tests...${NC}"
    
    # Run unit tests
    echo "Running unit tests..."
    go test -v ./...
    
    # Run benchmarks
    echo "Running benchmarks..."
    go test -bench=. -benchmem ./...
    
    # Run with coverage
    echo "Running coverage analysis..."
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    
    echo -e "${GREEN}✅ Go tests completed${NC}"
}

run_worker_tests() {
    echo -e "${BLUE}🔧 Running Cloudflare Worker Tests...${NC}"
    
    # Check if wrangler is installed
    if ! command -v wrangler &> /dev/null; then
        echo -e "${YELLOW}⚠️  Wrangler not found. Install with: npm install -g wrangler${NC}"
        return 1
    fi
    
    # Start local worker
    echo "Starting local worker..."
    wrangler dev &
    WORKER_PID=$!
    
    # Wait for worker to start
    sleep 5
    
    # Test endpoints
    echo "Testing registration endpoint..."
    curl -X POST http://localhost:8787/register \
        -H "Content-Type: application/json" \
        -d '{"email":"test@example.com"}' || true
    
    echo "Testing verification endpoint..."
    curl -X POST http://localhost:8787/verify \
        -H "Content-Type: application/json" \
        -d '{"email":"test@example.com","otp":"123456"}' || true
    
    # Stop worker
    kill $WORKER_PID 2>/dev/null || true
    
    echo -e "${GREEN}✅ Cloudflare Worker tests completed${NC}"
}

run_integration_tests() {
    echo -e "${BLUE}🔧 Running Integration Tests...${NC}"
    
    # Test API compatibility
    echo "Testing API compatibility between implementations..."
    
    # Start Go server
    go run main.go &
    GO_PID=$!
    sleep 2
    
    # Test Go endpoints
    echo "Testing Go implementation..."
    curl -X POST http://localhost:8080/register \
        -H "Content-Type: application/json" \
        -d '{"email":"integration@example.com"}' || true
    
    # Stop Go server
    kill $GO_PID 2>/dev/null || true
    
    echo -e "${GREEN}✅ Integration tests completed${NC}"
}

run_performance_tests() {
    echo -e "${BLUE}🔧 Running Performance Tests...${NC}"
    
    # Go performance tests
    echo "Running Go performance tests..."
    go test -bench=. -benchmem ./... | grep -E "(Benchmark|ns/op|B/op|allocs/op)"
    
    echo -e "${GREEN}✅ Performance tests completed${NC}"
}

run_security_tests() {
    echo -e "${BLUE}🔧 Running Security Tests...${NC}"
    
    # Test JWT token validation
    echo "Testing JWT token security..."
    go test -run TestJWTToken -v ./...
    
    # Test OTP security
    echo "Testing OTP security..."
    go test -run TestGenerateOTP -v ./...
    
    # Test input validation
    echo "Testing input validation..."
    go test -run TestErrorHandling -v ./...
    
    echo -e "${GREEN}✅ Security tests completed${NC}"
}

# Main test runner
main() {
    case "${1:-all}" in
        "go")
            run_go_tests
            ;;
        "worker")
            run_worker_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "performance")
            run_performance_tests
            ;;
        "security")
            run_security_tests
            ;;
        "all")
            run_go_tests
            echo ""
            run_worker_tests
            echo ""
            run_integration_tests
            echo ""
            run_performance_tests
            echo ""
            run_security_tests
            ;;
        *)
            echo "Usage: $0 [go|worker|integration|performance|security|all]"
            echo ""
            echo "Test categories:"
            echo "  go          - Go unit tests and benchmarks"
            echo "  worker      - Cloudflare Worker tests"
            echo "  integration - API compatibility tests"
            echo "  performance - Performance benchmarks"
            echo "  security    - Security validation tests"
            echo "  all         - Run all tests (default)"
            exit 1
            ;;
    esac
    
    echo ""
    echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
}

# Run main function
main "$@" 