package health

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SaDMikaSa/UPass/internal/domain"
	"github.com/SaDMikaSa/UPass/pkg/tyuiop"
	"golang.org/x/sync/errgroup"
)

type WeakPasswordResult struct {
	Service      string
	Score        int
	CrackTime    string
	CrackSeconds float64
}

type DuplicateGroup struct {
	Services []string
}

type BreachedResult struct {
	Service string
	Count   int
}

type ReusedLogin struct {
	Login    string
	Services []string
}

var hibpClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	},
}

func CheckWeakPasswords(records map[string]domain.Record, minScore int) []WeakPasswordResult {
	var results []WeakPasswordResult

	for _, rec := range records {
		strength := tyuiop.AnalyzeLocalStrength(rec.Password)
		if strength.Score < minScore {
			results = append(results, WeakPasswordResult{
				Service:      string(rec.Service),
				Score:        strength.Score,
				CrackTime:    strength.CrackTime,
				CrackSeconds: strength.CrackSeconds,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score < results[j].Score
	})

	return results
}

// CheckDuplicatePasswords identifies groups of services that share the same
// password by hashing passwords with SHA‑256 and grouping identical hashes.
func CheckDuplicatePasswords(records map[string]domain.Record) []DuplicateGroup {
	hashToServices := make(map[string][]string)

	for _, rec := range records {
		hash := sha256.Sum256(rec.Password)
		hashKey := fmt.Sprintf("%x", hash)
		hashToServices[hashKey] = append(hashToServices[hashKey], string(rec.Service))
	}

	var duplicates []DuplicateGroup
	for _, services := range hashToServices {
		if len(services) > 1 {
			sort.Strings(services)
			duplicates = append(duplicates, DuplicateGroup{Services: services})
		}
	}

	sort.Slice(duplicates, func(i, j int) bool {
		return len(duplicates[i].Services) > len(duplicates[j].Services)
	})

	return duplicates
}

func fetchHIBPRange(prefix string) ([]byte, error) {
	url := "https://api.pwnedpasswords.com/range/" + prefix
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Add-Padding", "true")
	req.Header.Set("User-Agent", "UPass-CLI")

	resp, err := hibpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hibp request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("HIBP rate limit exceeded (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hibp status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return body, nil
}

// searchSuffixInResponse parses the HIBP /range response body for a line
// starting with suffix and returns the breach count when found.
func searchSuffixInResponse(body []byte, suffix string) int {
	lines := bytes.Split(body, []byte("\n"))
	suffixBytes := []byte(suffix)
	for _, line := range lines {
		if bytes.HasPrefix(line, suffixBytes) {
			parts := bytes.SplitN(line, []byte(":"), 2)
			if len(parts) == 2 {
				count, err := strconv.Atoi(strings.TrimSpace(string(parts[1])))
				if err != nil {
					return 0
				}
				return count
			}
		}
	}
	return 0
}

// CheckBreached checks whether the provided password appears in the
// Have I Been Pwned database by sending only the first 5 hex characters of
// the SHA-1 hash (k-anonymity) and searching the returned suffix list.
func CheckBreached(password []byte) (int, error) {
	hash := sha1.Sum(password)
	hashStr := strings.ToUpper(fmt.Sprintf("%X", hash))
	prefix := hashStr[:5]
	suffix := hashStr[5:]

	body, err := fetchHIBPRange(prefix)
	if err != nil {
		return 0, err
	}

	return searchSuffixInResponse(body, suffix), nil
}

func CheckAllBreached(records map[string]domain.Record) ([]BreachedResult, error) {
	type recordInfo struct {
		service string
		suffix  string
	}

	prefixGroups := make(map[string][]recordInfo)
	for _, rec := range records {
		hash := sha1.Sum(rec.Password)
		hashStr := fmt.Sprintf("%X", hash)
		prefix := hashStr[:5]
		suffix := hashStr[5:]
		prefixGroups[prefix] = append(prefixGroups[prefix], recordInfo{
			service: string(rec.Service),
			suffix:  suffix,
		})
	}

	if len(prefixGroups) == 0 {
		return nil, nil
	}

	var results []BreachedResult
	var mu sync.Mutex
	var apiFailedCount int32

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(10)

	for prefix, infos := range prefixGroups {
		p := prefix
		i := infos

		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			body, err := fetchHIBPRange(p)
			if err != nil {
				atomic.AddInt32(&apiFailedCount, 1)
				return nil
			}

			for _, info := range i {
				count := searchSuffixInResponse(body, info.suffix)
				if count > 0 {
					mu.Lock()
					results = append(results, BreachedResult{
						Service: info.service,
						Count:   count,
					})
					mu.Unlock()
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if int(apiFailedCount) == len(prefixGroups) {
		return nil, fmt.Errorf("HIBP API is currently unavailable (all %d requests failed)", len(prefixGroups))
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Count > results[j].Count
	})

	return results, nil
}

// CheckReusedLogins finds login identifiers that are used across multiple
// services and returns lists of services per duplicated login.
func CheckReusedLogins(records map[string]domain.Record) []ReusedLogin {
	loginMap := make(map[string][]string)

	for _, rec := range records {
		login := string(rec.Login)
		loginMap[login] = append(loginMap[login], string(rec.Service))
	}

	var reused []ReusedLogin
	for login, services := range loginMap {
		if len(services) > 1 {
			sort.Strings(services)
			reused = append(reused, ReusedLogin{
				Login:    login,
				Services: services,
			})
		}
	}

	sort.Slice(reused, func(i, j int) bool {
		return len(reused[i].Services) > len(reused[j].Services)
	})

	return reused
}
