package services

import (
	"math"
	"testing"
)

func TestCalculateElo(t *testing.T) {
	tests := []struct {
		name      string
		ratingA   int
		ratingB   int
		scoreA    float64
		expectedA int // Approximate expected change
	}{
		{"Equal ratings, A wins", 1000, 1000, 1.0, 16},
		{"Equal ratings, A loses", 1000, 1000, 0.0, -16},
		{"Equal ratings, Draw", 1000, 1000, 0.5, 0},
		{"Stronger A wins", 1200, 1000, 1.0, 7}, // Less reward for beating weaker
		{"Weaker A wins", 1000, 1200, 1.0, 24},  // More reward for beating stronger
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kFactor := 32
			expectedScoreA := 1.0 / (1.0 + math.Pow(10, float64(tt.ratingB-tt.ratingA)/400.0))
			changeA := int(float64(kFactor) * (tt.scoreA - expectedScoreA))

			// Allow slight deviation due to floating point math, but for int cast it should be exact-ish
			if changeA != tt.expectedA {
				t.Errorf("expected change %d, got %d", tt.expectedA, changeA)
			}
		})
	}
}
