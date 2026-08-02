package tyuiop

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

	if !ValidateASCIIOnly(data) {
		return &PasswordStrength{Score: 0, CrackTime: "instantly", CrackSeconds: 0, Stats: &PasswordStats{}}
	}

	stats := AnalyzeBytes(data)

	patterns := findRepeats(data)
	patterns = append(patterns, findKeyboardPatterns(data)...)
	patterns = append(patterns, findCombinedPatterns(data)...)

	if common := findCommonPassword(data); common != nil {
		patterns = append(patterns, *common)
	}

	finalGuesses := calculateEntropy(data, patterns, stats)

	score := calculateScore(finalGuesses)
	crackTime, crackSeconds := calculateCrackTimeWithSeconds(finalGuesses, "bcrypt")
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
