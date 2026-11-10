package pricing_test

import (
	"errors"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-pricing"
)

func TestBasicIntraStateGST(t *testing.T) {
	eng := pricing.New()

	items := []pricing.Item{
		{
			ID:             "p1",
			Name:           "Basmati Rice",
			Quantity:       1,
			BasePrice:      10000, // ₹100.00 in paise
			TaxRate:        0.18,  // 18% GST
			IsTaxInclusive: false,
		},
	}

	inv, err := eng.Calculate(items, "")
	if err != nil {
		t.Fatalf("calculation failed: %v", err)
	}

	if inv.SubTotal != 10000 {
		t.Errorf("expected SubTotal 10000, got %d", inv.SubTotal)
	}
	if inv.TaxableAmount != 10000 {
		t.Errorf("expected TaxableAmount 10000, got %d", inv.TaxableAmount)
	}
	if inv.TotalTax != 1800 {
		t.Errorf("expected TotalTax 1800 (18%% of 100), got %d", inv.TotalTax)
	}
	if inv.CGST != 900 || inv.SGST != 900 {
		t.Errorf("expected CGST 900 and SGST 900, got CGST=%d SGST=%d", inv.CGST, inv.SGST)
	}
	if inv.GrandTotal != 11800 {
		t.Errorf("expected GrandTotal 11800 (₹118), got %d", inv.GrandTotal)
	}
}

func TestTaxInclusivePricing(t *testing.T) {
	eng := pricing.New()

	items := []pricing.Item{
		{
			ID:             "p2",
			Name:           "Branded Sunflower Oil",
			Quantity:       1,
			BasePrice:      11800, // ₹118.00 inclusive of 18% GST
			TaxRate:        0.18,
			IsTaxInclusive: true,
		},
	}

	inv, err := eng.Calculate(items, "")
	if err != nil {
		t.Fatalf("calculation failed: %v", err)
	}

	// Taxable should be ₹100 (10000 paise), Tax should be ₹18 (1800 paise)
	if inv.TaxableAmount != 10000 {
		t.Errorf("expected TaxableAmount 10000, got %d", inv.TaxableAmount)
	}
	if inv.TotalTax != 1800 {
		t.Errorf("expected TotalTax 1800, got %d", inv.TotalTax)
	}
	if inv.GrandTotal != 11800 {
		t.Errorf("expected GrandTotal 11800, got %d", inv.GrandTotal)
	}
}

func TestVolumeTierDiscount(t *testing.T) {
	eng := pricing.New(
		pricing.WithVolumeTier("p1", []pricing.VolumeTier{
			{MinQty: 10, UnitPrice: 8000}, // ₹80 if quantity >= 10
		}),
	)

	items := []pricing.Item{
		{
			ID:        "p1",
			Quantity:  10,
			BasePrice: 10000, // ₹100 base
			TaxRate:   0.0,
		},
	}

	inv, err := eng.Calculate(items, "")
	if err != nil {
		t.Fatalf("calculation failed: %v", err)
	}

	// 10 * 100 = 1000 subtotal, discount = 10 * 20 = 200, GrandTotal = 800
	if inv.SubTotal != 100000 {
		t.Errorf("expected SubTotal 100000, got %d", inv.SubTotal)
	}
	if inv.DiscountTotal != 20000 {
		t.Errorf("expected DiscountTotal 20000, got %d", inv.DiscountTotal)
	}
	if inv.GrandTotal != 80000 {
		t.Errorf("expected GrandTotal 80000, got %d", inv.GrandTotal)
	}
}

func TestMarginFloorViolation(t *testing.T) {
	eng := pricing.New(
		pricing.WithMinMargin(0.10), // 10% min profit margin
		pricing.WithVolumeTier("p_loss", []pricing.VolumeTier{
			{MinQty: 5, UnitPrice: 8500}, // sells at ₹85
		}),
	)

	items := []pricing.Item{
		{
			ID:        "p_loss",
			Quantity:  5,
			BasePrice: 10000,
			CostPrice: 9000, // Cost is ₹90, minimum 10% margin requires floor >= ₹99!
		},
	}

	_, err := eng.Calculate(items, "")
	if !errors.Is(err, pricing.ErrMarginFloorViolation) {
		t.Fatalf("expected ErrMarginFloorViolation, got %v", err)
	}
}

func TestInterStateIGST(t *testing.T) {
	eng := pricing.New(pricing.WithInterState(true))

	items := []pricing.Item{
		{
			ID:        "inter",
			Quantity:  1,
			BasePrice: 10000,
			TaxRate:   0.12, // 12% IGST
		},
	}

	inv, err := eng.Calculate(items, "")
	if err != nil {
		t.Fatalf("calculation failed: %v", err)
	}

	if inv.IGST != 1200 {
		t.Errorf("expected IGST 1200, got %d", inv.IGST)
	}
	if inv.CGST != 0 || inv.SGST != 0 {
		t.Errorf("CGST and SGST must be 0 for inter-state billing")
	}
}

