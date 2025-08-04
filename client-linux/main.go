package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	API_BASE = "http://localhost:8080"
)

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

type Environment struct {
	ID        string `json:"id"`
	ClientID  string `json:"client_id"`
	VPNServer string `json:"vpn_server"`
	VPNPort   int    `json:"vpn_port"`
	Status    string `json:"status"`
	Region    string `json:"region"`
}

func main() {
	fmt.Println("🔒 Soltar VPN Client (Linux)")
	fmt.Println("=============================")

	// Check if we have stored credentials
	clientID := os.Getenv("SOLTAR_CLIENT_ID")
	token := os.Getenv("SOLTAR_TOKEN")

	if clientID == "" || token == "" {
		fmt.Println("No stored credentials found. Starting registration process...")
		clientID, token = registerAndVerify()
	}

	if clientID == "" || token == "" {
		fmt.Println("❌ Failed to get credentials")
		return
	}

	fmt.Printf("✅ Authenticated as client: %s\n", clientID)
	fmt.Printf("🔑 Token: %s...\n", token[:20])

	// Test connection
	testConnection(clientID, token)
}

func registerAndVerify() (string, string) {
	var email string
	fmt.Print("Enter your email: ")
	fmt.Scanln(&email)

	// Step 1: Register
	fmt.Println("\n📧 Sending registration request...")
	resp, err := http.Post(API_BASE+"/register", "application/json",
		bytes.NewBufferString(fmt.Sprintf(`{"email":"%s"}`, email)))
	if err != nil {
		fmt.Printf("❌ Registration failed: %v\n", err)
		return "", ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Registration failed: %s\n", string(body))
		return "", ""
	}

	fmt.Println("✅ Registration successful! Check server logs for OTP.")

	// Step 2: Get OTP from user
	var otp string
	fmt.Print("Enter the OTP from server logs: ")
	fmt.Scanln(&otp)

	// Step 3: Verify OTP
	fmt.Println("\n🔐 Verifying OTP...")
	verifyData := OTPVerify{
		Email: email,
		OTP:   otp,
	}
	verifyJSON, _ := json.Marshal(verifyData)

	resp, err = http.Post(API_BASE+"/verify", "application/json", bytes.NewBuffer(verifyJSON))
	if err != nil {
		fmt.Printf("❌ Verification failed: %v\n", err)
		return "", ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Verification failed: %s\n", string(body))
		return "", ""
	}

	var authResp AuthResponse
	json.NewDecoder(resp.Body).Decode(&authResp)

	fmt.Printf("✅ Verification successful!\n")
	fmt.Printf("🆔 Client ID: %s\n", authResp.ClientID)
	fmt.Printf("🌐 VPN Server: %s\n", authResp.Environment.VPNServer)

	return authResp.ClientID, authResp.Token
}

func testConnection(clientID, token string) {
	fmt.Println("\n🔗 Testing connection...")

	req, _ := http.NewRequest("POST", API_BASE+"/connect", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Connection failed: %s\n", string(body))
		return
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println("✅ Connection successful!")
	fmt.Printf("📊 Status: %v\n", result["status"])
	fmt.Printf("🆔 Client ID: %v\n", result["client_id"])

	// Test config endpoint
	testConfig(token)
}

func testConfig(token string) {
	fmt.Println("\n⚙️  Testing config endpoint...")

	req, _ := http.NewRequest("GET", API_BASE+"/config", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Config failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Config failed: %s\n", string(body))
		return
	}

	var config map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&config)

	fmt.Println("✅ Config retrieved successfully!")
	fmt.Printf("🌐 Server: %v\n", config["server"])
	fmt.Printf("🔌 Port: %v\n", config["port"])
	fmt.Printf("🆔 Environment ID: %v\n", config["environment_id"])
}
