import Foundation
import Network
import Security

class SoltarVPNClient {
    private let baseURL = "https://your-worker.your-subdomain.workers.dev"
    private var authToken: String?
    private var clientID: String?
    
    struct OTPRequest: Codable {
        let email: String
    }
    
    struct OTPVerify: Codable {
        let email: String
        let otp: String
    }
    
    struct AuthResponse: Codable {
        let client_id: String
        let token: String
        let environment: Environment
    }
    
    struct Environment: Codable {
        let id: String
        let vpn_server: String
        let vpn_port: Int
        let created: String
        let status: String
    }
    
    struct VPNConfig: Codable {
        let server: String
        let port: Int
        let token: String
        let environment_id: String
    }
    
    struct Infrastructure: Codable {
        let vpn_instances: [String]
        let load_balancers: [String]
        let databases: [String]
        let storage: [String]
        let created: String
        let last_updated: String
    }
    
    func register(email: String) async throws -> Bool {
        let url = URL(string: "\(baseURL)/register")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        let body = OTPRequest(email: email)
        request.httpBody = try JSONEncoder().encode(body)
        
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw VPNError.registrationFailed
        }
        
        return true
    }
    
    func verifyOTP(email: String, otp: String) async throws -> (String, String, Environment) {
        let url = URL(string: "\(baseURL)/verify")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        let body = OTPVerify(email: email, otp: otp)
        request.httpBody = try JSONEncoder().encode(body)
        
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw VPNError.verificationFailed
        }
        
        let authResponse = try JSONDecoder().decode(AuthResponse.self, from: data)
        self.clientID = authResponse.client_id
        self.authToken = authResponse.token
        
        return (authResponse.client_id, authResponse.token, authResponse.environment)
    }
    
    func getVPNConfig() async throws -> VPNConfig {
        guard let token = authToken else {
            throw VPNError.notAuthenticated
        }
        
        let url = URL(string: "\(baseURL)/config")!
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw VPNError.configFailed
        }
        
        return try JSONDecoder().decode(VPNConfig.self, from: data)
    }
    
    func connect() async throws -> Bool {
        guard let token = authToken else {
            throw VPNError.notAuthenticated
        }
        
        let url = URL(string: "\(baseURL)/connect")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw VPNError.connectionFailed
        }
        
        return true
    }
}

enum VPNError: Error {
    case registrationFailed
    case verificationFailed
    case notAuthenticated
    case configFailed
    case connectionFailed
}

// VPN Manager for macOS
class VPNManager {
    private let client = SoltarVPNClient()
    private var vpnConnection: NWConnection?
    
    func setupVPN(email: String, otp: String) async throws {
        // Register and verify
        _ = try await client.register(email: email)
        let (clientID, token, environment) = try await client.verifyOTP(email: email, otp: otp)
        
        print("✅ Authenticated with environment: \(environment.vpn_server)")
        
        // Get VPN configuration
        let config = try await client.getVPNConfig()
        
        // Connect to VPN
        _ = try await client.connect()
        
        // Setup local VPN connection
        try await setupLocalVPNConnection(config: config)
    }
    
    private func setupLocalVPNConnection(config: SoltarVPNClient.VPNConfig) async throws {
        let endpoint = NWEndpoint.hostPort(
            host: NWEndpoint.Host(config.server),
            port: NWEndpoint.Port(integerLiteral: UInt16(config.port))
        )
        
        let parameters = NWParameters.tcp
        parameters.defaultProtocolStack.applicationProtocols.insert(
            NWProtocolFramer.Options(), at: 0
        )
        
        vpnConnection = NWConnection(to: endpoint, using: parameters)
        
        vpnConnection?.stateUpdateHandler = { state in
            switch state {
            case .ready:
                print("VPN connected")
            case .failed(let error):
                print("VPN failed: \(error)")
            case .cancelled:
                print("VPN cancelled")
            default:
                break
            }
        }
        
        vpnConnection?.start(queue: .main)
    }
    
    func disconnect() {
        vpnConnection?.cancel()
        vpnConnection = nil
    }
}

// CLI Interface
@main
struct SoltarVPNCLI {
    static func main() async {
        print("Soltar VPN Client")
        print("=================")
        
        let vpnManager = VPNManager()
        
        print("Enter your email: ", terminator: "")
        guard let email = readLine() else {
            print("Invalid email")
            return
        }
        
        do {
            print("Registering...")
            _ = try await vpnManager.client.register(email: email)
            print("OTP sent to your email")
            
            print("Enter OTP: ", terminator: "")
            guard let otp = readLine() else {
                print("Invalid OTP")
                return
            }
            
            print("Verifying OTP...")
            try await vpnManager.setupVPN(email: email, otp: otp)
            print("VPN connected successfully!")
            
            // Keep running
            RunLoop.main.run()
            
        } catch {
            print("Error: \(error)")
        }
    }
} 