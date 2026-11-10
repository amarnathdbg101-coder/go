# SPEC: go-pricing

## 1. Scope & Purpose
`go-pricing` computes promotional discounts, tiered pricing, minimum margin enforcement, and Indian GST (CGST, SGST, IGST) tax allocations for retail point-of-sale and e-commerce checkouts.

It does NOT handle payment gateway transactions or credit card processing.

## 2. Public API Contract
```go
type Item struct { ... }
type VolumeTier struct { ... }
type Coupon struct { ... }
type Invoice struct { ... }
type Engine struct { ... }

func New(opts ...Option) *Engine
func (e *Engine) Calculate(items []Item, couponCode string) (*Invoice, error)
```

## 3. Concurrency & Memory Guarantees
- Engine instances are thread-safe via `sync.RWMutex`.
- Calculations allocate temporary slices proportional to cart line size.
- Exact remainder handling guarantees `TaxableAmount + CGST + SGST + IGST == GrandTotal`.

## 4. How to Extend
- Add complex combo deals (Buy X get Y with condition Z).
- Add customer loyalty tier price modifiers.
