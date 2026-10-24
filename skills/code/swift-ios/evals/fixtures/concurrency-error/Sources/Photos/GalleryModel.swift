import SwiftUI
import Observation

@MainActor
@Observable
final class GalleryModel {
    var thumbnails: [URL: UIImage] = [:]
    let cache = ThumbnailCache()

    func load(_ urls: [URL]) {
        for url in urls {
            DispatchQueue.global().async {
                Task {
                    let image = try await self.cache.image(for: url)
                    self.thumbnails[url] = image
                }
            }
        }
    }
}
