package tyuiop

import (
	"context"
	"fmt"
	"math"
	"time"
)

func EstimateStrength(data []byte, pwnedChecker PwnedChecker) *PasswordStrength {
	stats := AnalyzeBytes(data)
	repeats := findRepeats(data)
	keyboard := findKeyboardPatterns(data)

	patterns := append(repeats, keyboard...)

	possibleChars := calculatePossibleChars(stats)
	length := stats.Length

	baseGuesses := math.Pow(float64(possibleChars), float64(length))
	finalGuesses := applyPenalties(baseGuesses, patterns, stats)

	score := calculateScore(finalGuesses)
	crackTime := calculateCrackTime(finalGuesses)
	feedback := generateFeedback(stats, patterns, score)

	var pwnedResult *PwnedCheckResult
	if pwnedChecker != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pwnedResult, _ = pwnedChecker.CheckPassword(ctx, data)
	}

	if pwnedResult != nil && pwnedResult.IsPwned {
		score = 0
		finalGuesses = 1
		feedback := generateFeedback(stats, patterns, score)
		feedback.Warning = fmt.Sprintf("Password found in data breaches (%d times)", pwnedResult.BreachCount)
		feedback.Suggestions = append(feedback.Suggestions, "Use a unique password that you haven't used on other sites")

		return &PasswordStrength{
			Score:     0,
			Guesses:   1,
			CrackTime: "instantly",
			Patterns:  patterns,
			Stats:     stats,
			Feedback:  feedback,
		}
	}

	return &PasswordStrength{
		Score:     score,
		Guesses:   finalGuesses,
		CrackTime: crackTime,
		Patterns:  patterns,
		Stats:     stats,
		Feedback:  feedback,
	}
}

func generateFeedback(stats *PasswordStats, patterns []Pattern, score int) Feedback {
	var fb Feedback
	hasKeyboard, hasRepeat := false, false
	for _, p := range patterns {
		if p.Type == "keyboard" {
			hasKeyboard = true
		}
		if p.Type == "repeat" {
			hasRepeat = true
		}
	}

	if hasRepeat {
		fb.Warning = "Contains repeating patterns"
	} else if hasKeyboard {
		fb.Warning = "Contains keyboard sequences"
	} else if stats.Digits == stats.Length {
		fb.Warning = "Password consists only of digits"
	} else if stats.Length < 8 {
		fb.Warning = "Password is too short"
	}

	if stats.Uppers == 0 {
		fb.Suggestions = append(fb.Suggestions, "Use uppercase letters")
	}
	if stats.Lowers == 0 {
		fb.Suggestions = append(fb.Suggestions, "Use lowercase letters")
	}
	if stats.Digits == 0 {
		fb.Suggestions = append(fb.Suggestions, "Add numbers")
	}
	if stats.Specials == 0 {
		fb.Suggestions = append(fb.Suggestions, "Add special characters")
	}
	if stats.Length < 12 {
		fb.Suggestions = append(fb.Suggestions, "Increase length to 12+ characters")
	}

	if score >= 3 && len(fb.Suggestions) == 0 {
		fb.Suggestions = append(fb.Suggestions, "Excellent password!")
	}
	return fb
}
