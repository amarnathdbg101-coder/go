package gosafecast

import (
	"errors"
	"math"
	"testing"
)

func TestToInt8(t *testing.T) {
	// Valid bounds
	val, err := ToInt8(127)
	if err != nil || val != 127 {
		t.Fatalf("expected 127, got %d, err %v", val, err)
	}

	val, err = ToInt8(-128)
	if err != nil || val != -128 {
		t.Fatalf("expected -128, got %d, err %v", val, err)
	}

	// Overflow
	_, err = ToInt8(128)
	if !errors.Is(err, ErrOverflow) {
		t.Errorf("expected ErrOverflow, got %v", err)
	}

	// Underflow
	_, err = ToInt8(-129)
	if !errors.Is(err, ErrUnderflow) {
		t.Errorf("expected ErrUnderflow, got %v", err)
	}
}

func TestToUint8(t *testing.T) {
	val, err := ToUint8(255)
	if err != nil || val != 255 {
		t.Fatalf("expected 255, got %d, err %v", val, err)
	}

	// Overflow
	_, err = ToUint8(256)
	if !errors.Is(err, ErrOverflow) {
		t.Errorf("expected ErrOverflow, got %v", err)
	}

	// Negative to unsigned underflow
	_, err = ToUint8(-1)
	if !errors.Is(err, ErrUnderflow) {
		t.Errorf("expected ErrUnderflow, got %v", err)
	}
}

func TestToInt32_And_ToUint32(t *testing.T) {
	val32, err := ToInt32(2147483647)
	if err != nil || val32 != 2147483647 {
		t.Fatalf("expected MaxInt32, got %d, err %v", val32, err)
	}

	_, err = ToInt32(int64(2147483648))
	if !errors.Is(err, ErrOverflow) {
		t.Errorf("expected ErrOverflow, got %v", err)
	}

	u32, err := ToUint32(uint64(4294967295))
	if err != nil || u32 != 4294967295 {
		t.Fatalf("expected MaxUint32, got %d, err %v", u32, err)
	}

	_, err = ToUint32(uint64(4294967296))
	if !errors.Is(err, ErrOverflow) {
		t.Errorf("expected ErrOverflow, got %v", err)
	}
}

func TestToInt64_And_ToUint64(t *testing.T) {
	// Uint64 to Int64 overflow check
	largeUint := uint64(math.MaxInt64) + 1
	_, err := ToInt64(largeUint)
	if !errors.Is(err, ErrOverflow) {
		t.Errorf("expected ErrOverflow for large uint64, got %v", err)
	}

	// Negative int64 to uint64 underflow check
	_, err = ToUint64(int64(-100))
	if !errors.Is(err, ErrUnderflow) {
		t.Errorf("expected ErrUnderflow for negative int64, got %v", err)
	}
}

func TestFloatConversions_And_SpecialValues(t *testing.T) {
	val, err := ToInt32(42.0)
	if err != nil || val != 42 {
		t.Fatalf("expected 42, got %d, err %v", val, err)
	}

	// NaN check
	_, err = ToInt(math.NaN())
	if !errors.Is(err, ErrOverflow) {
		t.Errorf("expected ErrOverflow for NaN, got %v", err)
	}

	// Inf check
	_, err = ToInt(math.Inf(1))
	if !errors.Is(err, ErrOverflow) {
		t.Errorf("expected ErrOverflow for +Inf, got %v", err)
	}
}

func TestMustHelper(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on Must error")
		}
	}()

	good := Must(ToInt8(50))
	if good != 50 {
		t.Errorf("expected 50, got %d", good)
	}

	// This should panic
	_ = Must(ToInt8(500))
}
