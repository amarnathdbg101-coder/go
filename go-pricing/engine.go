package pricing

import (
	"math"
	"sort"
	"sync"
)

// Engine calculates multi-tier discounts, profit margin constraints, and GST tax breakdowns.
type Engine struct {
	cfg Config
	mu  sync.RWMutex
}

// New creates a new pricing engine instance with the specified options.
func New(opts ...Option) *Engine {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Engine{
		cfg: cfg,
	}
}

// Calculate computes line-item discounts, coupon apportioning, and GST tax breakdown.
func (e *Engine) Calculate(items []Item, couponCode string) (*Invoice, error) {
	if len(items) == 0 {
		return nil, ErrEmptyCart
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	var subTotal int64
	lines := make([]InvoiceLine, len(items))

	// 1. Process line-item pricing and volume discounts
	for i, item := range items {
		if item.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}
		if item.BasePrice < 0 {
			return nil, ErrInvalidPrice
		}

		unitPrice := item.BasePrice

		// Evaluate volume discount tiers
		if tiers, ok := e.cfg.VolumeTiers[item.ID]; ok && len(tiers) > 0 {
			// Sort tiers by MinQty descending
			sort.Slice(tiers, func(a, b int) bool {
				return tiers[a].MinQty > tiers[b].MinQty
			})
			for _, tier := range tiers {
				if item.Quantity >= tier.MinQty {
					if tier.UnitPrice < unitPrice {
						unitPrice = tier.UnitPrice
					}
					break
				}
			}
		}

		// Enforce minimum margin floor
		if e.cfg.MinMarginPercent > 0 && item.CostPrice > 0 {
			floor := int64(math.Ceil(float64(item.CostPrice) * (1.0 + e.cfg.MinMarginPercent)))
			if unitPrice < floor {
				return nil, ErrMarginFloorViolation
			}
		}

		origTotal := item.BasePrice * int64(item.Quantity)
		lineTotalDiscount := (item.BasePrice - unitPrice) * int64(item.Quantity)

		subTotal += origTotal

		lines[i] = InvoiceLine{
			Item:           item,
			OriginalTotal:  origTotal,
			DiscountAmount: lineTotalDiscount,
		}
	}

	// 2. Evaluate order-level coupon discount
	var couponDiscountTotal int64
	appliedCouponCode := ""

	if couponCode != "" {
		coupon, exists := e.cfg.Coupons[couponCode]
		if !exists {
			return nil, ErrCouponRequirementNotMet
		}
		if subTotal < coupon.MinCartValue {
			return nil, ErrCouponRequirementNotMet
		}

		discount := int64(math.Round(float64(subTotal) * coupon.DiscountPct))
		if coupon.MaxDiscount > 0 && discount > coupon.MaxDiscount {
			discount = coupon.MaxDiscount
		}
		couponDiscountTotal = discount
		appliedCouponCode = couponCode
	}

	// 3. Apportion coupon discount and calculate GST tax per line
	var totalTaxable, totalCGST, totalSGST, totalIGST, totalTax, grandTotal int64
	totalDiscount := couponDiscountTotal

	for i := range lines {
		line := &lines[i]
		totalDiscount += line.DiscountAmount

		// Proportional coupon discount apportioning
		var lineCouponShare int64
		if subTotal > 0 && couponDiscountTotal > 0 {
			lineCouponShare = int64(math.Round(float64(couponDiscountTotal) * (float64(line.OriginalTotal) / float64(subTotal))))
		}
		line.DiscountAmount += lineCouponShare

		netLinePrice := line.OriginalTotal - line.DiscountAmount
		if netLinePrice < 0 {
			netLinePrice = 0
		}

		// GST calculation
		var taxable, tax int64
		if line.Item.IsTaxInclusive {
			// Taxable = Net / (1 + Rate)
			taxable = int64(math.Round(float64(netLinePrice) / (1.0 + line.Item.TaxRate)))
			tax = netLinePrice - taxable
		} else {
			// Taxable = Net
			taxable = netLinePrice
			tax = int64(math.Round(float64(taxable) * line.Item.TaxRate))
		}

		line.TaxableAmount = taxable
		line.TotalTax = tax

		if e.cfg.IsInterState {
			line.IGST = tax
			line.CGST = 0
			line.SGST = 0
		} else {
			// Split intra-state into CGST 50% and SGST 50%
			cgst := tax / 2
			sgst := tax - cgst // Guarantees exact sum without rounding loss
			line.CGST = cgst
			line.SGST = sgst
			line.IGST = 0
		}

		line.LineTotal = taxable + tax

		totalTaxable += line.TaxableAmount
		totalCGST += line.CGST
		totalSGST += line.SGST
		totalIGST += line.IGST
		totalTax += line.TotalTax
		grandTotal += line.LineTotal
	}

	return &Invoice{
		SubTotal:       subTotal,
		DiscountTotal:  totalDiscount,
		TaxableAmount:  totalTaxable,
		CGST:           totalCGST,
		SGST:           totalSGST,
		IGST:           totalIGST,
		TotalTax:       totalTax,
		GrandTotal:     grandTotal,
		Lines:          lines,
		AppliedCoupon:  appliedCouponCode,
		IsInterState:   e.cfg.IsInterState,
	}, nil
}
