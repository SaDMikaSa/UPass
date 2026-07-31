package tyuiop

import (
	"context"
	"net/http"
	"time"
)

type PasswordStats struct {
	Length     int
	Digits     int
	Lowers     int
	Uppers     int
	Specials   int
	UniqueChar int
	MaxRepeat  int
}

type Pattern struct {
	Type   string // "repeat", "keyboard", "common"
	Length int
	Start  int
	Value  []byte
}

type PasswordStrength struct {
	Score        int
	Guesses      float64
	CrackTime    string
	CrackSeconds float64
	Patterns     []Pattern
	Stats        *PasswordStats
	Feedback     Feedback
}

type Feedback struct {
	Warning     string
	Suggestions []string
}

type PwnedCheckResult struct {
	IsPwned     bool   // is password compromised
	BreachCount int    // number of breaches
	Service     string // service name
	Error       error  // error during check
}

type PwnedChecker interface {
	CheckPassword(ctx context.Context, password []byte) (*PwnedCheckResult, error)
}

type HIBPClient struct {
	httpClient *http.Client
	userAgent  string
	baseURL    string
}

func NewHIBPClient(userAgent string) *HIBPClient {
	return &HIBPClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgent: userAgent,
		baseURL:   "https://api.pwnedpasswords.com",
	}
}

type PwnedAggregator struct {
	checkers []PwnedChecker
}

func NewPwnedAggregator(checkers ...PwnedChecker) *PwnedAggregator {
	return &PwnedAggregator{
		checkers: checkers,
	}
}

var commonPasswordsBytes = [][]byte{
	[]byte("password"), []byte("123456"), []byte("12345678"), []byte("1234"),
	[]byte("qwerty"), []byte("12345"), []byte("dragon"), []byte("baseball"),
	[]byte("football"), []byte("letmein"), []byte("monkey"), []byte("abc123"),
	[]byte("mustang"), []byte("michael"), []byte("shadow"), []byte("master"),
	[]byte("111111"), []byte("password1"), []byte("password123"),
	[]byte("admin"), []byte("welcome"), []byte("login"), []byte("passw0rd"),
	[]byte("sunshine"), []byte("princess"), []byte("qwerty123"), []byte("123456789"),
	[]byte("superman"), []byte("iloveyou"), []byte("fuckyou"), []byte("666666"),
	[]byte("123123"), []byte("654321"), []byte("qwertyuiop"), []byte("password!"),
	[]byte("1234567890"),
}
