package gosafecast

import (
	"math"
)

// Signed represents all signed integer types.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned represents all unsigned integer types.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer represents any integer type.
type Integer interface {
	Signed | Unsigned
}

// Float represents standard floating point types.
type Float interface {
	~float32 | ~float64
}

// Number represents all built-in numeric types.
type Number interface {
	Integer | Float
}

// Must returns the cast value or panics if err != nil.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// ToInt8 safely casts any numeric type to int8 with bounds checking.
func ToInt8[T Number](v T) (int8, error) {
	val := float64(v)
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0, ErrOverflow
	}
	if val < math.MinInt8 {
		return 0, ErrUnderflow
	}
	if val > math.MaxInt8 {
		return 0, ErrOverflow
	}
	return int8(v), nil
}

// ToInt16 safely casts any numeric type to int16 with bounds checking.
func ToInt16[T Number](v T) (int16, error) {
	val := float64(v)
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0, ErrOverflow
	}
	if val < math.MinInt16 {
		return 0, ErrUnderflow
	}
	if val > math.MaxInt16 {
		return 0, ErrOverflow
	}
	return int16(v), nil
}

// ToInt32 safely casts any numeric type to int32 with bounds checking.
func ToInt32[T Number](v T) (int32, error) {
	val := float64(v)
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0, ErrOverflow
	}
	if val < math.MinInt32 {
		return 0, ErrUnderflow
	}
	if val > math.MaxInt32 {
		return 0, ErrOverflow
	}
	return int32(v), nil
}

// ToInt64 safely casts any numeric type to int64 with bounds checking.
func ToInt64[T Number](v T) (int64, error) {
	switch val := any(v).(type) {
	case uint64:
		if val > math.MaxInt64 {
			return 0, ErrOverflow
		}
		return int64(val), nil
	case uint:
		if uint64(val) > math.MaxInt64 {
			return 0, ErrOverflow
		}
		return int64(val), nil
	case uintptr:
		if uint64(val) > math.MaxInt64 {
			return 0, ErrOverflow
		}
		return int64(val), nil
	case float32:
		f := float64(val)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, ErrOverflow
		}
		if f < math.MinInt64 {
			return 0, ErrUnderflow
		}
		if f > math.MaxInt64 {
			return 0, ErrOverflow
		}
		return int64(val), nil
	case float64:
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return 0, ErrOverflow
		}
		if val < math.MinInt64 {
			return 0, ErrUnderflow
		}
		if val > math.MaxInt64 {
			return 0, ErrOverflow
		}
		return int64(val), nil
	default:
		// Fits within int64 naturally (int8, int16, int32, int, uint8, uint16, uint32)
		return int64(v), nil
	}
}

// ToInt safely casts any numeric type to architecture-dependent int.
func ToInt[T Number](v T) (int, error) {
	n, err := ToInt64(v)
	if err != nil {
		return 0, err
	}
	if n < math.MinInt {
		return 0, ErrUnderflow
	}
	if n > math.MaxInt {
		return 0, ErrOverflow
	}
	return int(n), nil
}

// ToUint8 safely casts any numeric type to uint8 with bounds checking.
func ToUint8[T Number](v T) (uint8, error) {
	val := float64(v)
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0, ErrOverflow
	}
	if val < 0 {
		return 0, ErrUnderflow
	}
	if val > math.MaxUint8 {
		return 0, ErrOverflow
	}
	return uint8(v), nil
}

// ToUint16 safely casts any numeric type to uint16 with bounds checking.
func ToUint16[T Number](v T) (uint16, error) {
	val := float64(v)
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0, ErrOverflow
	}
	if val < 0 {
		return 0, ErrUnderflow
	}
	if val > math.MaxUint16 {
		return 0, ErrOverflow
	}
	return uint16(v), nil
}

// ToUint32 safely casts any numeric type to uint32 with bounds checking.
func ToUint32[T Number](v T) (uint32, error) {
	val := float64(v)
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0, ErrOverflow
	}
	if val < 0 {
		return 0, ErrUnderflow
	}
	if val > math.MaxUint32 {
		return 0, ErrOverflow
	}
	return uint32(v), nil
}

// ToUint64 safely casts any numeric type to uint64 with bounds checking.
func ToUint64[T Number](v T) (uint64, error) {
	switch val := any(v).(type) {
	case int:
		if val < 0 {
			return 0, ErrUnderflow
		}
		return uint64(val), nil
	case int8:
		if val < 0 {
			return 0, ErrUnderflow
		}
		return uint64(val), nil
	case int16:
		if val < 0 {
			return 0, ErrUnderflow
		}
		return uint64(val), nil
	case int32:
		if val < 0 {
			return 0, ErrUnderflow
		}
		return uint64(val), nil
	case int64:
		if val < 0 {
			return 0, ErrUnderflow
		}
		return uint64(val), nil
	case float32:
		f := float64(val)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, ErrOverflow
		}
		if f < 0 {
			return 0, ErrUnderflow
		}
		if f > math.MaxUint64 {
			return 0, ErrOverflow
		}
		return uint64(val), nil
	case float64:
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return 0, ErrOverflow
		}
		if val < 0 {
			return 0, ErrUnderflow
		}
		if val > math.MaxUint64 {
			return 0, ErrOverflow
		}
		return uint64(val), nil
	default:
		// Fits within uint64 naturally
		return uint64(v), nil
	}
}

// ToUint safely casts any numeric type to architecture-dependent uint.
func ToUint[T Number](v T) (uint, error) {
	n, err := ToUint64(v)
	if err != nil {
		return 0, err
	}
	if n > math.MaxUint {
		return 0, ErrOverflow
	}
	return uint(n), nil
}
