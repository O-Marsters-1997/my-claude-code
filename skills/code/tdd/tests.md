# Good and Bad Tests

## Good Tests

**Integration-style**: Test through real interfaces, not mocks of internal parts.

### TypeScript

```typescript
// GOOD: Tests observable behavior
test("user can checkout with valid cart", async () => {
  const cart = createCart();
  cart.add(product);
  const result = await checkout(cart, paymentMethod);
  expect(result.status).toBe("confirmed");
});
```

### Go

```go
// GOOD: Tests observable behavior
func TestCheckout_WithValidCart(t *testing.T) {
    cart := NewCart()
    cart.Add(product)
    result, err := Checkout(cart, paymentMethod)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.Status != "confirmed" {
        t.Errorf("got status %q; want %q", result.Status, "confirmed")
    }
}
```

Characteristics:

- Tests behavior users/callers care about
- Uses public API only
- Survives internal refactors
- Describes WHAT, not HOW
- One logical assertion per test

## Bad Tests

**Implementation-detail tests**: Coupled to internal structure.

### TypeScript

```typescript
// BAD: Tests implementation details
test("checkout calls paymentService.process", async () => {
  const mockPayment = jest.mock(paymentService);
  await checkout(cart, payment);
  expect(mockPayment.process).toHaveBeenCalledWith(cart.total);
});
```

### Go

```go
// BAD: Tests internal call count
func TestCheckout_CallsGateway(t *testing.T) {
    called := 0
    gw := &MockGateway{ChargeFunc: func(amount int) error { called++; return nil }}
    checkout(cart, gw)
    if called != 1 {
        t.Errorf("Charge called %d times; want 1", called)
    }
}
```

Red flags:

- Mocking internal collaborators
- Testing private methods
- Asserting on call counts/order
- Test breaks when refactoring without behavior change
- Test name describes HOW not WHAT
- Verifying through external means instead of interface

## Bypass vs Interface

### TypeScript

```typescript
// BAD: Bypasses interface to verify
test("createUser saves to database", async () => {
  await createUser({ name: "Alice" });
  const row = await db.query("SELECT * FROM users WHERE name = ?", ["Alice"]);
  expect(row).toBeDefined();
});

// GOOD: Verifies through interface
test("createUser makes user retrievable", async () => {
  const user = await createUser({ name: "Alice" });
  const retrieved = await getUser(user.id);
  expect(retrieved.name).toBe("Alice");
});
```

### Go

```go
// BAD: Bypasses interface to verify
func TestCreateUser_SavesToDB(t *testing.T) {
    CreateUser(User{Name: "Alice"})
    row := db.QueryRow("SELECT * FROM users WHERE name = ?", "Alice")
    if row == nil {
        t.Fatal("expected row")
    }
}

// GOOD: Verifies through interface
func TestCreateUser_MakesUserRetrievable(t *testing.T) {
    user, err := CreateUser(User{Name: "Alice"})
    if err != nil {
        t.Fatalf("CreateUser failed: %v", err)
    }
    got, err := GetUser(user.ID)
    if err != nil {
        t.Fatalf("GetUser failed: %v", err)
    }
    if got.Name != "Alice" {
        t.Errorf("got name %q; want %q", got.Name, "Alice")
    }
}
```
