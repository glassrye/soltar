package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	ID       string    `json:"id"`
	Email    string    `json:"email"`
	Created  time.Time `json:"created"`
	LastSeen time.Time `json:"last_seen"`
}

type Environment struct {
	ID        string    `json:"id"`
	ClientID  string    `json:"client_id"`
	VPNServer string    `json:"vpn_server"`
	VPNPort   int       `json:"vpn_port"`
	Created   time.Time `json:"created"`
	Status    string    `json:"status"`
	Region    string    `json:"region"`
	Instances []string  `json:"instances"`
	Databases []string  `json:"databases"`
	Storage   []string  `json:"storage"`
}

type Infrastructure struct {
	VPNInstances  []string  `json:"vpn_instances"`
	LoadBalancers []string  `json:"load_balancers"`
	Databases     []string  `json:"databases"`
	Storage       []string  `json:"storage"`
	Created       time.Time `json:"created"`
	LastUpdated   time.Time `json:"last_updated"`
}

type OTPRequest struct {
	Email string `json:"email"`
}

type OTPVerify struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type AuthResponse struct {
	ClientID    string      `json:"client_id"`
	Token       string      `json:"token"`
	Environment Environment `json:"environment"`
}

type VPNConfig struct {
	Server        string `json:"server"`
	Port          int    `json:"port"`
	Token         string `json:"token"`
	EnvironmentID string `json:"environment_id"`
}

type InfrastructureUpdate struct {
	Infrastructure Infrastructure `json:"infrastructure"`
}

// Storage interface
type Storage interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error
}

// Redis Storage implementation
type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(redisURL string) (*RedisStorage, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisStorage{client: client}, nil
}

func (rs *RedisStorage) Get(key string) ([]byte, error) {
	log.Printf("Redis Get: key='%s'", key)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			log.Printf("Redis Get error for key '%s': key not found", key)
			return nil, fmt.Errorf("key not found: %s", key)
		}
		log.Printf("Redis Get error for key '%s': %v", key, err)
		return nil, err
	}

	log.Printf("Redis Get success for key '%s': %d bytes", key, len(data))
	return data, nil
}

func (rs *RedisStorage) Put(key string, value []byte) error {
	log.Printf("Redis Put: key='%s', value=%d bytes", key, len(value))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := rs.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		log.Printf("Redis Put error for key '%s': %v", key, err)
	}
	return err
}

func (rs *RedisStorage) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return rs.client.Del(ctx, key).Err()
}

// InMemoryStorage implementation for fallback
type InMemoryStorage struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewInMemoryStorage() Storage {
	return &InMemoryStorage{
		data: make(map[string][]byte),
	}
}

func (m *InMemoryStorage) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if data, exists := m.data[key]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

func (m *InMemoryStorage) Put(key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *InMemoryStorage) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

var (
	storage Storage
	secret  = []byte(getEnv("JWT_SECRET", "your-secret-key-change-in-production"))
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Initialize Redis storage with retry
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	var err error

	// Retry Redis connection
	for i := 0; i < 20; i++ {
		storage, err = NewRedisStorage(redisURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to Redis (attempt %d/20): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Printf("Warning: Failed to initialize Redis storage after 20 attempts: %v", err)
		log.Printf("Starting with in-memory storage fallback")
		// Use in-memory storage as fallback
		storage = NewInMemoryStorage()
	} else {
		log.Printf("Connected to Redis at %s", redisURL)
	}

	// Start HTTP server
	port := getEnv("PORT", "8080")
	log.Printf("Starting Soltar VPN server on port %s", port)

	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Handle API requests
	if strings.HasPrefix(r.URL.Path, "/register") ||
		strings.HasPrefix(r.URL.Path, "/verify") ||
		strings.HasPrefix(r.URL.Path, "/connect") ||
		strings.HasPrefix(r.URL.Path, "/config") ||
		strings.HasPrefix(r.URL.Path, "/infrastructure") ||
		strings.HasPrefix(r.URL.Path, "/health") ||
		strings.HasPrefix(r.URL.Path, "/debug") {

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		parts := strings.Split(path, "/")

		switch {
		case r.Method == "GET" && parts[0] == "health":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "healthy",
				"service": "soltar-vpn",
			})
		case r.Method == "GET" && parts[0] == "debug" && len(parts) > 1:
			handleDebug(w, r, parts[1])
		case r.Method == "GET" && parts[0] == "debug" && len(parts) == 1:
			handleDebugList(w, r)
		case r.Method == "POST" && parts[0] == "register":
			handleRegister(w, r)
		case r.Method == "POST" && parts[0] == "verify":
			handleVerify(w, r)
		case r.Method == "POST" && parts[0] == "connect":
			handleConnect(w, r)
		case r.Method == "GET" && parts[0] == "config":
			handleConfig(w, r)
		case r.Method == "POST" && parts[0] == "infrastructure":
			handleInfrastructure(w, r)
		case r.Method == "GET" && parts[0] == "infrastructure":
			handleGetInfrastructure(w, r)
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
		return
	}

	// Serve static files from webapp directory
	fs := http.FileServer(http.Dir("webapp"))
	fs.ServeHTTP(w, r)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received registration request from %s", r.RemoteAddr)

	var req OTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Registration request for email: %s", req.Email)

	if req.Email == "" {
		log.Printf("Missing email in request")
		http.Error(w, "Missing email", http.StatusBadRequest)
		return
	}

	// Generate OTP
	otp := generateOTP()
	log.Printf("Generated OTP for %s: %s", req.Email, otp)

	// Store OTP temporarily (5 minutes expiry)
	// Use a safe key format for Redis
	otpKey := fmt.Sprintf("otp:%s", req.Email)
	otpData := map[string]interface{}{
		"otp":      otp,
		"expires":  time.Now().Add(5 * time.Minute).Unix(),
		"attempts": 0,
	}

	otpBytes, _ := json.Marshal(otpData)
	storage.Put(otpKey, otpBytes)

	// Send OTP via email (implement your email service)
	sendOTPEmail(req.Email, otp)

	log.Printf("Registration successful for %s", req.Email)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "OTP sent to email",
	})
}

