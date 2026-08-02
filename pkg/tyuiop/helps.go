package tyuiop

import (
	"bytes"
	"math"
	"sort"
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

func calculateEntropy(data []byte, patterns []Pattern, stats *PasswordStats) float64 {
	if stats.Length == 0 {
		return 1.0
	}

	baseCharSet := calculatePossibleChars(stats)

	if len(patterns) == 0 {
		return math.Pow(2, math.Log2(float64(baseCharSet))*float64(stats.Length))
	}

	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Length > patterns[j].Length
	})

	covered := make([]bool, stats.Length)
	patternEntropy := 0.0

	for _, p := range patterns {
		overlapping := false
		for i := p.Start; i < p.Start+p.Length && i < len(covered); i++ {
			if covered[i] {
				overlapping = true
				break
			}
		}
		if overlapping {
			continue
		}

		for i := p.Start; i < p.Start+p.Length && i < len(covered); i++ {
			covered[i] = true
		}

		switch p.Type {
		case "common":
			patternEntropy += 6.0 // Очень низкая энтропия для общих паролей
		case "repeat":
			patternEntropy += 1.0 * float64(p.Length) // Минимальная энтропия
		case "keyboard":
			patternEntropy += 0.5 * float64(p.Length) // Почти нулевая энтропия
		case "common_mutated":
			patternEntropy += 8.0
		case "compound":
			words := bytes.Split(p.Value, []byte{'-'})
			wordEntropy := float64(len(words)) * 6.0 // Уменьшено с 7
			patternEntropy += wordEntropy
		default:
			patternEntropy += math.Log2(float64(baseCharSet)) * float64(p.Length)
		}
	}

	uncovered := 0
	for _, c := range covered {
		if !c {
			uncovered++
		}
	}

	if uncovered > 0 {
		patternEntropy += math.Log2(float64(baseCharSet)) * float64(uncovered)
	}

	// Очень сильный штраф за однообразие
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
		patternEntropy *= 0.1 // Почти уничтожаем энтропию
	} else if charSets == 2 {
		patternEntropy *= 0.3
	} else if charSets == 3 {
		patternEntropy *= 0.6
	}

	// Жесткое ограничение: максимум 4 бита на символ
	maxEntropy := float64(stats.Length) * 4.0
	if patternEntropy > maxEntropy {
		patternEntropy = maxEntropy
	}

	if patternEntropy < 1.0 && stats.Length > 0 {
		patternEntropy = 1.0
	}

	return math.Pow(2, patternEntropy)
}

func calculateScore(guesses float64) int {
	switch {
	case guesses < 1e3:
		return 0
	case guesses < 1e5:
		return 1
	case guesses < 1e7:
		return 2
	case guesses < 1e9:
		return 3
	default:
		return 4
	}
}

func findCommonPassword(data []byte) *Pattern {
	lowerData := bytes.ToLower(data)
	dataLen := len(data)

	// Сначала проверяем точные совпадения
	for _, common := range commonPasswordsBytes {
		if bytes.Equal(lowerData, common) {
			return &Pattern{
				Type:   "common",
				Length: len(common),
				Start:  0,
				Value:  common,
			}
		}
	}

	// Для коротких паролей (до 8 символов) - проверяем вхождение
	if dataLen <= 8 {
		for _, common := range commonPasswordsBytes {
			if bytes.Contains(lowerData, common) {
				// Проверяем, что это не часть более длинного слова
				idx := bytes.Index(lowerData, common)
				// Если длина совпадает или пароль состоит только из этого слова
				if dataLen == len(common) {
					return &Pattern{
						Type:   "common",
						Length: len(common),
						Start:  idx,
						Value:  common,
					}
				}
				// Для паролей типа 123456, qwerty - они должны быть точным совпадением
				if dataLen <= len(common)+2 && dataLen >= len(common) {
					return &Pattern{
						Type:   "common",
						Length: len(common),
						Start:  idx,
						Value:  common,
					}
				}
			}
		}
		return nil
	}

	// Для длинных паролей (> 8 символов)
	for _, common := range commonPasswordsBytes {
		// Проверяем префикс + цифры/спецсимволы
		if dataLen > len(common) && bytes.HasPrefix(lowerData, common) {
			suffix := lowerData[len(common):]
			if isOnlyDigitsOrSpecial(suffix) && len(suffix) <= 4 {
				return &Pattern{
					Type:   "common",
					Length: len(common) + len(suffix),
					Start:  0,
					Value:  data[:len(common)+len(suffix)],
				}
			}
		}

		// Проверяем отдельное слово (окруженное спецсимволами или границами)
		idx := bytes.Index(lowerData, common)
		if idx != -1 {
			startOk := idx == 0 || specialChars[lowerData[idx-1]]
			end := idx + len(common)
			endOk := end == dataLen || specialChars[lowerData[end]]
			if startOk && endOk {
				// Проверяем, что это не myPassword (содержит password как часть)
				if dataLen > len(common)+2 {
					// Проверяем, что это действительно отдельное слово
					prefix := lowerData[:idx]
					suffix := lowerData[end:]
					if (len(prefix) == 0 || specialChars[prefix[len(prefix)-1]]) &&
						(len(suffix) == 0 || specialChars[suffix[0]]) {
						return &Pattern{
							Type:   "common",
							Length: len(common),
							Start:  idx,
							Value:  common,
						}
					}
				} else {
					return &Pattern{
						Type:   "common",
						Length: len(common),
						Start:  idx,
						Value:  common,
					}
				}
			}
		}
	}
	return nil
}

