package pricing

import "errors"

var (
	// ErrEmptyCart is returned when evaluating pricing on an empty cart.
	ErrEmptyCart = errors.New("pricing: cart contains no items")

	// ErrInvalidQuantity is returned when an item quantity is zero or negative.
	ErrInvalidQuantity = errors.New("pricing: item quantity must be greater than zero")

	// ErrMarginFloorViolation is returned when discounts push unit price below minimum allowable margin.
	ErrMarginFloorViolation = errors.New("pricing: price violates minimum profit margin floor")

	// ErrCouponRequirementNotMet is returned when a coupon min order threshold is not satisfied.
	ErrCouponRequirementNotMet = errors.New("pricing: cart value does not meet coupon threshold")

	// ErrInvalidPrice is returned when item price is negative.
	ErrInvalidPrice = errors.New("pricing: base price cannot be negative")
)
