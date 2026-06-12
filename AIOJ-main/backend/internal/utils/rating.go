package utils

import "math"

const (
	DefaultUserRating    = 1000
	DefaultProblemRating = 800
	MinRating            = 100
	MaxRating            = 4000
)

// CalculateUserRatingUpdate computes the new user rating after solving a problem.
// Uses an Elo-like formula where:
//   - expected = probability of solving based on rating difference
//   - K factor scales with the rating gap (harder problems give more points)
//   - Solving problems below user rating gives diminishing returns
func CalculateUserRatingUpdate(userRating, problemRating int) int {
	if userRating <= 0 {
		userRating = DefaultUserRating
	}
	if problemRating <= 0 {
		problemRating = DefaultProblemRating
	}

	// 题目难度 <= 用户水平时不加分
	if problemRating <= userRating {
		return userRating
	}

	diff := float64(problemRating - userRating)

	// Expected solve probability: sigmoid based on rating difference
	// When problemRating == userRating, expected = 0.5
	// When problemRating >> userRating, expected → 0 (very hard)
	// When problemRating << userRating, expected → 1 (very easy)
	expected := 1.0 / (1.0 + math.Pow(10.0, diff/400.0))

	// K factor: base 32, scales up for harder problems
	// For very easy problems (expected > 0.9), K is very small
	k := 32.0
	if diff > 0 {
		// Harder problems: increase K proportionally
		k = 32.0 + diff*0.05
		if k > 64 {
			k = 64
		}
	} else {
		// Easier problems: reduce K
		k = 32.0 + diff*0.03
		if k < 4 {
			k = 4
		}
	}

	// actual = 1 (solved), so delta = K * (1 - expected)
	delta := k * (1.0 - expected)

	// Round to nearest integer
	newRating := userRating + int(math.Round(delta))

	// Clamp
	if newRating < MinRating {
		newRating = MinRating
	}
	if newRating > MaxRating {
		newRating = MaxRating
	}

	return newRating
}

// CalculateProblemRating estimates a problem's rating based on difficulty and acceptance rate.
// Used when admin hasn't explicitly set a rating.
func CalculateProblemRating(difficulty string, acceptRate float64) int {
	// Base rating from difficulty level
	var baseRating int
	switch difficulty {
	case "简单":
		baseRating = 800
	case "中等":
		baseRating = 1400
	case "困难":
		baseRating = 2000
	default:
		baseRating = 1200
	}

	// Adjust based on acceptance rate
	// Lower acceptance rate → higher rating
	if acceptRate > 0 && acceptRate <= 100 {
		// Scale: 80% accept → -200, 20% accept → +200
		adjustment := (50.0 - acceptRate) * 5.0
		baseRating += int(adjustment)
	}

	// Clamp
	if baseRating < MinRating {
		baseRating = MinRating
	}
	if baseRating > MaxRating {
		baseRating = MaxRating
	}

	return baseRating
}
