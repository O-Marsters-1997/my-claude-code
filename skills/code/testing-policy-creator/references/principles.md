# Choosing what to test

Decision rules for mapping a codebase's layers to test types, in the order to apply them.

Sources:
- **Khorikov**: *Unit Testing Principles, Practices, and Patterns* (https://www.manning.com/books/unit-testing)
- **SWE@G**: Winters et al., *Software Engineering at Google*, chapters 11–14 (https://abseil.io/resources/swe-book)
- **GOOS**: Freeman & Pryce, *Growing Object-Oriented Software, Guided by Tests*
- **Feathers**: *Working Effectively with Legacy Code*

## 1. Classify the code

Place each layer on two axes: how complex or domain-significant it is, and how many
collaborators it has (Khorikov ch. 7).

| Quadrant | Example | Test |
|---|---|---|
| Domain logic, algorithms (complex, few collaborators) | pricing rules, parsers, state machines | Small tests, heavily. This is where tests pay most. |
| Controllers, orchestration (simple, many collaborators) | handlers, services that sequence calls | A few integration tests: one happy path per controller plus each edge case small tests can't reach. |
| Trivial (simple, few collaborators) | getters, DTOs, wiring, `main` | Never directly. Covered through callers. |
| Overcomplicated (complex, many collaborators) | a handler that computes and does I/O | Not as it stands. It is a boundary gap: split decisions from I/O (Humble Object) first. |

If a layer won't fit a quadrant, it is mixing decisions with I/O. That is a boundary gap.

## 2. What earns a direct test

- **Test behaviours, not methods.** Write one test per promise the layer makes to its callers
  (SWE@G ch. 12).
- **A unit is a unit of behaviour**, not a class or file. One test may span several types that
  together produce one outcome (Khorikov ch. 2;
  [Fowler, UnitTest](https://martinfowler.com/bliki/UnitTest.html)).
- **Test through the public API real callers use.** A test that reaches into internals means
  the boundary is wrong (SWE@G ch. 12). If a private function is too complex to cover through
  its caller, it is a missing module: extract it (Khorikov ch. 11).
- **Assert observable behaviour only.** That means return values, state clients can see, and
  side effects at the system edge. Everything else is an implementation detail (Khorikov ch. 4–5).
- **Score each candidate test on four pillars:** regression protection, resistance to
  refactoring, fast feedback and maintainability. Resistance to refactoring is not negotiable.
  Trade the others by test size (Khorikov ch. 4).
- **The ideal test changes only when requirements change.** A test that needs editing after a
  pure refactor is mis-targeted (SWE@G ch. 12; [Beck, Test Desiderata](https://testdesiderata.com)).
- **Beyoncé Rule.** Anything that must keep working gets a test, including failure handling,
  security and performance budgets (SWE@G ch. 11).

## 3. Size and scope

- **Size is a resource limit:**
  - Small: one process, no I/O, no sleep.
  - Medium: one machine, localhost only.
  - Large: multiple machines.

  Choose the smallest size that can observe the behaviour (SWE@G ch. 11).
- **Match scope to the layer:**
  - Domain logic: narrow scope, small size.
  - A component together with its own database: medium.
  - Large or end-to-end: only critical user journeys, plus wiring nothing smaller can see
    (SWE@G ch. 11, 14).
- **The portfolio shape follows the architecture:**
  - Pyramid: logic-heavy code ([Vocke](https://martinfowler.com/articles/practical-test-pyramid.html)).
  - Honeycomb: services whose risk is in integration
    ([Spotify](https://engineering.atspotify.com/2018/01/testing-of-microservices)).
  - Trophy: UI apps ([Dodds](https://kentcdodds.com/blog/the-testing-trophy-and-testing-classifications)).

  Most disputes between these are really about what "unit" means. Decide by boundary, not label
  ([Fowler, test shapes](https://martinfowler.com/articles/2021-test-shapes.html)).
- **Coverage is a floor, measured on small tests only, never a target** (SWE@G ch. 11).

## 4. Doubles by boundary

- **Order of preference:** real, then fake, then stub, then mock (SWE@G ch. 13). Use the real
  implementation whenever it is fast and deterministic.
- **Managed dependency** (only this app touches it, e.g. its own database): use the real thing
  in medium tests, never a mock. Its interactions are implementation details (Khorikov ch. 8).
- **Unmanaged dependency** (others observe it: a third-party API, a queue, email, an LLM):
  replace it with a fake or a recording stub. Interactions with it are part of the contract
  (Khorikov ch. 8).
- **Double only types you own.** Wrap third-party clients in an adapter and double the adapter,
  at the last type before the wire (GOOS ch. 8; Khorikov ch. 9).
- **One authoritative fake per dependency,** owned next to the real one, with one contract suite
  run against both (SWE@G ch. 13; [Fowler, ContractTest](https://martinfowler.com/bliki/ContractTest.html)).
- **Inject time, randomness and environment as values,** not ambient globals (Khorikov ch. 8).

## 5. Interaction assertions

- **Prefer state over interaction.** A call happening doesn't prove the result is right
  (SWE@G ch. 12–13).
- **Verify interactions only for state-changing calls on unmanaged dependencies:** a sent email,
  a published message, a billed API call. Never verify reads, and never assert on a stub
  (Khorikov ch. 5).
- **Count or order calls only when the count or order is itself the behaviour** (SWE@G ch. 13).
- **A test that mirrors the code line by line is a change-detector.** It fails on every refactor
  and catches nothing, so delete it
  ([Change-Detector Tests Considered Harmful](https://testing.googleblog.com/2015/01/testing-on-toilet-change-detector-tests.html)).
- **In UI code, query what the user sees, not component internals**
  ([Dodds](https://kentcdodds.com/blog/testing-implementation-details)).

## 6. Untestability is a design finding

When a test is hard to write, the design is what needs to change, not the test machinery
(GOOS ch. 20). Record each case as a boundary gap with the seam it needs:

| Symptom in the test | Missing seam |
|---|---|
| Mocks returning mocks | Decisions tangled with I/O. Split them with Humble Object (Khorikov ch. 7). |
| Needs private access | An abstraction is missing. Extract it (Khorikov ch. 11). |
| Needs a real clock, network or filesystem to stay small | An ambient dependency. Inject it (Khorikov ch. 8). |
| Huge setup | Too many responsibilities in one unit (GOOS ch. 20). |
| Can't swap a collaborator | The code constructs its own collaborators. Pass them in (Feathers ch. 4). |

- **Code with no seams at all:** pin its current behaviour with characterisation tests at the
  nearest seam, then refactor (Feathers ch. 13).
- **Never add test-only flags or branches to production code.** Add a seam instead
  (Khorikov ch. 11).

## 7. Checking the tests themselves

- **Mutation testing finds untested behaviour; don't track it as a score.** Mutate changed
  lines, and add a test for each surviving mutant that matters
  ([Petrović et al., Google 2021](https://arxiv.org/abs/2102.11378)).
- **Sabotage once when writing a medium test.** Remove the constraint or rollback it guards and
  watch the test fail.
- **No test maximises every desirable property.** Choose which ones each layer needs
  ([Beck, Test Desiderata](https://testdesiderata.com)).
