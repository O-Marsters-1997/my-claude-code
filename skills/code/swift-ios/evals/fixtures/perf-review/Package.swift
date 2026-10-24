// swift-tools-version: 6.2
import PackageDescription

let package = Package(
    name: "Feed",
    platforms: [.iOS(.v18)],
    targets: [
        .target(
            name: "Feed",
            swiftSettings: [.defaultIsolation(MainActor.self)]
        )
    ]
)
