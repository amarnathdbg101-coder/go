package gogeohash

import (
	"fmt"
	"math"
	"strings"
)

const (
	base32Alphabet = "0123456789bcdefghjkmnpqrstuvwxyz"
	earthRadiusM   = 6371000.0 // Mean Earth radius in meters
)

var (
	base32Lookup [256]int
)

func init() {
	for i := range base32Lookup {
		base32Lookup[i] = -1
	}
	for i, c := range base32Alphabet {
		base32Lookup[c] = i
	}
}

// Box represents a geographic bounding box [min, max].
type Box struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

// Center calculates the center coordinates of the bounding box.
func (b Box) Center() (lat, lon float64) {
	return (b.MinLat + b.MaxLat) / 2.0, (b.MinLon + b.MaxLon) / 2.0
}

// Neighbors holds the 8 adjacent geohash cells around a central cell.
type Neighbors struct {
	North     string
	NorthEast string
	East      string
	SouthEast string
	South     string
	SouthWest string
	West      string
	NorthWest string
}

// Slice returns all 8 neighbor hashes as a slice.
func (n Neighbors) Slice() []string {
	return []string{
		n.North, n.NorthEast, n.East, n.SouthEast,
		n.South, n.SouthWest, n.West, n.NorthWest,
	}
}

// Encode converts latitude and longitude into a geohash string of desired precision.
func Encode(lat, lon float64, precision int) (string, error) {
	if lat < -90.0 || lat > 90.0 || lon < -180.0 || lon > 180.0 {
		return "", ErrInvalidCoordinates
	}
	if precision <= 0 || precision > 12 {
		return "", ErrInvalidPrecision
	}

	latInterval := [2]float64{-90.0, 90.0}
	lonInterval := [2]float64{-180.0, 180.0}

	var geohash strings.Builder
	geohash.Grow(precision)

	isEven := true
	bit := 0
	ch := 0

	for geohash.Len() < precision {
		if isEven {
			mid := (lonInterval[0] + lonInterval[1]) / 2.0
			if lon >= mid {
				ch |= (1 << (4 - bit))
				lonInterval[0] = mid
			} else {
				lonInterval[1] = mid
			}
		} else {
			mid := (latInterval[0] + latInterval[1]) / 2.0
			if lat >= mid {
				ch |= (1 << (4 - bit))
				latInterval[0] = mid
			} else {
				latInterval[1] = mid
			}
		}

		isEven = !isEven
		if bit < 4 {
			bit++
		} else {
			geohash.WriteByte(base32Alphabet[ch])
			bit = 0
			ch = 0
		}
	}

	return geohash.String(), nil
}

// BoundingBox computes the latitude/longitude boundary limits for a geohash string.
func BoundingBox(hash string) (Box, error) {
	hash = strings.ToLower(strings.TrimSpace(hash))
	if len(hash) == 0 {
		return Box{}, ErrInvalidGeohash
	}

	latInterval := [2]float64{-90.0, 90.0}
	lonInterval := [2]float64{-180.0, 180.0}
	isEven := true

	for i := 0; i < len(hash); i++ {
		val := base32Lookup[hash[i]]
		if val == -1 {
			return Box{}, fmt.Errorf("%w: invalid character %q", ErrInvalidGeohash, hash[i])
		}

		for mask := 16; mask > 0; mask >>= 1 {
			if isEven {
				mid := (lonInterval[0] + lonInterval[1]) / 2.0
				if (val & mask) != 0 {
					lonInterval[0] = mid
				} else {
					lonInterval[1] = mid
				}
			} else {
				mid := (latInterval[0] + latInterval[1]) / 2.0
				if (val & mask) != 0 {
					latInterval[0] = mid
				} else {
					latInterval[1] = mid
				}
			}
			isEven = !isEven
		}
	}

	return Box{
		MinLat: latInterval[0],
		MaxLat: latInterval[1],
		MinLon: lonInterval[0],
		MaxLon: lonInterval[1],
	}, nil
}

// Decode extracts the approximate center coordinates for a geohash string.
func Decode(hash string) (lat, lon float64, err error) {
	box, err := BoundingBox(hash)
	if err != nil {
		return 0, 0, err
	}
	lat, lon = box.Center()
	return lat, lon, nil
}

// GetNeighbors returns the 8 surrounding geohash cells at the same precision level.
func GetNeighbors(hash string) (*Neighbors, error) {
	box, err := BoundingBox(hash)
	if err != nil {
		return nil, err
	}

	lat, lon := box.Center()
	latHeight := box.MaxLat - box.MinLat
	lonWidth := box.MaxLon - box.MinLon
	precision := len(hash)

	encodeNeighbor := func(dLat, dLon float64) string {
		nLat := lat + dLat
		nLon := lon + dLon
		// Wrap around longitude
		if nLon > 180.0 {
			nLon -= 360.0
		} else if nLon < -180.0 {
			nLon += 360.0
		}
		// Clamp latitude
		if nLat > 90.0 {
			nLat = 90.0
		} else if nLat < -90.0 {
			nLat = -90.0
		}
		res, _ := Encode(nLat, nLon, precision)
		return res
	}

	return &Neighbors{
		North:     encodeNeighbor(latHeight, 0),
		NorthEast: encodeNeighbor(latHeight, lonWidth),
		East:      encodeNeighbor(0, lonWidth),
		SouthEast: encodeNeighbor(-latHeight, lonWidth),
		South:     encodeNeighbor(-latHeight, 0),
		SouthWest: encodeNeighbor(-latHeight, -lonWidth),
		West:      encodeNeighbor(0, -lonWidth),
		NorthWest: encodeNeighbor(latHeight, -lonWidth),
	}, nil
}

// Haversine calculates great-circle distance between two coordinates in meters.
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	radLat1 := lat1 * math.Pi / 180.0
	radLat2 := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(radLat1)*math.Cos(radLat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusM * c
}
