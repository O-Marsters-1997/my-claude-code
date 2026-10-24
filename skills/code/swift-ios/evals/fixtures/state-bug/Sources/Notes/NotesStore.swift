import Foundation
import Observation

struct Note: Identifiable, Hashable {
    var id = UUID()
    var title: String
    var pinned = false
}

@MainActor
@Observable
final class NotesStore {
    var notes: [Note] = []

    func add(_ title: String) {
        notes.insert(Note(title: title), at: 0)
    }

    func delete(at offsets: IndexSet) {
        notes.remove(atOffsets: offsets)
    }
}
