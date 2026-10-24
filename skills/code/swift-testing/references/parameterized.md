# Parameterized tests

## When

Use one parameterized test when several tests share logic and differ only in data. Each argument
becomes its own test case with its own failure, and can be rerun on its own in Xcode. A `for`
loop inside a test gives none of that.

```swift
@Test("Free features are not premium", arguments: [Feature.recording, .darkMode])
func freeFeatures(_ feature: Feature) {
    #expect(feature.isPremium == false)
}
```

Any `Sendable` collection works: arrays, ranges, dictionaries.

```swift
@Test(arguments: 18...21)
func validAges(_ age: Int) {
    #expect(isValidAge(age))
}
```

## Input with expected output: an array of tuples

Pairs sit side by side and can't drift out of alignment.

```swift
@Test(arguments: [
    (Ingredient.rice, Dish.onigiri),
    (.potato, .fries),
    (.egg, .omelette),
])
func cook(_ ingredient: Ingredient, into dish: Dish) {
    #expect(cook(ingredient) == dish)
}
```

A dictionary literal works too when keys are `Hashable`.

## Two collections are a Cartesian product

`@Test(arguments: a, b)` runs every combination: 4 regions × 5 plans is 20 cases. At most two
collections are allowed. Use it when every combination matters:

```swift
@Test(arguments: [Region.eu, .us], [Plan.free, .pro])
func vatInvoiceAccess(region: Region, plan: Plan) { … }
```

## `zip`, and why to avoid it

`zip(a, b)` pairs elements, but it stops at the shorter collection with no error, so a missing
expected value silently drops a case. `zip(A.allCases, B.allCases)` breaks when someone reorders
an enum. Use tuples. Use `zip` only for two explicit, equal-length arrays that must stay separate.

## Expected values are literals

If the expected value comes from the same input, the test checks itself:

```swift
// Passes even if format(day) has the same casing bug as rawValue
@Test(arguments: Day.allCases)
func dayLabel(day: Day) {
    #expect(format(day) == day.rawValue)
}

// Each case is an independent data point
@Test(arguments: [(Day.monday, "Monday"), (.friday, "Friday")])
func dayLabel(day: Day, expected: String) {
    #expect(format(day) == expected)
}
```

`allCases` is right for properties true of every case ("rotating four times returns the
original"), not for case-by-case mappings.

## No branching in the body

An `if`/`switch` on the argument mirrors the implementation. Split the special case into its own
test.

```swift
@Test func fridayGreeting() {
    #expect(greet(.friday) == "TGIF!")
}

@Test(arguments: [Day.monday, .tuesday, .wednesday, .thursday, .saturday, .sunday])
func standardGreeting(day: Day) {
    #expect(greet(day) == "Hello, \(day)!")
}
```

## Keep it readable

- Arguments inline in `@Test`, not in a distant static property, unless several tests reuse them.
- One concern per parameterized test.
- Meaningful labels. If arguments print badly, conform them to `CustomTestStringConvertible`.
- Watch the case count; a large product slows CI for little extra coverage.
