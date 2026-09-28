# Refactor Candidates

After TDD cycle, look for:

- **Duplication** → Extract function/class
- **Long methods** → Break into private helpers (keep tests on public interface)
- **Shallow modules** → Combine or deepen
- **Feature envy** → Move logic to where data lives
- **Primitive obsession** → Introduce value objects
- **Existing code** the new code reveals as problematic

## Test Fixtures and Setup Hooks

Extract repeated setup into helpers. Don't repeat the same initialization verbatim across test cases.

Factory function + `beforeEach`:

```typescript
function makeUserService(overrides = {}) {
  const repo = { save: jest.fn(), findById: jest.fn(), ...overrides };
  return { service: new UserService(repo), repo };
}

describe("UserService", () => {
  let service, repo;
  beforeEach(() => ({ service, repo } = makeUserService()));

  it("creates user", () => { ... });
  it("rejects empty name", () => { ... });
});
```

## Group Tests That Share Dependencies

When cases share the same mocked service, DB, or fixture, group them under one parent scope. Cases inside the group should differ only in inputs and assertions. Setup steps may differ when a case genuinely needs different state — that's fine — but never copy-paste identical setup across siblings.

**Smell**: same `new Foo()` / `setupDB()` repeated verbatim in three separate test functions → extract to parent scope.

## What Not to Share

Never share **mutable state** across cases. Each test must be independent. Share a dependency only when it is genuinely read-only and identical between test cases. For everything else, such as if different for another test case, create a fresh instance.
