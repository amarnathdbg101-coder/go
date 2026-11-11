# Technical Specification: go-geodistance/geohash

## 1. Domain & Purpose
`go-geodistance/geohash` is a zero-dependency, high-precision geospatial encoding, decoding, bounding box, and 8-neighbor proximity search engine for Go. It implements the standard Gustavo Niemeyer 32-character geocoding system, powering location-based queries, geospatial indexing in Redis/PostgreSQL, delivery tracking, and proximity searches.

## 2. Architecture & Design Principles
- **Standard Base32 Alphabet**: Employs `"0123456789bcdefghjkmnpqrstuvwxyz"` with pre-computed lookup tables for $O(1)$ decoding.
- **Interleaved Binary Spatial Partitioning**: Interleaves longitude (even bits) and latitude (odd bits) to produce a hierarchical spatial prefix index.
- **Proximity Neighborhood Discovery**: Computes the 8 surrounding cardinal and diagonal cells (`North`, `South`, `East`, `West`, `NE`, `NW`, `SE`, `SW`) with automatic longitude wrapping and polar latitude clamping.
- **Spherical Haversine Metric**: Built-in great-circle spherical distance calculator (meters/km).
- **Zero Heavy Dependencies**: Pure Go standard library (`math`, `strings`).

## 3. API Surface
- `type Box struct { MinLat, MaxLat, MinLon, MaxLon float64 }`
- `func (b Box) Center() (lat, lon float64)`
- `type Neighbors struct { North, NorthEast, East, SouthEast, South, SouthWest, West, NorthWest string }`
- `func (n Neighbors) Slice() []string`
- `func Encode(lat, lon float64, precision int) (string, error)`
- `func Decode(hash string) (lat, lon float64, err error)`
- `func BoundingBox(hash string) (Box, error)`
- `func GetNeighbors(hash string) (*Neighbors, error)`
- `func Haversine(lat1, lon1, lat2, lon2 float64) float64`

## 4. Error Handling
- `ErrInvalidCoordinates`: Latitude not in `[-90, 90]` or longitude not in `[-180, 180]`.
- `ErrInvalidPrecision`: Precision outside range `[1, 12]`.
- `ErrInvalidGeohash`: Invalid geohash character or empty string.

## 5. Verification & Testing
- Unit tests verify famous landmarks (Taj Mahal, Statue of Liberty, Sydney Opera House), decoded coordinate tolerances, 8-neighbor uniqueness, Haversine distance accuracy (Mumbai to Delhi), and invalid character rejection.
