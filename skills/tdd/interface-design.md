# Interface Design for Testability

Good interfaces make testing natural:

1. **Accept dependencies, don't create them**

   ### TypeScript
   ```typescript
   // Testable
   function processOrder(order, paymentGateway) {}

   // Hard to test
   function processOrder(order) {
     const gateway = new StripeGateway();
   }
   ```

   ### Go
   ```go
   // Testable
   func ProcessOrder(order Order, gateway PaymentGateway) error {}

   // Hard to test
   func ProcessOrder(order Order) error {
       gateway := stripe.New()
       ...
   }
   ```

2. **Return results, don't produce side effects**

   ### TypeScript
   ```typescript
   // Testable
   function calculateDiscount(cart): Discount {}

   // Hard to test
   function applyDiscount(cart): void {
     cart.total -= discount;
   }
   ```

   ### Go
   ```go
   // Testable
   func CalculateDiscount(cart Cart) Discount {}

   // Hard to test
   func ApplyDiscount(cart *Cart) {
       cart.Total -= discount
   }
   ```

3. **Small surface area**
   - Fewer methods = fewer tests needed
   - Fewer params = simpler test setup