func findRepeats(data []byte) []Pattern {
	var patterns []Pattern
	if len(data) < 3 {
		return patterns
	}

	seen := make(map[string]bool)

	for length := 1; length <= len(data)/2; length++ {
		for start := 0; start <= len(data)-length*2; start++ {
			block := data[start : start+length]

			repeatCount := 1
			pos := start + length
			for pos+length <= len(data) && bytes.Equal(data[pos:pos+length], block) {
				repeatCount++
				pos += length
			}

			if repeatCount >= 2 {
				totalLen := repeatCount * length
				key := string(block)
				if !seen[key] {
					seen[key] = true
					patterns = append(patterns, Pattern{
						Type:   "repeat",
						Length: totalLen,
						Start:  start,
						Value:  block,
					})
				}
			}
		}
	}

	return mergeOverlappingPatterns(patterns)
}

func mergeOverlappingPatterns(patterns []Pattern) []Pattern {
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Length > patterns[j].Length
	})

	var result []Pattern
	used := make([]bool, len(patterns))

	for i, p := range patterns {
		if used[i] {
			continue
		}
		result = append(result, p)
		for j := i + 1; j < len(patterns); j++ {
			if patterns[j].Start < p.Start+p.Length &&
				patterns[j].Start+patterns[j].Length > p.Start {
				used[j] = true
			}
		}
	}
	return result
}

func findCombinedPatterns(data []byte) []Pattern {
	var patterns []Pattern
	lowerData := bytes.ToLower(data)

	simpleWords := [][]byte{
		[]byte("love"), []byte("sex"), []byte("god"), []byte("angel"),
		[]byte("devil"), []byte("star"), []byte("sun"), []byte("moon"),
		[]byte("sky"), []byte("fire"), []byte("ice"), []byte("dark"),
		[]byte("light"), []byte("life"), []byte("death"), []byte("king"),
		[]byte("queen"), []byte("prince"), []byte("rose"), []byte("tiger"),
		[]byte("eagle"), []byte("wolf"), []byte("bear"), []byte("lion"),
		[]byte("hero"), []byte("zero"), []byte("one"), []byte("two"),
		[]byte("three"), []byte("four"), []byte("five"), []byte("six"),
		[]byte("seven"), []byte("eight"), []byte("nine"), []byte("ten"),
	}

	for _, word := range simpleWords {
		idx := bytes.Index(lowerData, word)
		if idx == -1 {
			continue
		}

		if idx > 0 {
			prefix := lowerData[:idx]
			if isOnlyDigitsOrSpecial(prefix) && len(prefix) <= 5 {
				patterns = append(patterns, Pattern{
					Type:   "common_mutated",
					Length: len(prefix) + len(word),
					Start:  0,
					Value:  data[:len(prefix)+len(word)],
				})
			}
		}

		if idx+len(word) < len(data) {
			suffix := lowerData[idx+len(word):]
			if isOnlyDigitsOrSpecial(suffix) && len(suffix) <= 5 {
				patterns = append(patterns, Pattern{
					Type:   "common_mutated",
					Length: len(word) + len(suffix),
					Start:  idx,
					Value:  data[idx : idx+len(word)+len(suffix)],
				})
			}
		}
	}

	return patterns
}

