# New features

These APIs are newer than most training data. Follow the examples as written rather than
guessing from older APIs. Check the toolchain first (`swift --version`); each section names the
minimum Swift version. If the toolchain is older, don't use the feature.

## Contents

- Raw identifiers (6.2)
- Range-based confirmations (6.1)
- Test scoping traits (6.1)
- Exit tests (6.2)
- Attachments (6.2)
- Evaluating a `ConditionTrait` (6.2)
- `#expect(throws:)` returns the error (6.1)

## Raw identifiers

**Swift 6.2+.** A backticked function name can be a sentence, which replaces the display-name
string:

```swift
@Test func `Strip HTML tags from string`() { … }

@Test(arguments: [(32, 0), (212, 100), (-40, -40)])
func `Fahrenheit converts to Celsius`(values: (input: Double, output: Double)) { … }
```

Operators like `+` are allowed as long as they aren't the whole name. Many users find this style
surprising. Use it only if the project already does; otherwise suggest it.

## Range-based confirmations

**Swift 6.1+.**

```swift
await confirmation(expectedCount: 5...10) { confirm in
    for await _ in loader { confirm() }
}
```

`5...` means at least five. `...10` is rejected as ambiguous.

## Test scoping traits

**Swift 6.1+.** A trait that wraps each test it's applied to with setup and teardown, including
async teardown, which `deinit` can't do. Pair it with `@TaskLocal` to give each test its own
configuration without shared mutable state.

```swift
struct Player {
    var name: String
    @TaskLocal static var current = Player(name: "Anonymous")
}

struct DefaultPlayerTrait: TestTrait, SuiteTrait, TestScoping {
    func provideScope(
        for test: Test,
        testCase: Test.Case?,
        performing function: () async throws -> Void
    ) async throws {
        try await Player.$current.withValue(Player(name: "Natsuki Subaru")) {
            try await function()
        }
    }
}

extension Trait where Self == DefaultPlayerTrait {
    static var defaultPlayer: Self { Self() }
}

@Test(.defaultPlayer) func welcomeShowsName() {
    #expect(makeWelcome().contains("Natsuki Subaru"))
}
```

- Conform to `SuiteTrait` as well to allow `@Suite(.defaultPlayer)`.
- Several scopes combine (`@Test(.first, .second)`) and apply in the order listed.
- Async teardown goes after `try await function()`. Wrap the call in `do`/`catch` and rethrow if
  teardown must run when the test throws.

## Exit tests

**Swift 6.2+.** Tests code that calls `precondition`, `fatalError` or `exit`. The body runs in a
child process. Not available on iOS devices or simulators: macOS, Linux and Windows only.

```swift
@Test func zeroSidedDiceTraps() async {
    await #expect(processExitsWith: .failure) {
        _ = Dice().roll(sides: 0)
    }
}
```

The `await` is required. Before Swift 6.3 the closure can't capture state from the test, so
build what it needs inside it.

## Attachments

**Swift 6.2+.** Attach data to a test so it appears in the results when the test fails.

```swift
@Test func defaultCharacterName() {
    let character = makeCharacter()
    Attachment.record(character, named: "Character")
    #expect(character.name == "Rem")
}
```

- `String`, `Data` and any `Encodable` type work out of the box (Encodable needs `import
  Foundation`). The type must conform to `Attachable`; add the conformance in the test target.
- Images need Swift 6.3.
- No lifetime controls, unlike `XCTAttachment`.

## Evaluating a `ConditionTrait`

**Swift 6.2+.** Reuse a test condition outside a test:

```swift
let trait = ConditionTrait.disabled(if: TestEnvironment.isSmokeRun)
if try await trait.evaluate() { … }
```

## `#expect(throws:)` returns the error

**Swift 6.1+.** See `expectations.md` § Testing throws. Replace the deprecated
`#expect { … } throws: { … }` form.
