package pricing

// Item represents a product line entry before price calculation.
type Item struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	HSN            string  `json:"hsn,omitempty"`
	Quantity       int     `json:"quantity"`
	BasePrice      int64   `json:"base_price"` // in paise / cents
	CostPrice      int64   `json:"cost_price"` // in paise / cents
	TaxRate        float64 `json:"tax_rate"`   // e.g. 0.18 for 18% GST
	IsTaxInclusive bool    `json:"is_tax_inclusive"`
}

// VolumeTier specifies quantity-based unit price discounts.
type VolumeTier struct {
	MinQty    int   `json:"min_qty"`
	UnitPrice int64 `json:"unit_price"`
}

// Coupon defines order-level promo code discounts.
type Coupon struct {
	Code         string  `json:"code"`
	MinCartValue int64   `json:"min_cart_value"`
	DiscountPct  float64 `json:"discount_pct"` // e.g. 0.10 for 10% off
	MaxDiscount  int64   `json:"max_discount"`  // in paise
}

// InvoiceLine details the computed pricing, discounts, and taxes for a line item.
type InvoiceLine struct {
	Item           Item    `json:"item"`
	OriginalTotal  int64   `json:"original_total"`
	DiscountAmount int64   `json:"discount_amount"`
	TaxableAmount  int64   `json:"taxable_amount"`
	CGST           int64   `json:"cgst"`
	SGST           int64   `json:"sgst"`
	IGST           int64   `json:"igst"`
	TotalTax       int64   `json:"total_tax"`
	LineTotal      int64   `json:"line_total"`
}

// Invoice contains the comprehensive pricing summary and audit breakdown.
type Invoice struct {
	SubTotal       int64         `json:"sub_total"`
	DiscountTotal  int64         `json:"discount_total"`
	TaxableAmount  int64         `json:"taxable_amount"`
	CGST           int64         `json:"cgst"`
	SGST           int64         `json:"sgst"`
	IGST           int64         `json:"igst"`
	TotalTax       int64         `json:"total_tax"`
	GrandTotal     int64         `json:"grand_total"`
	Lines          []InvoiceLine `json:"lines"`
	AppliedCoupon  string        `json:"applied_coupon,omitempty"`
	IsInterState   bool          `json:"is_inter_state"`
}
