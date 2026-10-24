import SwiftUI

struct Message: Identifiable {
    let id: Int
    var sender: String
    var subject: String
    var unread: Bool
}

struct InboxView: View {
    @State private var messages: [Message] = [
        Message(id: 1, sender: "Ana", subject: "Lunch?", unread: true),
        Message(id: 2, sender: "Ben", subject: "Invoice", unread: false),
    ]

    var body: some View {
        NavigationView {
            List {
                ForEach(messages) { message in
                    NavigationLink(destination: MessageDetail(message: message)) {
                        VStack(alignment: .leading) {
                            Text(message.sender)
                                .font(.headline)
                                .foregroundColor(message.unread ? .primary : .secondary)
                            Text(message.subject)
                                .font(.subheadline)
                        }
                    }
                }
            }
            .navigationBarTitle("Inbox")
            .animation(.default)
        }
    }
}
