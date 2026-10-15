---
name: testing-policy-creator
description: >
  Writes, or brings up to date, a repo's own testing-policy skill at
  .claude/skills/testing-policy/SKILL.md: which code gets which kind of test, with which doubles,
  asserting what, and why. Surveys the codebase for its layers and boundaries, proposes a draft
  map, grills the user on it row by row, then lists the tests and boundaries that break the
  policy. Use when the user says "create a testing policy", "write our testing-policy skill",
  "decide what we test in this repo", "which tests should we keep", "our tests are a mess, set
  some rules", "purge tests that don't conform", or asks to update an existing testing policy.
  Not for writing a test (use the language's testing skill and tdd) or reviewing a diff.
---

# Testing policy creator

Produces one file in the target repo: `.claude/skills/testing-policy/SKILL.md`. That skill
decides **what** is tested and why. How a test is written (assertions, tables, helpers) belongs
to the language's testing skill (`go-testing` for Go); the policy names it and never restates it.

A good policy does two jobs at once. It stops regressions, and it makes the codebase legible: a
reader who knows the map knows what each layer is for, because each row says what behaviour the
layer promises and which test proves it. Code that fits no row is a design finding, not a gap in
the map.

Read [references/principles.md](references/principles.md) before drafting. It holds the rules for
choosing what earns a test, picking its size and doubles, and reading untestability as coupling.
[references/policy-template.md](references/policy-template.md) is the output shape, and
[references/example-go-policy.md](references/example-go-policy.md) is a finished policy from a Go
service to calibrate against.

## 1. Check for an existing policy

If `.claude/skills/testing-policy/SKILL.md` exists, this is an update. Treat it as the draft:
explore as below, then grill only on what changed (new layers, rows the code no longer matches,
rules the tests ignore). Never overwrite it wholesale.

## 2. Survey the codebase

Dispatch `Explore` sub-agents in parallel, one per question, and keep only their conclusions:

- **Stack.** Languages, test runner, test file convention (`foo_test.go`, `foo.test.ts`,
  `test_foo.py`), fixture and snapshot locations, and which testing skill applies.
- **Layers.** The architectural layers or modules and the exported surface each offers callers
  (services, stores, handlers, adapters, consumers, UI components). Name each by what it does.
- **Boundaries.** Every out-of-process dependency, split into managed (only this app touches it,
  e.g. its own database) and unmanaged (others observe it: third-party APIs, queues, email,
  LLMs). This decides doubles.
- **Existing tests.** For each test file: the source file it pairs with, the layer it exercises,
  its size, its doubles, and what it asserts. Flag files pairing with no source file, fixtures no
  test reads, tests of unexported code, and assertions on calls between the repo's own packages.
- **Boundary gaps.** Code that cannot be tested through an exported surface without reaching
  into internals, standing up the whole app, or mocking the repo's own packages. Note the seam
  each would need.

## 3. Draft the map

From the survey and the principles, draft one row per layer: code, test file, size, doubles,
what it asserts, and a short "guards" note naming the regression or promise the row protects.
Add the never-tested list and the boundary gaps. Start from the core rule in the template.

Where an existing test already shows a row's shape well, lift a trimmed version of it as that
row's example. Examples illustrate a rule; they are chosen because they are good, not because
they exist.

## 4. Grill

Call the Skill tool with "grilling" and give it this agenda, one branch per item:

1. The core rule, and whether this repo's file convention overrides it.
2. Each map row: does the layer exist as described, is this the right size and doubles, does the
   assertion target behaviour a caller observes.
3. The never-tested list.
4. Doubles: fidelity order, where fakes live, whether a fake needs a contract suite.
5. When interaction assertions are allowed.
6. Each boundary gap: keep it in the policy as a named gap, or drop it.

## 5. Write the policy

Write `.claude/skills/testing-policy/SKILL.md` from the template. It states rules that hold
whenever it is read, so it never describes the repo as it was on the day it was written: no
inventory of current tests, no "X is the pilot", no "bring code onto the policy when touched",
no counts or dates. Its description uses the standards-skill shape (`TRIGGER — load BEFORE
writing, editing or reviewing a test … SKIP …`) so it loads on plain "add a test" requests.

Write the reasoning behind each row (the evidence from the survey, the principle applied, the
grilling answer) to a rationale file outside the repo, in the session scratchpad or the system
temp directory. It is a review aid only; tell the user it is safe to delete once they have read it.

## 6. Report and hand off

1. Add one line to the repo's `AGENTS.md` (create it if missing) pointing at the policy:
   `Before writing or reviewing a test, load the testing-policy skill.`
2. Print the tests that break the policy, grouped by the rule they break, then the boundary gaps
   with the seam each needs. Recommend delete, rewrite or move for each test; change nothing.
3. Ask once whether to file both lists as one purge issue. On yes, call the Skill tool with
   "file-issue".
4. Give the path to the rationale file.
