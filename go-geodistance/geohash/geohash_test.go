package gogeohash

import (
	"math"
	"strings"
	"testing"
)

func TestEncodeDecode_Landmarks(t *testing.T) {
	landmarks := []struct {
		name      string
		lat       float64
		lon       float64
		precision int
	}{
		{"Taj Mahal", 27.1751, 78.0421, 7},
		{"Statue of Liberty", 40.6892, -74.0445, 8},
		{"Sydney Opera House", -33.8568, 151.2153, 7},
		{"Greenwich Observatory", 51.4769, 0.0005, 9},
	}

	for _, lm := range landmarks {
		hash, err := Encode(lm.lat, lm.lon, lm.precision)
		if err != nil {
			t.Fatalf("%s encode failed: %v", lm.name, err)
		}

		if len(hash) != lm.precision {
			t.Errorf("%s: expected hash len %d, got %d", lm.name, lm.precision, len(hash))
		}

		decLat, decLon, err := Decode(hash)
		if err != nil {
			t.Fatalf("%s decode failed: %v", lm.name, err)
		}

		// Distance between original and center of geohash box should be small (within precision cell)
		distM := Haversine(lm.lat, lm.lon, decLat, decLon)
		box, _ := BoundingBox(hash)

		if lm.lat < box.MinLat || lm.lat > box.MaxLat || lm.lon < box.MinLon || lm.lon > box.MaxLon {
			t.Errorf("%s coordinate not inside bounding box", lm.name)
		}

		if distM > 1000.0 && lm.precision >= 7 {
			t.Errorf("%s decoded point too far: %.2f meters", lm.name, distM)
		}
	}
}

func TestNeighbors(t *testing.T) {
	// Geohash for central Paris: "u09tv"
	hash := "u09tv"
	neighbors, err := GetNeighbors(hash)
	if err != nil {
		t.Fatalf("GetNeighbors failed: %v", err)
	}

	slice := neighbors.Slice()
	if len(slice) != 8 {
		t.Fatalf("expected 8 neighbors, got %d", len(slice))
	}

	seen := make(map[string]bool)
	for _, n := range slice {
		if len(n) != len(hash) {
			t.Errorf("neighbor len %d != hash len %d", len(n), len(hash))
		}
		if n == hash {
			t.Errorf("neighbor should not equal center hash")
		}
		seen[n] = true
	}

	if len(seen) != 8 {
		t.Errorf("expected 8 unique neighbors, got %d", len(seen))
	}
}

func TestHaversineDistance(t *testing.T) {
	// Mumbai (19.0760, 72.8777) to Delhi (28.7041, 77.1025)
	// Approximate distance: ~1,148 km
	distMeters := Haversine(19.0760, 72.8777, 28.7041, 77.1025)
	distKm := distMeters / 1000.0

	if math.Abs(distKm-1148.0) > 20.0 {
		t.Errorf("expected ~1148km distance between Mumbai and Delhi, got %.2f km", distKm)
	}
}

func TestInvalidInputs(t *testing.T) {
	// Invalid Latitude > 90
	_, err := Encode(95.0, 70.0, 5)
	if err != ErrInvalidCoordinates {
		t.Errorf("expected ErrInvalidCoordinates, got %v", err)
	}

	// Invalid Precision
	_, err = Encode(20.0, 70.0, 0)
	if err != ErrInvalidPrecision {
		t.Errorf("expected ErrInvalidPrecision, got %v", err)
	}

	// Invalid Hash character (e.g. 'a', which is excluded from Geohash base32)
	_, _, err = Decode("u09tva")
	if err == nil || !strings.Contains(err.Error(), "invalid geohash character") {
		t.Errorf("expected error decoding character 'a', got %v", err)
	}
}
