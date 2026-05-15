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

**TypeScript** — factory function + `beforeEach`:

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

**Go** — `t.Helper()` setup function + `t.Cleanup()`:

```go
func setupUserService(t *testing.T) (*UserService, *MockRepo) {
    t.Helper()
    repo := &MockRepo{}
    svc := NewUserService(repo)
    t.Cleanup(func() { repo.AssertExpectations(t) })
    return svc, repo
}

func TestUserService(t *testing.T) {
    svc, repo := setupUserService(t)

    t.Run("creates user", func(t *testing.T) {
        repo.SaveFunc = func(u User) error { return nil }
        _, err := svc.Create(User{Name: "Alice"})
        if err != nil { t.Fatalf("unexpected error: %v", err) }
    })

    t.Run("rejects empty name", func(t *testing.T) {
        _, err := svc.Create(User{Name: ""})
        if err == nil { t.Fatal("expected error for empty name") }
    })
}
```

Use `t.TempDir()` for temporary file paths — Go cleans these up automatically.

## Group Tests That Share Dependencies

When cases share the same mocked service, DB, or fixture, group them under one parent scope. Cases inside the group should differ only in inputs and assertions. Setup steps may differ when a case genuinely needs different state — that's fine — but never copy-paste identical setup across siblings.

**Smell**: same `new Foo()` / `setupDB()` repeated verbatim in three separate test functions → extract to parent scope.

**Table-driven** (Go) — ideal when cases share identical setup and differ only in args + assertion:

```go
func TestCreateUser(t *testing.T) {
    svc, _ := setupUserService(t)

    tests := []struct {
        name    string
        input   User
        wantErr bool
    }{
        {"valid user", User{Name: "Alice"}, false},
        {"empty name", User{Name: ""}, true},
        {"whitespace name", User{Name: "  "}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := svc.Create(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got err=%v; wantErr=%v", err, tt.wantErr)
            }
        })
    }
}
```

## What Not to Share

Never share **mutable state** across cases. Each test must be independent. Share a dependency only when it is genuinely read-only and identical between test cases. For everything else, such as if different for another test case, create a fresh instance.
