# Traits and tags

Traits go in `@Test(...)` or `@Suite(...)` after the optional display name. Traits on a suite
apply to everything inside it.

## Conditions

```swift
@Test(.enabled(if: Runtime.isCI))
func ciOnlySmokeTest() { … }

@Test(.disabled("Backend sandbox down"), .bug("https://github.com/acme/app/issues/12"))
func refundFlow() { … }
```

- `.disabled("reason")` instead of commenting a test out. Give a reason and ideally a `.bug`.
- `.enabled(if:)` / `.disabled(if:)` take a runtime condition.
- OS gating uses `@available` on the test function, not a trait (see `fundamentals.md`).
- When the code should keep running while broken, use `withKnownIssue` (see `expectations.md`).

## Bugs

Link the bug a test covers or reproduces. It shows in reports and tells the next reader why the
test exists.

```swift
@Test("Headings are always italic", .bug(id: 182))
func headingsAreItalic() { … }

@Test(.bug("https://github.com/acme/app/issues/182"))
func headingsAreItalicFromURL() { … }
```

## Time limits

```swift
@Test(.timeLimit(.minutes(1)))
func loadNames() async { … }
```

The only unit is `.minutes`. `.seconds` does not compile, whatever your training data says. On a
suite the limit applies to each test; where test and suite both set one, the shorter wins.

## Tags

Declare tags once in the test target, then apply them to tests or suites.

```swift
extension Tag {
    @Tag static var networking: Self
    @Tag static var slow: Self
}

@Suite(.tags(.networking))
struct APIClientTests { … }

@Test(.tags(.slow))
func importsLargeArchive() async throws { … }
```

- Tag anything network-related `.networking`, even when it's mocked.
- Other useful tags: `.slow`, `.integration`, `.regression`, `.smoke`. Reuse the repo's existing
  tags before inventing new ones.
- Tags are cross-cutting metadata, not a replacement for suite structure.

## Test plans and filtering

- Filter in Xcode test plans and the test navigator by tag, not by name patterns that break on
  rename.
- Typical plans: a fast core plan excluding `.integration`; an integration plan; a release gate.
- From the command line: `swift test --filter`, or `xcodebuild test -testPlan <plan>`.

## Serialization

`.serialized` is a trait too. Read `isolation-and-parallelism.md` before using it.

## Custom traits

Scoping traits that wrap a test in setup and teardown are covered in `new-features.md`.
