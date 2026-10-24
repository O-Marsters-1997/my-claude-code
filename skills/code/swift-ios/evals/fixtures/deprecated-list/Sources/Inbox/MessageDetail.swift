import SwiftUI

struct MessageDetail: View {
    let message: Message

    var body: some View {
        Text(message.subject)
            .font(.title)
            .foregroundColor(.blue)
            .cornerRadius(8)
            .navigationBarTitle(message.sender, displayMode: .inline)
    }
}