func handleVerify(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received verification request from %s", r.RemoteAddr)

	var req OTPVerify
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode verification request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Verification request for email: %s, OTP: %s", req.Email, req.OTP)

	// Verify OTP
	// Use the same safe key format for Redis
	otpKey := fmt.Sprintf("otp:%s", req.Email)
	otpBytes, err := storage.Get(otpKey)
	if err != nil {
		log.Printf("Failed to get OTP for %s: %v", req.Email, err)
		http.Error(w, "Invalid OTP", http.StatusBadRequest)
		return
	}

	var otpData map[string]interface{}
	if err := json.Unmarshal(otpBytes, &otpData); err != nil {
		log.Printf("Failed to unmarshal OTP data: %v", err)
		http.Error(w, "Invalid OTP", http.StatusBadRequest)
		return
	}

	log.Printf("Stored OTP data: %+v", otpData)
	log.Printf("Comparing stored OTP '%s' with provided OTP '%s'", otpData["otp"], req.OTP)

	if otpData["otp"] != req.OTP {
		log.Printf("OTP mismatch for %s", req.Email)
		http.Error(w, "Invalid OTP", http.StatusBadRequest)
		return
	}

	// Check expiry
	if time.Now().Unix() > int64(otpData["expires"].(float64)) {
		log.Printf("OTP expired for %s", req.Email)
		storage.Delete(otpKey)
		http.Error(w, "OTP expired", http.StatusBadRequest)
		return
	}

	log.Printf("OTP verification successful for %s", req.Email)

	// Create or get client with infrastructure
	clientData := getOrCreateClientWithInfrastructure(req.Email)

	// Generate JWT token
	token := generateToken(clientData.ID)

	// Clean up OTP
	storage.Delete(otpKey)

	log.Printf("Verification completed successfully for %s", req.Email)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		ClientID:    clientData.ID,
		Token:       token,
		Environment: clientData.Environment,
	})
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	clientID, err := validateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get client infrastructure
	clientData := getClientInfrastructure(clientID)
	if clientData == nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	// Update last seen
	updateClientLastSeen(clientID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "connected",
		"client_id":   clientID,
		"environment": clientData.Environment,
	})
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	clientID, err := validateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get client infrastructure
	clientData := getClientInfrastructure(clientID)
	if clientData == nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	config := VPNConfig{
		Server:        clientData.Environment.VPNServer,
		Port:          clientData.Environment.VPNPort,
		Token:         token,
		EnvironmentID: clientData.Environment.ID,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(config)
}

func handleInfrastructure(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	clientID, err := validateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var req InfrastructureUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update client infrastructure
	updateClientInfrastructure(clientID, req.Infrastructure)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Infrastructure updated",
		"client_id": clientID,
	})
}

func handleGetInfrastructure(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	clientID, err := validateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get client infrastructure
	clientData := getClientInfrastructure(clientID)
	if clientData == nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"client_id":      clientID,
		"infrastructure": clientData.Infrastructure,
		"environment":    clientData.Environment,
	})
}

func generateOTP() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	return fmt.Sprintf("%06d", int(bytes[0])<<16|int(bytes[1])<<8|int(bytes[2]))[:6]
}

type ClientData struct {
	ID             string         `json:"id"`
	Email          string         `json:"email"`
	Created        time.Time      `json:"created"`
	LastSeen       time.Time      `json:"last_seen"`
	Environment    Environment    `json:"environment"`
	Infrastructure Infrastructure `json:"infrastructure"`
}

