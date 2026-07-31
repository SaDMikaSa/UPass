package tyuiop

import (
	"bytes"
)

var specialChars = map[byte]bool{
	'!': true, '~': true, '`': true, '@': true, '"': true,
	'#': true, '$': true, ';': true, '%': true, '^': true,
	':': true, '&': true, '?': true, '*': true, '(': true,
	')': true, '-': true, '_': true, '=': true, '+': true,
	'[': true, ']': true, '{': true, '}': true, '\\': true,
	'|': true, ',': true, '.': true, '<': true, '>': true,
	'/': true, ' ': true, '\t': true, '\n': true,
}

func calculatePossibleChars(stats *PasswordStats) int {
	chars := 0
	if stats.Digits > 0 {
		chars += 10
	}
	if stats.Lowers > 0 {
		chars += 26
	}
	if stats.Uppers > 0 {
		chars += 26
	}
	if stats.Specials > 0 {
		chars += 34
	}
	if chars == 0 {
		return 1
	}
	return chars
}

func applyPenalties(guesses float64, patterns []Pattern, stats *PasswordStats) float64 {
	result := guesses
	for _, p := range patterns {
		switch p.Type {
		case "repeat":
			if p.Length == stats.Length {
				result *= 0.0001
			} else if p.Length >= 6 {
				result *= 0.1
			} else if p.Length >= 3 {
				result *= 0.5
			}
		case "keyboard":
			result *= 0.1
		case "common":
			result *= 0.001
		}
	}
	return result
}

func calculateScore(guesses float64) int {
	switch {
	case guesses < 1e4:
		return 0
	case guesses < 1e6:
		return 1
	case guesses < 1e8:
		return 2
	case guesses < 1e10:
		return 3
	default:
		return 4
	}
}

func calculateCrackTime(guesses float64) string {
	switch {
	case guesses < 1000:
		return "instantly"
	case guesses < 1000000:
		return "seconds"
	case guesses < 100000000:
		return "minutes"
	case guesses < 10000000000:
		return "hours"
	case guesses < 1000000000000:
		return "days"
	default:
		return "years"
	}
}

func findCommonPassword(data []byte) *Pattern {
	lowerData := bytes.ToLower(data)
	for _, common := range commonPasswordsBytes {
		if bytes.Contains(lowerData, common) {
			return &Pattern{
				Type:   "common",
				Length: len(common),
				Start:  bytes.Index(lowerData, common),
				Value:  common,
			}
		}
	}
	return nil
}

func findRepeats(data []byte) []Pattern {
	var patterns []Pattern
	if len(data) < 2 {
		return patterns
	}

	seen := make(map[int]bool)

	for blockSize := 1; blockSize <= len(data)/2; blockSize++ {
		if len(data)%blockSize != 0 {
			continue
		}
		firstBlock := data[:blockSize]
		isRepeating := true
		for j := blockSize; j < len(data); j += blockSize {
			for k := 0; k < blockSize; k++ {
				if data[j+k] != firstBlock[k] {
					isRepeating = false
					break
				}
			}
			if !isRepeating {
				break
			}
		}
		if isRepeating {
			key := 0*1000 + len(data)
			if !seen[key] {
				seen[key] = true
				patterns = append(patterns, Pattern{Type: "repeat", Length: len(data), Start: 0, Value: firstBlock})
			}
			break
		}
	}

	for i := 0; i < len(data); {
		j := i
		for j < len(data) && data[j] == data[i] {
			j++
		}
		repeatLen := j - i
		if repeatLen >= 3 {
			key := i*1000 + repeatLen
			if !seen[key] {
				seen[key] = true
				patterns = append(patterns, Pattern{Type: "repeat", Length: repeatLen, Start: i, Value: data[i:j]})
			}
		}
		i = j
	}
	return patterns
}

func findKeyboardPatterns(data []byte) []Pattern {
	var patterns []Pattern
	if len(data) < 3 {
		return patterns
	}

	keyboardRows := [][]byte{
		[]byte("qwertyuiop"), []byte("asdfghjkl"), []byte("zxcvbnm"), []byte("1234567890"),
	}

	seen := make(map[string]bool)
	lowerData := bytes.ToLower(data)

	for _, row := range keyboardRows {
		for length := 3; length <= len(row); length++ {
			for i := 0; i <= len(row)-length; i++ {
				subseq := row[i : i+length]
				if bytes.Contains(lowerData, subseq) {
					val := string(subseq)
					if !seen[val] {
						seen[val] = true
						patterns = append(patterns, Pattern{Type: "keyboard", Length: length, Start: bytes.Index(lowerData, subseq), Value: subseq})
					}
				}
			}
		}

		reversed := make([]byte, len(row))
		for i, b := range row {
			reversed[len(row)-1-i] = b
		}
		for length := 3; length <= len(reversed); length++ {
			for i := 0; i <= len(reversed)-length; i++ {
				subseq := reversed[i : i+length]
				if bytes.Contains(lowerData, subseq) {
					val := string(subseq)
					if !seen[val] {
						seen[val] = true
						patterns = append(patterns, Pattern{Type: "keyboard", Length: length, Start: bytes.Index(lowerData, subseq), Value: subseq})
					}
				}
			}
		}
	}
	return patterns
}

func calculateCrackTimeWithSeconds(guesses float64) (string, float64) {
	guessesPerSecond := 1e10
	seconds := guesses / guessesPerSecond

	var timeStr string
	switch {
	case seconds < 1:
		timeStr = "instantly"
	case seconds < 60:
		timeStr = "seconds"
	case seconds < 3600:
		timeStr = "minutes"
	case seconds < 86400:
		timeStr = "hours"
	case seconds < 86400*365:
		timeStr = "days"
	default:
		timeStr = "years"
	}
	return timeStr, seconds
}
