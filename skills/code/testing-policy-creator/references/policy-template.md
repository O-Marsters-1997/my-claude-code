# Policy template

The generated `.claude/skills/testing-policy/SKILL.md` follows this shape. Replace every `<…>`.
Drop a per-layer section when the map row says all there is to say.

````markdown
---
name: testing-policy
description: >
  This repo's testing policy: which code gets which test, which doubles it uses, and what it
  asserts. TRIGGER — load BEFORE writing, editing or reviewing any <test file pattern>, fixture,
  snapshot, fake or fuzz target; or when the prompt says "write a test", "add tests", "cover
  this", "fix this test", or implements a feature with TDD. Load it alongside <testing skill>,
  which owns how a test is written. SKIP for <out-of-scope code>.
---

# Testing policy

How a test is written is in `<testing skill>`; it isn't repeated here.

## Core rule

- Only runtime code is tested: code that ships and runs. No tests of fixtures, config, docs,
  generated code or other tests.
- A test sits beside the code it tests and its name matches: `<foo.ext>` → `<foo test name>`.
  A test file with no matching source file breaks the policy.
- Fixtures, snapshots and golden files exist only as inputs to such a test. One that no test
  reads is deleted.
- Fuzz targets are allowed when they target a runtime function.

<If the grilling overrode any of these, state the replacement rule here instead.>

## Before writing a test

1. Name the row of the map below that the test fills. If no row fits, stop and ask.
2. Check the never-tested list. If the code is on it, test it through its caller.
3. Test through an exported surface. Needing internals is a design smell: test through the
   caller, or move the logic behind a surface that deserves a test.

A **unit** is the behaviour a <module/package>'s exported surface offers, not a function or a file.

## The map

| Code | File | Size | Doubles | Asserts | Guards |
|---|---|---|---|---|---|
| <layer> | <test file> | <small/medium/large> | <doubles> | <observable outcome> | <regression or promise it protects> |

**Never tested directly:** <unexported helpers, generated code, wiring, entrypoints, plain data
types, anything whose only behaviour is calling something else>.

## Doubles

Fidelity order: real > fake > stub. <Mock library policy.>

- **<Managed dependency> → <real thing or stateful fake + contract suite>.**
- **<Another module's surface> → behaviour stub named by what it does.**
- **<Unmanaged edge> → <recorded server / recording stub>.**

**Interaction assertions** only for <commands across an unmanaged boundary>. Asserting that one
of this repo's own modules was called is a change-detector: rewrite it at the output or delete it.

## <Layer> tests

<Rules specific to this layer, with one trimmed example lifted from a good existing test.>

## Boundary gaps

Code that has no testable surface yet. Don't write a test that reaches around the gap; add the
seam first.

- **<area>:** <why it can't be tested through a surface> → <the seam it needs>.

## Naming

- <Test function naming, subtest naming by behaviour in plain words.>
- The file name says the row, so a reader knows the size and doubles from it alone.
- When a large test catches a bug, add the smaller test that should have caught it.
````