func getOrCreateClientWithInfrastructure(email string) *ClientData {
	// Check if client exists
	clientKey := fmt.Sprintf("client:%s", email)
	clientBytes, err := storage.Get(clientKey)

	if err == nil {
		// Client exists, return existing data
		var client ClientData
		json.Unmarshal(clientBytes, &client)
		return &client
	}

	// Create new client with infrastructure
	clientID := uuid.New().String()
	environmentID := uuid.New().String()

	// Generate unique environment for this client
	environment := Environment{
		ID:        environmentID,
		ClientID:  clientID,
		VPNServer: fmt.Sprintf("vpn-%s.soltar.com", clientID[:8]),
		VPNPort:   443,
		Created:   time.Now(),
		Status:    "active",
		Region:    getEnv("REGION", "us-east-1"),
		Instances: []string{},
		Databases: []string{},
		Storage:   []string{},
	}

	infrastructure := Infrastructure{
		VPNInstances:  []string{},
		LoadBalancers: []string{},
		Databases:     []string{},
		Storage:       []string{},
		Created:       time.Now(),
		LastUpdated:   time.Now(),
	}

	client := ClientData{
		ID:             clientID,
		Email:          email,
		Created:        time.Now(),
		LastSeen:       time.Now(),
		Environment:    environment,
		Infrastructure: infrastructure,
	}

	newClientBytes, _ := json.Marshal(client)
	storage.Put(clientKey, newClientBytes)

	// Also store by ID for reverse lookup
	idKey := fmt.Sprintf("client_id:%s", clientID)
	storage.Put(idKey, newClientBytes)

	// Store environment separately
	envKey := fmt.Sprintf("environment:%s", environmentID)
	envBytes, _ := json.Marshal(environment)
	storage.Put(envKey, envBytes)

	return &client
}

func getClientInfrastructure(clientID string) *ClientData {
	clientBytes, err := storage.Get(fmt.Sprintf("client_id:%s", clientID))
	if err != nil {
		return nil
	}

	var client ClientData
	json.Unmarshal(clientBytes, &client)
	return &client
}

func updateClientInfrastructure(clientID string, infrastructure Infrastructure) {
	clientData := getClientInfrastructure(clientID)
	if clientData == nil {
		return
	}

	clientData.Infrastructure = infrastructure
	clientData.Infrastructure.LastUpdated = time.Now()

	updatedBytes, _ := json.Marshal(clientData)
	storage.Put(fmt.Sprintf("client_id:%s", clientID), updatedBytes)
	storage.Put(fmt.Sprintf("client:%s", clientData.Email), updatedBytes)
}

func updateClientLastSeen(clientID string) {
	clientData := getClientInfrastructure(clientID)
	if clientData == nil {
		return
	}

	clientData.LastSeen = time.Now()

	updatedBytes, _ := json.Marshal(clientData)
	storage.Put(fmt.Sprintf("client_id:%s", clientID), updatedBytes)
	storage.Put(fmt.Sprintf("client:%s", clientData.Email), updatedBytes)
}

func generateToken(clientID string) string {
	claims := jwt.RegisteredClaims{
		Subject:   clientID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(secret)
	return tokenString
}

func validateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	return claims.Subject, nil
}

func sendOTPEmail(email, otp string) {
	// For development, just print the OTP to console
	// In production, implement actual email sending
	log.Printf("OTP for %s: %s", email, otp)

	// TODO: Implement actual email sending with your preferred service:
	// - SendGrid: https://sendgrid.com/
	// - AWS SES: https://aws.amazon.com/ses/
	// - SMTP with your own server
	// - Resend: https://resend.com/

	// Example with SMTP:
	/*
		from := getEnv("SMTP_FROM", "noreply@soltar.com")
		smtpHost := getEnv("SMTP_HOST", "smtp.gmail.com")
		smtpPort := getEnv("SMTP_PORT", "587")
		smtpUser := getEnv("SMTP_USER", "")
		smtpPass := getEnv("SMTP_PASS", "")

		msg := fmt.Sprintf("From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: Soltar VPN OTP\r\n\r\n"+
			"Your one-time password is: %s\r\n"+
			"This code will expire in 10 minutes.\r\n", from, email, otp)

		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{email}, []byte(msg))
		if err != nil {
			log.Printf("Failed to send email: %v", err)
		}
	*/
}

func handleDebug(w http.ResponseWriter, r *http.Request, key string) {
	log.Printf("Debug request for key: %s", key)

	data, err := storage.Get(key)
	if err != nil {
		log.Printf("Debug: failed to get key '%s': %v", key, err)
		http.Error(w, fmt.Sprintf("Key not found: %v", err), http.StatusNotFound)
		return
	}

	log.Printf("Debug: found data for key '%s': %d bytes", key, len(data))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":  key,
		"data": string(data),
		"size": len(data),
	})
}

func handleDebugList(w http.ResponseWriter, r *http.Request) {
	log.Printf("Debug list request")

	// Try to get all keys from Redis
	// Note: Redis doesn't have a direct "list all keys" method in this implementation
	// We'll try some common key patterns

	keys := []string{}

	// Try to get the client data
	_, err := storage.Get("client_ac5f3df0-4f70-4cb2-846d-5cc0e4f2e2c9")
	if err == nil {
		keys = append(keys, "client_ac5f3df0-4f70-4cb2-846d-5cc0e4f2e2c9")
	}

	// Try to get OTP data
	_, err = storage.Get("otp:glassrye@gmail.com")
	if err == nil {
		keys = append(keys, "otp:glassrye@gmail.com")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available_keys": keys,
		"total_keys":     len(keys),
	})
}
