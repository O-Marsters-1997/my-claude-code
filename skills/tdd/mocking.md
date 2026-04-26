# When to Mock

Mock at **system boundaries** only:

- External APIs (payment, email, etc.)
- Databases (sometimes - prefer test DB)
- Time/randomness
- File system (sometimes)

Don't mock:

- Your own classes/modules
- Internal collaborators
- Anything you control

## Designing for Mockability

At system boundaries, design interfaces that are easy to mock:

**1. Use dependency injection**

Pass external dependencies in rather than creating them internally:

### TypeScript

```typescript
// Easy to mock
function processPayment(order, paymentClient) {
  return paymentClient.charge(order.total);
}

// Hard to mock
function processPayment(order) {
  const client = new StripeClient(process.env.STRIPE_KEY);
  return client.charge(order.total);
}
```

### Go

```go
// Easy to mock — takes interface
type PaymentClient interface {
    Charge(amount int) error
}

func ProcessPayment(order Order, client PaymentClient) error {
    return client.Charge(order.Total)
}

// Mock in tests
type MockPaymentClient struct {
    ChargeFunc func(amount int) error
}
func (m *MockPaymentClient) Charge(amount int) error { return m.ChargeFunc(amount) }

// Hard to mock — creates dependency internally
func ProcessPayment(order Order) error {
    client := stripe.New(os.Getenv("STRIPE_KEY"))
    return client.Charge(order.Total)
}
```

**2. Prefer per-operation methods over generic dispatchers**

Give each external operation its own named method/function. One generic `do(endpoint, opts)` forces conditional logic in mocks — every test has to decide what to return based on inspecting the arguments.

### TypeScript

```typescript
// GOOD: Each function is independently mockable
const api = {
  getUser: (id) => fetch(`/users/${id}`),
  getOrders: (userId) => fetch(`/users/${userId}/orders`),
  createOrder: (data) => fetch('/orders', { method: 'POST', body: data }),
};

// BAD: Mocking requires conditional logic inside the mock
const api = {
  fetch: (endpoint, options) => fetch(endpoint, options),
};
```

### Go

```go
// GOOD: Interface with named methods — each mock returns one specific shape
type UserAPI interface {
    GetUser(id string) (*User, error)
    GetOrders(userID string) ([]Order, error)
    CreateOrder(data OrderData) (*Order, error)
}

// BAD: Generic dispatcher — mock must inspect endpoint string to decide what to return
type GenericAPI interface {
    Do(endpoint string, opts map[string]any) ([]byte, error)
}
```

Named-method interfaces mean:
- Each mock returns one specific shape
- No conditional logic in test setup
- Easier to see which operations a test exercises
- Type safety per operation
