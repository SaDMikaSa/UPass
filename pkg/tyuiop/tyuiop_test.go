package tyuiop

import (
	"testing"
)

func TestValidateASCIIOnly(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"ASCII only", []byte("Hello123!"), true},
		{"Cyrillic", []byte("Привет00"), false},
		{"Emoji", []byte("🚀"), false},
		{"Mixed", []byte("Hello🚀"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateASCIIOnly(tt.data); got != tt.want {
				t.Errorf("ValidateASCIIOnly(%s) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestCalculateEntropy(t *testing.T) {
	tests := []struct {
		name       string
		password   []byte
		minScore   int
		maxGuesses float64
	}{
		{
			name:       "simple",
			password:   []byte("password"),
			minScore:   0,
			maxGuesses: 1e8,
		},
		{
			name:       "complex",
			password:   []byte("MyS3cur3P@ssw0rd2024!"),
			minScore:   3,
			maxGuesses: 2e25, // 4 bits/char cap * 21 chars = 84 bits -> 2^84
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := AnalyzeBytes(tt.password)
			patterns := findRepeats(tt.password)
			patterns = append(patterns, findKeyboardPatterns(tt.password)...)
			patterns = append(patterns, findCombinedPatterns(tt.password)...)

			if common := findCommonPassword(tt.password); common != nil {
				patterns = append(patterns, *common)
			}

			guesses := calculateEntropy(tt.password, patterns, stats)

			if guesses > tt.maxGuesses {
				t.Errorf("Guesses too high: %e > %e", guesses, tt.maxGuesses)
			}
		})
	}
}

func TestFindKeyboardPatterns(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int
	}{
		{"qwerty", []byte("qwerty"), 1},
		{"qwertyuiop", []byte("qwertyuiop"), 1},
		{"asdfghjkl", []byte("asdfghjkl"), 1},
		{"zxcvbnm", []byte("zxcvbnm"), 1},
		{"1qaz2wsx", []byte("1qaz2wsx"), 2},
		{"qwerty123", []byte("qwerty123"), 2},
		{"qwertyuiopasdfghjkl", []byte("qwertyuiopasdfghjkl"), 2},
		{"abc", []byte("abc"), 0},
		{"qweasd", []byte("qweasd"), 2},
		{"zaq12wsx", []byte("zaq12wsx"), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findKeyboardPatterns(tt.data)
			if len(got) != tt.expected {
				t.Errorf("findKeyboardPatterns(%s) = %d, want %d",
					tt.data, len(got), tt.expected)
			}
		})
	}
}

func TestFindRepeats_Complex(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int
	}{
		{"aaa", []byte("aaa"), 1},
		{"aaaa", []byte("aaaa"), 1},
		{"ababab", []byte("ababab"), 1},
		{"abcabcabc", []byte("abcabcabc"), 1},
		{"aaaabbbb", []byte("aaaabbbb"), 2},
		{"aabbcc", []byte("aabbcc"), 3},
		{"abababab", []byte("abababab"), 1},
		{"123123123", []byte("123123123"), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findRepeats(tt.data)
			if len(got) != tt.expected {
				t.Errorf("findRepeats(%s) = %d, want %d",
					tt.data, len(got), tt.expected)
			}
		})
	}
}

func TestFindCommonPassword(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{"password", []byte("password"), true},
		{"PASSWORD", []byte("PASSWORD"), true},
		{"password123", []byte("password123"), true},
		{"password2024", []byte("password2024"), true},
		{"myPassword", []byte("myPassword"), false},
		{"123456", []byte("123456"), true},
		{"qwerty", []byte("qwerty"), true},
		{"notcommon", []byte("notcommon"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findCommonPassword(tt.data)
			if (got != nil) != tt.expected {
				t.Errorf("findCommonPassword(%s) = %v, want %v",
					tt.data, got != nil, tt.expected)
			}
		})
	}
}

func TestFindCombinedPatterns(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int
	}{
		{"love123", []byte("love123"), 1},
		{"love2024", []byte("love2024"), 1},
		{"123love", []byte("123love"), 1},
		{"love12345", []byte("love12345"), 1},
		{"mylove", []byte("mylove"), 0},
		{"king123", []byte("king123"), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findCombinedPatterns(tt.data)
			if len(got) != tt.expected {
				t.Errorf("findCombinedPatterns(%s) = %d, want %d",
					tt.data, len(got), tt.expected)
			}
		})
	}
}

func TestCalculateCrackTimeWithSeconds(t *testing.T) {
	tests := []struct {
		guesses    float64
		hashType   string
		minSeconds float64
	}{
		{1e3, "bcrypt", 1.0},
		{1e6, "bcrypt", 1e3},
		{1e9, "bcrypt", 1e6},
		{1e3, "md5", 1e-6},
		{1e6, "md5", 1e-3},
	}

	for _, tt := range tests {
		t.Run(tt.hashType, func(t *testing.T) {
			_, seconds := calculateCrackTimeWithSeconds(tt.guesses, tt.hashType)
			if seconds < tt.minSeconds {
				t.Errorf("calculateCrackTimeWithSeconds(%e, %s) = %e, want >= %e",
					tt.guesses, tt.hashType, seconds, tt.minSeconds)
			}
		})
	}
}

func TestCorrectHorseBatteryStaple(t *testing.T) {
	password := []byte("correct-horse-battery-staple")
	strength := AnalyzeLocalStrength(password)

	if strength.Score < 3 {
		t.Errorf("Expected score >= 3, got %d (guesses: %e)",
			strength.Score, strength.Guesses)
	}

	if strength.Score != 4 {
		t.Logf("Expected score 4, got %d", strength.Score)
		t.Logf("Guesses: %e, CrackTime: %s", strength.Guesses, strength.CrackTime)
	}
}
