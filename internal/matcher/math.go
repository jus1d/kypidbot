package matcher

import "math"

func dotProduct(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func magnitude(v []float64) float64 {
	sum := 0.0
	for _, val := range v {
		sum += val * val
	}
	return math.Sqrt(sum)
}

func cosineSimilarity(a, b []float64) float64 {
	dot := dotProduct(a, b)
	magA := magnitude(a)
	magB := magnitude(b)

	if magA == 0 || magB == 0 {
		return 0
	}

	return dot / (magA * magB)
}
