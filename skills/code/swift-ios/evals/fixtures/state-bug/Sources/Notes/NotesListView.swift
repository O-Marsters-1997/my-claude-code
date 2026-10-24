import SwiftUI

struct NotesListView: View {
    @State var store = NotesStore()
    @State private var draft = ""

    var body: some View {
        NavigationStack {
            List {
                ForEach(store.notes.indices, id: \.self) { index in
                    NoteRow(note: store.notes[index])
                }
                .onDelete { store.delete(at: $0) }
            }
            .toolbar {
                TextField("New note", text: $draft)
                Image(systemName: "plus")
                    .onTapGesture {
                        store.add(draft)
                        draft = ""
                    }
            }
        }
    }
}

struct NoteRow: View {
    let note: Note
    @State var isPinned: Bool

    init(note: Note) {
        self.note = note
        _isPinned = State(initialValue: note.pinned)
    }

    var body: some View {
        HStack {
            Text(note.title)
            Spacer()
            Toggle("Pinned", isOn: $isPinned)
                .labelsHidden()
        }
    }
}
