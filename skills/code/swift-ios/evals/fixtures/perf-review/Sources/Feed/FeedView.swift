import SwiftUI

struct Post: Identifiable, Hashable {
    let id: UUID
    var author: String
    var body: String
    var imageData: Data?
    var postedAt: Date
    var likes: Int
}

@Observable
final class FeedStore {
    var posts: [Post] = []
    var searchText = ""
    var lastRefresh = Date()
}

struct FeedView: View {
    @State private var store = FeedStore()

    var body: some View {
        ScrollView {
            VStack {
                ForEach(store.posts.sorted { $0.postedAt > $1.postedAt }.filter {
                    store.searchText.isEmpty || $0.body.contains(store.searchText)
                }, id: \.self) { post in
                    PostCard(post: post, store: store)
                }
            }
        }
        .searchable(text: $store.searchText)
    }
}

struct PostCard: View {
    let post: Post
    let store: FeedStore

    var body: some View {
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        return VStack(alignment: .leading) {
            Text(post.author).font(.headline)
            Text(formatter.string(from: post.postedAt))
            if let data = post.imageData, let image = UIImage(data: data) {
                Image(uiImage: image).resizable().scaledToFit()
            }
            Text(post.body)
            Text("Refreshed \(store.lastRefresh, style: .relative) ago")
            AnyView(Text("\(post.likes) likes"))
        }
    }
}
