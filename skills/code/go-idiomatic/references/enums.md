# Enums — Closed Value Sets in Go

## Contents

- [Rules](#rules)
- [Worked example](#worked-example)
- [Exhaustive switches](#exhaustive-switches)
- [Stored values](#stored-values)

## Rules

1. **One definition per concept.** Define the type once, in the package that owns the concept, and
   import it. A second copy (a `PRState` in both the GitHub client and the planner, a DTO constant
   that mirrors a generated one) drifts the first time a value is added.
2. **String-backed type, lookup table.** `type State string` with named constants. Behaviour that
   varies by value (labels, ordering, transitions) lives in a map keyed by the constant, not in a
   50-line `switch`. Adding a value is then one constant and one map entry.
3. **Zero value means unknown.** The empty string is never a valid member, so an unset field is
   detectable. Do not give the zero value a meaning such as "pending".
4. **Parse returns an error on an unknown value.** Every boundary that turns outside input into the
   type (JSON, flags, a database scan) goes through one `Parse` that rejects what it does not
   recognise. A bare cast `State(s)` skips validation, and a `default:` that returns a fallback
   hides corrupt data.
5. **No `default` in a switch meant to be exhaustive.** Let the `exhaustive` linter fail the build
   when a value is added.

## Worked example

```go
package order

import "fmt"

type Status string

const (
	StatusPlaced    Status = "placed"
	StatusShipped   Status = "shipped"
	StatusDelivered Status = "delivered"
)

var statusLabels = map[Status]string{
	StatusPlaced:    "Placed",
	StatusShipped:   "On its way",
	StatusDelivered: "Delivered",
}

func ParseStatus(s string) (Status, error) {
	st := Status(s)
	if _, ok := statusLabels[st]; !ok {
		return "", fmt.Errorf("parse status %q: unknown value", s)
	}
	return st, nil
}

func (s Status) Label() string { return statusLabels[s] }

func (s Status) Valid() bool {
	_, ok := statusLabels[s]
	return ok
}
```

`ParseStatus("")` is an error, and `Status("")` stays the unknown zero value. Callers that receive
a `Status` from a trusted source need no check; callers at a boundary call `ParseStatus`.

For JSON, implement `UnmarshalText` by calling `ParseStatus`, so decoding rejects unknown values
without every handler remembering to.

## Exhaustive switches

Branching on the value, as opposed to looking up data about it, uses a `switch` with a case for
every constant and no `default`:

```go
func (s Status) Terminal() bool {
	switch s {
	case StatusDelivered:
		return true
	case StatusPlaced, StatusShipped:
		return false
	}
	return false
}
```

Enable the `exhaustive` linter and set `default-signifies-exhaustive: false`:

```yaml
linters:
  enable: [exhaustive]
  settings:
    exhaustive:
      default-signifies-exhaustive: false
```

With the setting `true`, a `default:` branch counts as covering every value, so the linter stays
silent exactly when a new constant is unhandled. That cancels the check.

## Stored values

When the value is persisted, store it as `text` with a `CHECK` constraint listing the same members,
not a native Postgres enum type. See the `postgres` skill for the SQL side. Read it back through
`ParseStatus`, so a row the constraint missed surfaces as an error instead of a silent cast.
