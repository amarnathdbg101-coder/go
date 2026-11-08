package gogeohash

import "errors"

var (
	// ErrInvalidCoordinates indicates latitude not in [-90, 90] or longitude not in [-180, 180].
	ErrInvalidCoordinates = errors.New("gogeohash: coordinates out of valid range")

	// ErrInvalidPrecision indicates precision not in range [1, 12].
	ErrInvalidPrecision = errors.New("gogeohash: precision must be between 1 and 12")

	// ErrInvalidGeohash indicates a geohash string containing characters outside the 32-char alphabet.
	ErrInvalidGeohash = errors.New("gogeohash: invalid geohash character or empty string")
)
