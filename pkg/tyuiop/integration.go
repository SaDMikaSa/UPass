package tyuiop

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *HIBPClient) CheckPassword(ctx context.Context, password []byte) (*PwnedCheckResult, error) {
	hash := sha1.Sum(password)
	hashHex := strings.ToUpper(hex.EncodeToString(hash[:]))

	prefix := hashHex[:5]
	suffix := hashHex[5:]

	url := fmt.Sprintf("%s/range/%s", c.baseURL, prefix)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Add-Padding", "true")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HIBP API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		if parts[0] == suffix {
			count := 0
			_, err := fmt.Sscanf(parts[1], "%d", &count)
			if err != nil {
				return nil, fmt.Errorf("failed to parse breach count: %w", err)
			}
			return &PwnedCheckResult{
				IsPwned:     true,
				BreachCount: count,
				Service:     "HIBP",
			}, nil
		}
	}

	return &PwnedCheckResult{
		IsPwned:     false,
		BreachCount: 0,
		Service:     "HIBP",
	}, nil
}

func (a *PwnedAggregator) CheckPassword(ctx context.Context, password []byte) (*PwnedCheckResult, error) {
	for _, checker := range a.checkers {
		result, err := checker.CheckPassword(ctx, password)
		if err != nil {
			continue
		}
		if result.IsPwned {
			return result, nil
		}
	}

	return &PwnedCheckResult{
		IsPwned: false,
		Service: "all",
	}, nil
}
