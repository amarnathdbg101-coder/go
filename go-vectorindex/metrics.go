package vectorindex

import (
	"math"
)

// DotProduct calculates the dot product of two vectors.
func DotProduct(a, b []float32) float32 {
	var sum float32
	// 4-way loop unrolling for high-speed CPU instruction pipelining
	n := len(a)
	i := 0
	for ; i+3 < n; i += 4 {
		sum += a[i]*b[i] + a[i+1]*b[i+1] + a[i+2]*b[i+2] + a[i+3]*b[i+3]
	}
	for ; i < n; i++ {
		sum += a[i] * b[i]
	}
	return sum
}

// EuclideanDistance calculates the L2 Euclidean distance between two vectors.
func EuclideanDistance(a, b []float32) float32 {
	var sum float32
	n := len(a)
	i := 0
	for ; i+3 < n; i += 4 {
		d0 := a[i] - b[i]
		d1 := a[i+1] - b[i+1]
		d2 := a[i+2] - b[i+2]
		d3 := a[i+3] - b[i+3]
		sum += d0*d0 + d1*d1 + d2*d2 + d3*d3
	}
	for ; i < n; i++ {
		d := a[i] - b[i]
		sum += d * d
	}
	return float32(math.Sqrt(float64(sum)))
}

// CosineSimilarity calculates the cosine similarity between two vectors.
func CosineSimilarity(a, b []float32) (float32, error) {
	dot := DotProduct(a, b)
	magA := DotProduct(a, a)
	magB := DotProduct(b, b)

	if magA <= 0 || magB <= 0 {
		return 0, ErrZeroVector
	}

	denom := float32(math.Sqrt(float64(magA * magB)))
	return dot / denom, nil
}

// Normalize returns an L2 unit-norm normalized copy of the vector.
func Normalize(v []float32) ([]float32, error) {
	mag := DotProduct(v, v)
	if mag <= 0 {
		return nil, ErrZeroVector
	}
	scale := 1.0 / float32(math.Sqrt(float64(mag)))
	res := make([]float32, len(v))
	for i := range v {
		res[i] = v[i] * scale
	}
	return res, nil
}
