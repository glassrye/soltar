package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Mock storage for testing
type MockStorage struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewMockStorage() Storage {
	return &MockStorage{
		data: make(map[string][]byte),
	}
}

func (m *MockStorage) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if data, exists := m.data[key]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

func (m *MockStorage) Put(key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *MockStorage) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

// Test helper functions
func createTestRequest(method, path string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func createAuthRequest(method, path string, token string, body interface{}) *http.Request {
	req := createTestRequest(method, path, body)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// Test OTP generation
func TestGenerateOTP(t *testing.T) {
	otp := generateOTP()

	if len(otp) != 6 {
		t.Errorf("Expected OTP length 6, got %d", len(otp))
	}

	// Test that OTP is numeric
	for _, char := range otp {
		if char < '0' || char > '9' {
			t.Errorf("OTP contains non-numeric character: %c", char)
		}
	}
}

// Test client creation and retrieval
func TestGetOrCreateClient(t *testing.T) {
	storage = NewMockStorage()

	email := "test@example.com"

	// Test creating new client
	clientData := getOrCreateClientWithInfrastructure(email)

	if clientData == nil {
		t.Error("Expected client data to be generated")
	}

	if clientData.ID == "" {
		t.Error("Expected client ID to be generated")
	}

	// Test retrieving existing client
	existingClientData := getOrCreateClientWithInfrastructure(email)

	if existingClientData.ID != clientData.ID {
		t.Error("Expected same client ID for existing client")
	}
}

// Test JWT token generation and validation
func TestJWTToken(t *testing.T) {
	clientID := uuid.New().String()

	// Test token generation
	token := generateToken(clientID)

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Test token validation
	validatedID, err := validateToken(token)
	if err != nil {
		t.Errorf("Expected valid token, got error: %v", err)
	}

	if validatedID != clientID {
		t.Errorf("Expected client ID %s, got %s", clientID, validatedID)
	}

	// Test invalid token
	_, err = validateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}

	// Test expired token
	expiredToken := createExpiredToken(clientID)
	_, err = validateToken(expiredToken)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func createExpiredToken(clientID string) string {
	claims := jwt.RegisteredClaims{
		Subject:   clientID,
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(secret)
	return tokenString
}

// Test registration endpoint
func TestHandleRegister(t *testing.T) {
	storage = NewMockStorage()

	req := createTestRequest("POST", "/register", OTPRequest{
		Email: "test@example.com",
	})

	w := httptest.NewRecorder()
	handleRegister(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "OTP sent to email" {
		t.Errorf("Expected success message, got %s", response["message"])
	}

	// Verify OTP was stored
	otpData, err := storage.Get("otp:test@example.com")
	if err != nil {
		t.Error("Expected OTP to be stored")
	}

	var otpInfo map[string]interface{}
	json.Unmarshal(otpData, &otpInfo)

	if otpInfo["otp"] == "" {
		t.Error("Expected OTP to be generated")
	}
}

// Test OTP verification
func TestHandleVerify(t *testing.T) {
	storage = NewMockStorage()

	email := "test@example.com"
	otp := "123456"

	// Store OTP
	otpData := map[string]interface{}{
		"otp":      otp,
		"expires":  time.Now().Add(5 * time.Minute).Unix(),
		"attempts": 0,
	}
	otpBytes, _ := json.Marshal(otpData)
	storage.Put("otp:test@example.com", otpBytes)

	// Test successful verification
	req := createTestRequest("POST", "/verify", OTPVerify{
		Email: email,
		OTP:   otp,
	})

	w := httptest.NewRecorder()
	handleVerify(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response AuthResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.ClientID == "" {
		t.Error("Expected client ID in response")
	}

	if response.Token == "" {
		t.Error("Expected token in response")
	}

	// Test invalid OTP
	req = createTestRequest("POST", "/verify", OTPVerify{
		Email: email,
		OTP:   "000000",
	})

	w = httptest.NewRecorder()
	handleVerify(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid OTP, got %d", w.Code)
	}
}

// Test connection endpoint
func TestHandleConnect(t *testing.T) {
	storage = NewMockStorage()

	// Create client and token
	clientData := getOrCreateClientWithInfrastructure("test@example.com")
	token := generateToken(clientData.ID)

	// Test successful connection
	req := createAuthRequest("POST", "/connect", token, nil)

	w := httptest.NewRecorder()
	handleConnect(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["status"] != "connected" {
		t.Errorf("Expected connected status, got %v", response["status"])
	}

	// Test unauthorized request
	req = createTestRequest("POST", "/connect", nil)

	w = httptest.NewRecorder()
	handleConnect(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for unauthorized request, got %d", w.Code)
	}
}

// Test config endpoint
func TestHandleConfig(t *testing.T) {
	storage = NewMockStorage()

	// Create client and token
	clientData := getOrCreateClientWithInfrastructure("test@example.com")
	token := generateToken(clientData.ID)

	// Test successful config retrieval
	req := createAuthRequest("GET", "/config", token, nil)

	w := httptest.NewRecorder()
	handleConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var config VPNConfig
	json.Unmarshal(w.Body.Bytes(), &config)

	if config.Server == "" {
		t.Error("Expected server in config")
	}

	if config.Port == 0 {
		t.Error("Expected port in config")
	}

	if config.Token == "" {
		t.Error("Expected token in config")
	}
}

// Test CORS headers
func TestCORSHeaders(t *testing.T) {
	req := httptest.NewRequest("OPTIONS", "/register", nil)
	w := httptest.NewRecorder()

	handleRequest(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}

	headers := w.Header()
	if headers.Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS headers")
	}
}

// Test error handling
func TestErrorHandling(t *testing.T) {
	storage = NewMockStorage()

	// Test invalid JSON
	req := httptest.NewRequest("POST", "/register", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handleRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
	}

	// Test missing email
	req = createTestRequest("POST", "/register", map[string]string{})

	w = httptest.NewRecorder()
	handleRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing email, got %d", w.Code)
	}
}

// Test concurrent access
func TestConcurrentAccess(t *testing.T) {
	storage = NewMockStorage()

	// Test concurrent client creation
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			email := fmt.Sprintf("test%d@example.com", id)
			clientData := getOrCreateClientWithInfrastructure(email)
			if clientData == nil || clientData.ID == "" {
				t.Errorf("Expected client ID for concurrent access")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Benchmark tests
func BenchmarkGenerateOTP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateOTP()
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	clientID := uuid.New().String()
	for i := 0; i < b.N; i++ {
		generateToken(clientID)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	clientID := uuid.New().String()
	token := generateToken(clientID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateToken(token)
	}
}