func isOnlyDigitsOrSpecial(data []byte) bool {
	for _, b := range data {
		if !(b >= '0' && b <= '9' || specialChars[b]) {
			return false
		}
	}
	return len(data) > 0
}

func findKeyboardPatterns(data []byte) []Pattern {
	var patterns []Pattern
	if len(data) < 3 {
		return patterns
	}

	lowerData := bytes.ToLower(data)
	seen := make(map[string]bool)

	type path struct {
		start  int
		length int
		value  []byte
	}
	var allPaths []path

	for start := 0; start < len(lowerData)-2; start++ {
		if !isValidKeyboardStart(lowerData[start]) {
			continue
		}

		maxLen := 0
		for length := 3; start+length <= len(lowerData); length++ {
			if isValidKeyboardPath(lowerData[start : start+length]) {
				maxLen = length
			} else {
				break
			}
		}

		if maxLen >= 3 {
			allPaths = append(allPaths, path{
				start:  start,
				length: maxLen,
				value:  lowerData[start : start+maxLen],
			})
		}
	}

	// Сортируем по длине (самые длинные первыми)
	sort.Slice(allPaths, func(i, j int) bool {
		return allPaths[i].length > allPaths[j].length
	})

	// Жадный выбор непересекающихся паттернов
	used := make([]bool, len(lowerData))
	for _, p := range allPaths {
		// Проверяем перекрытие
		overlap := false
		for i := p.start; i < p.start+p.length && i < len(used); i++ {
			if used[i] {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}

		// Проверяем, не является ли этот паттерн частью более длинного
		isSubpath := false
		for _, existing := range patterns {
			if p.start >= existing.Start && p.start+p.length <= existing.Start+existing.Length {
				isSubpath = true
				break
			}
		}
		if isSubpath {
			continue
		}

		// Помечаем как использованные
		for i := p.start; i < p.start+p.length && i < len(used); i++ {
			used[i] = true
		}

		key := string(p.value)
		if !seen[key] {
			seen[key] = true
			patterns = append(patterns, Pattern{
				Type:   "keyboard",
				Length: p.length,
				Start:  p.start,
				Value:  p.value,
			})
		}
	}

	// Если ничего не найдено, проверяем простые паттерны
	if len(patterns) == 0 && len(data) >= 3 {
		simplePatterns := [][]byte{
			[]byte("qwerty"), []byte("asdfgh"), []byte("zxcvbn"),
			[]byte("qwe"), []byte("asd"), []byte("zxc"),
		}
		for _, sp := range simplePatterns {
			if bytes.Contains(lowerData, sp) {
				start := bytes.Index(lowerData, sp)
				key := string(sp)
				if !seen[key] {
					seen[key] = true
					patterns = append(patterns, Pattern{
						Type:   "keyboard",
						Length: len(sp),
						Start:  start,
						Value:  sp,
					})
				}
			}
		}
	}

	return patterns
}

func isValidKeyboardStart(b byte) bool {
	_, ok := keyboardGraph[b]
	return ok
}

func isValidKeyboardPath(seq []byte) bool {
	if len(seq) < 2 {
		return false
	}

	for i := 1; i < len(seq); i++ {
		neighbors, ok := keyboardGraph[seq[i-1]]
		if !ok {
			return false
		}

		found := false
		for _, n := range neighbors {
			if n == seq[i] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func calculateCrackTimeWithSeconds(guesses float64, hashType string) (string, float64) {
	var guessesPerSecond float64

	switch hashType {
	case "bcrypt":
		guessesPerSecond = 1e3
	case "scrypt":
		guessesPerSecond = 1e2
	case "argon2":
		guessesPerSecond = 1e1
	default:
		guessesPerSecond = 1e9
	}

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
	case seconds < 86400*7:
		timeStr = "days"
	case seconds < 86400*30:
		timeStr = "weeks"
	default:
		timeStr = "months"
	}

	return timeStr, seconds
}

func ValidateASCIIOnly(data []byte) bool {
	for _, b := range data {
		if b > 127 {
			return false
		}
	}
	return true
}
