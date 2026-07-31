package tyuiop

import "testing"

func TestAnalyzeLocalStrength(t *testing.T) {
	tests := []struct {
		name     string
		password []byte
		minScore int
	}{
		{"weak password", []byte("password"), 0},
		{"medium password", []byte("P@ssw0rd"), 2},
		{"strong password", []byte("MyS3cur3P@ssw0rd2024!"), 4},
		{"empty password", []byte(""), 0},
		{"digits only", []byte("123456"), 0},
		{"keyboard pattern", []byte("qwerty"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzeLocalStrength(tt.password)
			if result.Score < tt.minScore {
				t.Errorf("Expected score >= %d, got %d for password %s",
					tt.minScore, result.Score, tt.password)
			}
		})
	}
}

func TestFindRepeats(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int // количество паттернов
	}{
		{"no repeats", []byte("abc"), 0},
		{"simple repeat", []byte("aaa"), 1},
		{"mixed repeats", []byte("aaabbbccc"), 3},
		{"empty", []byte(""), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := findRepeats(tt.data)
			if len(patterns) != tt.expected {
				t.Errorf("Expected %d patterns, got %d", tt.expected, len(patterns))
			}
		})
	}
}
