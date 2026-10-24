# Expectations

## `#expect` by default

Write a plain Swift expression. The macro captures every sub-expression, so the failure shows
the values.

```swift
#expect(total == 20)
#expect(total > 0)
#expect([10, 20, 30].contains(total))
```

Never negate with `!`. `#expect(!isLoggedIn)` reports only "false"; `#expect(isLoggedIn ==
false)` shows the value. The same goes for `#require`.

Add a message when it tells the reader something the expression doesn't:
`#expect(names.isEmpty == false, "Loading should populate names")`.

## `#require` for preconditions

`#require` throws on failure and stops the test. Use it when later lines are meaningless if the
check fails, and to unwrap optionals. The test must be `throws`.

```swift
@Test func outstandingTasksStringIsPlural() throws {
    let sut = try makeUser(projects: 3, itemsPerProject: 10)
    try #require(sut.projects.isEmpty == false)
    #expect(sut.outstandingTasksString == "30 items")
}

let url = try #require(URL(string: value), "URL should parse")
```

No `!` force-unwraps in tests: a crash takes down the whole run, not just that test.

## Testing throws

Name the error you expect. `Error.self` passes for any error, including the wrong one.

```swift
#expect(throws: GameError.notInstalled) {
    try game.play()
}
```

`#expect(throws:)` and `#require(throws:)` return the caught error (Swift 6.1+), so you can
inspect it:

```swift
let error = #expect(throws: ValidationError.self) {
    try form.submit()
}
#expect(error?.field == "email")
```

The old two-closure form (`#expect { … } throws: { … }`) is deprecated; rewrite it as above.

Does not throw: `#expect(throws: Never.self) { try game.play() }`.

When you need branching on several error cases, `do`/`catch` with `Issue.record` is fine:

```swift
do {
    try game.play()
    Issue.record("Expected an error")
} catch GameError.notPurchased {
} catch {
    Issue.record("Wrong error: \(error)")
}
```

## `Issue.record`

Records a failure without stopping the test. It replaces `XCTFail`.

## Known issues

`withKnownIssue` wraps code with a known bug. It expects a failure and *fails* if none is
recorded, so you notice when the bug is fixed. Scope it to the failing lines only.

```swift
withKnownIssue("Server rounds VAT up", .bug(id: 412)) {
    #expect(invoice.vat == 4.17)
}
```

`isIntermittent: true` passes when no issue is recorded, for flaky failures you are investigating.
Prefer `withKnownIssue` to `.disabled` when you want the code to keep compiling and running.

## Floating point

Swift Testing has no tolerance comparison. If the project already uses Swift Numerics, use
`#expect(value.isApproximatelyEqual(to: 0, absoluteTolerance: 1e-6))`. Otherwise ask before
adding it, or write `#expect(abs(value - expected) < 1e-6)`.

## Readable failures

Conform a type to `CustomTestStringConvertible` *in the test target* for clearer output:

```swift
extension GameError: @retroactive CustomTestStringConvertible {
    public var testDescription: String {
        switch self {
        case .notPurchased: "Game not purchased"
        case .notInstalled: "Game not installed"
        }
    }
}
```

## Verification helpers

A helper that wraps several expectations takes a `SourceLocation` so failures point at the
calling test, not the helper. The macro still needs its underscore.

```swift
func verifyDivision(
    _ result: (quotient: Int, remainder: Int),
    quotient: Int,
    remainder: Int,
    sourceLocation: SourceLocation = #_sourceLocation
) {
    #expect(result.quotient == quotient, sourceLocation: sourceLocation)
    #expect(result.remainder == remainder, sourceLocation: sourceLocation)
}
```

Pass `sourceLocation:` to every `#expect` and `#require` inside.
