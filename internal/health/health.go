package health

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SaDMikaSa/UPass/internal/domain"
	"github.com/nbutton23/zxcvbn-go"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    10 * time.Second,
		DisableCompression: false,
	},
}

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

// CheckWeakPasswords evaluates password strength for each record using
// zxcvbn and returns those with score < minScore.
func CheckWeakPasswords(records map[string]domain.Record, minScore int) []WeakPasswordResult {
	var results []WeakPasswordResult

	for _, rec := range records {
		passwordStr := string(rec.Password)
		strength := zxcvbn.PasswordStrength(passwordStr, nil)
		if strength.Score < minScore {
			results = append(results, WeakPasswordResult{
				Service:      string(rec.Service),
				Score:        strength.Score,
				CrackTime:    strength.CrackTimeDisplay,
				CrackSeconds: strength.CrackTime,
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

// fetchHIBPRange queries the Have I Been Pwned password range API for the
// given 5-character hex prefix and returns the raw body. The caller is
// responsible for parsing the response. The request sets Add-Padding:true for
// additional privacy and uses a short timeout.
func fetchHIBPRange(prefix string) ([]byte, error) {
	url := "https://api.pwnedpasswords.com/range/" + prefix

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Add-Padding", "true")
	req.Header.Set("User-Agent", "UPass-CLI")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hibp request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("hibp status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
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

// CheckAllBreached performs HIBP checks for all records in the vault.
// It groups passwords by the SHA‑1 prefix to minimize the number of network
// requests (k-anonymity) and returns a list of breached services with counts.
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

	var results []BreachedResult
	apiFailedCount := 0

	for prefix, infos := range prefixGroups {
		body, err := fetchHIBPRange(prefix)
		if err != nil {
			apiFailedCount++
			continue
		}

		for _, info := range infos {
			count := searchSuffixInResponse(body, info.suffix)
			if count > 0 {
				results = append(results, BreachedResult{
					Service: info.service,
					Count:   count,
				})
			}
		}
	}

	if apiFailedCount == len(prefixGroups) && len(prefixGroups) > 0 {
		return nil, fmt.Errorf("HIBP API is currently unavailable")
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
