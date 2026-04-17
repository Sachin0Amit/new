package vector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCosineSimilarity(t *testing.T) {
	metric := &CosineSimilarity{}

	cases := []struct {
		a, b     []float32
		expected float32
	}{
		{[]float32{1, 0, 0}, []float32{1, 0, 0}, 1.0},
		{[]float32{1, 0, 0}, []float32{0, 1, 0}, 0.0},
		{[]float32{1, 1, 0}, []float32{1, 1, 0}, 1.0},
		{[]float32{1, 1, 1}, []float32{-1, -1, -1}, -1.0},
	}

	for _, tc := range cases {
		score, err := metric.Compare(tc.a, tc.b)
		assert.NoError(t, err)
		assert.InDelta(t, tc.expected, score, 0.001)
	}
}

func TestEuclideanDistance(t *testing.T) {
	metric := &EuclideanDistance{}

	cases := []struct {
		a, b     []float32
		expected float32
	}{
		{[]float32{0, 0}, []float32{3, 4}, 5.0}, // Pythagorean triple
		{[]float32{1, 1}, []float32{1, 1}, 0.0},
	}

	for _, tc := range cases {
		score, err := metric.Compare(tc.a, tc.b)
		assert.NoError(t, err)
		assert.InDelta(t, tc.expected, score, 0.001)
	}
}

func TestSearch(t *testing.T) {
	query := []float32{1, 0, 0}
	candidates := [][]float32{
		{0, 1, 0},
		{1, 0.1, 0},
		{0.5, 0.5, 0},
	}

	metric := &CosineSimilarity{}
	indices, scores, err := Search(query, candidates, metric)
	
	assert.NoError(t, err)
	assert.Equal(t, 1, indices[0], "Best match should be index 1")
	assert.Greater(t, scores[0], scores[1])
}
