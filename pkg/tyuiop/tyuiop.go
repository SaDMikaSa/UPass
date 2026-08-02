package tyuiop

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
