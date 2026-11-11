# go-geodistance/geohash

High-precision Geohash encoding, decoding, bounding box, and 8-neighbor proximity search for Go.

---

## Features
- **Geohash Encoding & Decoding**: Converts GPS latitude and longitude into standard Base32 geohashes and back.
- **8-Neighbor Proximity Search**: Instantly calculates the 8 surrounding cells (`North`, `South`, `East`, `West`, `NE`, `NW`, `SE`, `SW`) for radius and proximity queries.
- **Bounding Box Calculation**: Retrieves spatial boundaries (`MinLat`, `MaxLat`, `MinLon`, `MaxLon`) for map viewport queries.
- **Haversine Distance**: Computes great-circle distances between geographic coordinates in meters and kilometers.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"

	gogeohash "github.com/amarnathdbg101-coder/go/go-geodistance/geohash"
)

func main() {
	// Taj Mahal GPS coordinates
	lat, lon := 27.1751, 78.0421

	// 1. Encode with precision 7 (~150m accuracy)
	hash, err := gogeohash.Encode(lat, lon, 7)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Geohash: %s\n", hash)

	// 2. Decode back to center coordinates
	decLat, decLon, _ := gogeohash.Decode(hash)
	fmt.Printf("Decoded: (%.4f, %.4f)\n", decLat, decLon)

	// 3. Find 8 surrounding neighbors for hyperlocal search
	neighbors, _ := gogeohash.GetNeighbors(hash)
	fmt.Println("Adjacent neighbor cells:")
	for _, n := range neighbors.Slice() {
		fmt.Printf("- %s\n", n)
	}

	// 4. Calculate distance
	distMeters := gogeohash.Haversine(19.0760, 72.8777, 28.7041, 77.1025)
	fmt.Printf("Mumbai to Delhi: %.1f km\n", distMeters/1000.0)
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
