package tyuiop

import (
	"math"
)

func AnalyzeBytes(data []byte) *PasswordStats {
	var stats PasswordStats
	stats.Length = len(data)

	var seen [256]bool
	uniqueCount := 0
	maxRepeat := 0
	currentRepeat := 1

	for i := 0; i < len(data); i++ {
		b := data[i]
		switch {
		case b >= '0' && b <= '9':
			stats.Digits++
		case b >= 'a' && b <= 'z':
			stats.Lowers++
		case b >= 'A' && b <= 'Z':
			stats.Uppers++
		default:
			stats.Specials++
		}

		if !seen[b] {
			seen[b] = true
			uniqueCount++
		}

		if i > 0 && b == data[i-1] {
			currentRepeat++
			if currentRepeat > maxRepeat {
				maxRepeat = currentRepeat
			}
		} else {
			currentRepeat = 1
		}
	}

	stats.UniqueChar = uniqueCount
	stats.MaxRepeat = maxRepeat
	return &stats
}

func AnalyzeLocalStrength(data []byte) *PasswordStrength {
	if len(data) == 0 {
		return &PasswordStrength{Score: 0, CrackTime: "instantly", CrackSeconds: 0, Stats: &PasswordStats{}}
	}

	stats := AnalyzeBytes(data)
	patterns := append(findRepeats(data), findKeyboardPatterns(data)...)

	if common := findCommonPassword(data); common != nil {
		patterns = append(patterns, *common)
	}

	possibleChars := calculatePossibleChars(stats)
	baseGuesses := math.Pow(float64(possibleChars), float64(stats.Length))

	finalGuesses := applyPenalties(baseGuesses, patterns, stats)

	diversityPenalty := 1.0
	charSets := 0
	if stats.Digits > 0 {
		charSets++
	}
	if stats.Lowers > 0 {
		charSets++
	}
	if stats.Uppers > 0 {
		charSets++
	}
	if stats.Specials > 0 {
		charSets++
	}

	if charSets == 1 {
		diversityPenalty = 0.0001
	} else if charSets == 2 {
		diversityPenalty = 0.01
	}

	finalGuesses *= diversityPenalty

	score := calculateScore(finalGuesses)
	crackTime, crackSeconds := calculateCrackTimeWithSeconds(finalGuesses)
	feedback := generateFeedback(stats, patterns, score)

	return &PasswordStrength{
		Score:        score,
		Guesses:      finalGuesses,
		CrackTime:    crackTime,
		CrackSeconds: crackSeconds,
		Patterns:     patterns,
		Stats:        stats,
		Feedback:     feedback,
	}
}
