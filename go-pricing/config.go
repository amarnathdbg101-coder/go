package pricing

// Config holds global pricing engine settings and promotional rules.
type Config struct {
	// MinMarginPercent protects against selling below cost (e.g. 0.05 = 5% minimum profit margin).
	// Default is 0.0 (no margin floor enforcement).
	MinMarginPercent float64

	// IsInterState determines whether tax is split into CGST+SGST (intra-state) or IGST (inter-state).
	IsInterState bool

	// VolumeTiers maps item IDs to their quantity discount brackets.
	VolumeTiers map[string][]VolumeTier

	// Coupons maps promo codes to coupon discount structures.
	Coupons map[string]Coupon
}

// Option is a functional option for configuring the pricing engine.
type Option func(*Config)

// DefaultConfig returns resilient defaults for pricing calculations.
func DefaultConfig() Config {
	return Config{
		MinMarginPercent: 0.0,
		IsInterState:     false,
		VolumeTiers:      make(map[string][]VolumeTier),
		Coupons:          make(map[string]Coupon),
	}
}

// WithMinMargin sets the minimum allowable profit margin ratio above cost price.
func WithMinMargin(margin float64) Option {
	return func(c *Config) {
		if margin >= 0 {
			c.MinMarginPercent = margin
		}
	}
}

// WithInterState toggles IGST calculation for inter-state billing.
func WithInterState(interState bool) Option {
	return func(c *Config) {
		c.IsInterState = interState
	}
}

// WithVolumeTier registers a volume discount tier for an item.
func WithVolumeTier(itemID string, tiers []VolumeTier) Option {
	return func(c *Config) {
		c.VolumeTiers[itemID] = tiers
	}
}

// WithCoupon registers a promotional coupon code.
func WithCoupon(coupon Coupon) Option {
	return func(c *Config) {
		c.Coupons[coupon.Code] = coupon
	}
}
