import UIKit

final class ThumbnailCache {
    private var images: [URL: UIImage] = [:]

    func image(for url: URL) async throws -> UIImage {
        if let cached = images[url] { return cached }
        let (data, _) = try await URLSession.shared.data(from: url)
        guard let image = UIImage(data: data)?.preparingThumbnail(of: CGSize(width: 200, height: 200)) else {
            throw URLError(.cannotDecodeContentData)
        }
        images[url] = image
        return image
    }
}
