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

var keyboardGraph = map[byte][]byte{
	'`':  {'1', '~'},
	'1':  {'`', '2', 'q', '!'},
	'2':  {'1', '3', 'q', 'w', '@'},
	'3':  {'2', '4', 'w', 'e', '#'},
	'4':  {'3', '5', 'e', 'r', '$'},
	'5':  {'4', '6', 'r', 't', '%'},
	'6':  {'5', '7', 't', 'y', '^'},
	'7':  {'6', '8', 'y', 'u', '&'},
	'8':  {'7', '9', 'u', 'i', '*'},
	'9':  {'8', '0', 'i', 'o', '('},
	'0':  {'9', '-', 'o', 'p', ')'},
	'-':  {'0', '=', 'p', '['},
	'=':  {'-', ']'},
	'~':  {'1'},
	'!':  {'1', '2', 'q'},
	'@':  {'2', '3', 'q', 'w'},
	'#':  {'3', '4', 'w', 'e'},
	'$':  {'4', '5', 'e', 'r'},
	'%':  {'5', '6', 'r', 't'},
	'^':  {'6', '7', 't', 'y'},
	'&':  {'7', '8', 'y', 'u'},
	'*':  {'8', '9', 'u', 'i'},
	'(':  {'9', '0', 'i', 'o'},
	')':  {'0', '-', 'o', 'p'},
	'q':  {'w', 'a', 's', '1', '2', '`'},
	'w':  {'q', 'e', 'a', 's', 'd', '2', '3'},
	'e':  {'w', 'r', 's', 'd', 'f', '3', '4'},
	'r':  {'e', 't', 'd', 'f', 'g', '4', '5'},
	't':  {'r', 'y', 'f', 'g', 'h', '5', '6'},
	'y':  {'t', 'u', 'g', 'h', 'j', '6', '7'},
	'u':  {'y', 'i', 'h', 'j', 'k', '7', '8'},
	'i':  {'u', 'o', 'j', 'k', 'l', '8', '9'},
	'o':  {'i', 'p', 'k', 'l', '9', '0'},
	'p':  {'o', '[', 'l', ';', '0', '-'},
	'a':  {'q', 'w', 's', 'z', 'x'},
	's':  {'a', 'w', 'e', 'd', 'z', 'x', 'c'},
	'd':  {'s', 'e', 'r', 'f', 'x', 'c', 'v'},
	'f':  {'d', 'r', 't', 'g', 'c', 'v', 'b'},
	'g':  {'f', 't', 'y', 'h', 'v', 'b', 'n'},
	'h':  {'g', 'y', 'u', 'j', 'b', 'n', 'm'},
	'j':  {'h', 'u', 'i', 'k', 'n', 'm'},
	'k':  {'j', 'i', 'o', 'l', 'm'},
	'l':  {'k', 'o', 'p', ';', ':'},
	';':  {'l', 'p', '[', ':'},
	':':  {';', 'l', 'p'},
	'[':  {'p', ';', ']'},
	']':  {'[', '='},
	'\\': {']'},
	'z':  {'a', 's', 'x'},
	'x':  {'z', 's', 'd', 'c'},
	'c':  {'x', 'd', 'f', 'v'},
	'v':  {'c', 'f', 'g', 'b'},
	'b':  {'v', 'g', 'h', 'n'},
	'n':  {'b', 'h', 'j', 'm'},
	'm':  {'n', 'j', 'k', ',', '<'},
	',':  {'m', '.', '<'},
	'<':  {','},
	'.':  {',', '/', '>'},
	'>':  {'.'},
	'/':  {'.'},
}

var commonPasswordsBytes = [][]byte{
	// === Базовые (уже есть) ===
	[]byte("password"), []byte("dragon"), []byte("baseball"), []byte("password!"),
	[]byte("football"), []byte("letmein"), []byte("monkey"), []byte("abc123"),
	[]byte("123456"), []byte("12345678"), []byte("123456789"), []byte("qwerty"),
	[]byte("mustang"), []byte("michael"), []byte("shadow"), []byte("master"),
	[]byte("password1"), []byte("password123"),
	[]byte("admin"), []byte("welcome"), []byte("login"), []byte("passw0rd"),
	[]byte("sunshine"), []byte("princess"), []byte("qwerty123"),
	[]byte("superman"), []byte("iloveyou"), []byte("fuckyou"),
	[]byte("jennifer"), []byte("thomas"), []byte("charlie"), []byte("robert"),
	[]byte("jessica"), []byte("daniel"), []byte("ashley"), []byte("matthew"),
	[]byte("samantha"), []byte("andrew"), []byte("michelle"), []byte("william"),
	[]byte("kimberly"), []byte("richard"), []byte("brittany"), []byte("christopher"),
	[]byte("p@ssw0rd"), []byte("p455w0rd"), []byte("pa$$word"), []byte("p@$$w0rd"),
	[]byte("5up3rm4n"), []byte("dr4g0n"), []byte("m0nk3y"), []byte("b4s3b4ll"),
	[]byte("f00tb4ll"), []byte("l3tm31n"), []byte("m1ch43l"), []byte("sh4d0w"),
	[]byte("welcome1"), []byte("welcome123"), []byte("admin123"), []byte("admin2024"),
	[]byte("letmein123"), []byte("letmein1"), []byte("login123"), []byte("login1"),
	[]byte("love123"), []byte("loveyou"), []byte("iloveyou1"), []byte("iloveyou123"),
	[]byte("sunshine1"), []byte("princess1"), []byte("superman1"), []byte("dragon1"),
	[]byte("1990"), []byte("1991"), []byte("1992"), []byte("1993"),
	[]byte("1994"), []byte("1995"), []byte("1996"), []byte("1997"),
	[]byte("1998"), []byte("1999"), []byte("2000"), []byte("2001"),
	[]byte("2002"), []byte("2003"), []byte("2004"), []byte("2005"),
	[]byte("2006"), []byte("2007"), []byte("2008"), []byte("2009"),
	[]byte("2010"), []byte("2011"), []byte("2012"), []byte("2013"),
	[]byte("2014"), []byte("2015"), []byte("2016"), []byte("2017"),
	[]byte("2018"), []byte("2019"), []byte("2020"), []byte("2021"),
	[]byte("2022"), []byte("2023"), []byte("2024"),
	[]byte("whatever"), []byte("whatever1"), []byte("whatever123"),
	[]byte("trustno1"), []byte("trustnoone"), []byte("hello123"),
	[]byte("hello1"), []byte("nicole"), []byte("jordan"),
	[]byte("blink182"), []byte("1984"), []byte("charlie1"),
	[]byte("abc123"), []byte("abc123!"), []byte("qwe123"), []byte("zxc123"),
	[]byte("qwerty1"), []byte("qwerty123!"), []byte("asd123"), []byte("asdf123"),
	[]byte("123abc"), []byte("123qwe"), []byte("123asd"), []byte("123zxc"),
	[]byte("1q2w3e"), []byte("1q2w3e4r"), []byte("1qaz2wsx"), []byte("qazwsx"),
	[]byte("zaq12wsx"), []byte("qweasd"), []byte("qwertyasdf"),
}
