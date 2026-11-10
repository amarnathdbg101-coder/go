# go-pricing 🏷️

> **Enterprise Rule-Based Dynamic Pricing, Volume Breaks, Coupon Apportioning, Margin Floor Clamps, and Indian GST Tax Engine for Go (Go 1.23+).**  
> Eliminates precision drift by using integer paise math and automatically computes CGST/SGST/IGST tax breakdowns for compliant invoices.

---

## 🚀 Key Features

1. **Zero-Float-Drift Financial Integrity**:
   - Stores all monetary amounts in 64-bit integer paise (e.g. `10000` = ₹100.00).
2. **Indian GST Compliance**:
   - **Intra-State**: Splits tax 50/50 into CGST and SGST with exact remainder allocation (zero fractional paise loss).
   - **Inter-State**: Automatically computes IGST.
   - **Tax-Inclusive / Exclusive**: Seamlessly back-calculates taxable value from MRP prices.
3. **Volume Discount Brackets**:
   - Quantity-based price breaks (e.g. 10+ items at ₹80 instead of ₹100).
4. **Margin Floor Protection**:
   - `WithMinMargin` ensures promotional discounts never sell below cost price plus required profit margin.
5. **Proportional Coupon Apportioning**:
   - Distributes order-level coupons across line items proportionally.
6. **Zero Dependencies**:
   - 100% pure Go standard library (`math`, `sort`, `sync`).

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-pricing"
)

func main() {
	engine := pricing.New(
		pricing.WithInterState(false), // Intra-state (CGST + SGST)
		pricing.WithVolumeTier("item_oil", []pricing.VolumeTier{
			{MinQty: 10, UnitPrice: 9000}, // ₹90 if >= 10 pcs
		}),
	)

	items := []pricing.Item{
		{
			ID:             "item_oil",
			Name:           "Refined Oil 1L",
			Quantity:       10,
			BasePrice:      10000, // ₹100.00 in paise
			TaxRate:        0.05,  // 5% GST
			IsTaxInclusive: false,
		},
	}

	invoice, _ := engine.Calculate(items, "")
	fmt.Printf("SubTotal: ₹%.2f\n", float64(invoice.SubTotal)/100)
	fmt.Printf("Discount: ₹%.2f\n", float64(invoice.DiscountTotal)/100)
	fmt.Printf("CGST: ₹%.2f, SGST: ₹%.2f\n", float64(invoice.CGST)/100, float64(invoice.SGST)/100)
	fmt.Printf("Grand Total: ₹%.2f\n", float64(invoice.GrandTotal)/100)
}
```
