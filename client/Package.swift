// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "SoltarVPNClient",
    platforms: [
        .macOS(.v13)
    ],
    products: [
        .executable(
            name: "soltar-vpn",
            targets: ["SoltarVPNClient"]
        )
    ],
    targets: [
        .executableTarget(
            name: "SoltarVPNClient",
            path: ".",
            sources: ["main.swift"]
        )
    ]
) 