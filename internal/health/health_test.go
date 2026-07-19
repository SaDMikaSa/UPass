package health

import (
	"testing"

	"github.com/SaDMikaSa/UPass/internal/domain"
)

func TestCheckWeakPasswords(t *testing.T) {
	records := map[string]domain.Record{
		"github": {
			Service:  []byte("github"),
			Login:    []byte("user@test.com"),
			Password: []byte("123456"), // weak
		},
		"bank": {
			Service:  []byte("bank"),
			Login:    []byte("john"),
			Password: []byte("correct-horse-battery-staple"), // strong
		},
		"empty_pass": {
			Service:  []byte("empty_pass"),
			Login:    []byte("user@test.com"),
			Password: []byte(""), // CRITICAL EDGE CASE: must not panic with unsafe.String
		},
	}

	results := CheckWeakPasswords(records, 3)

	foundWeak := false
	for _, res := range results {
		if res.Service == "github" && res.Score < 3 {
			foundWeak = true
			break
		}
	}
	if !foundWeak {
		t.Errorf("expected 'github' to be flagged as weak, but it was not found in results")
	}

	for _, res := range results {
		if res.Service == "bank" {
			t.Errorf("expected 'bank' to be strong, but it was flagged as weak with score %d", res.Score)
		}
	}
}

func TestCheckDuplicatePasswords(t *testing.T) {
	records := map[string]domain.Record{
		"amazon": {
			Service:  []byte("amazon"),
			Password: []byte("samepass"),
		},
		"ebay": {
			Service:  []byte("ebay"),
			Password: []byte("samepass"),
		},
		"github": {
			Service:  []byte("github"),
			Password: []byte("different"),
		},
	}

	results := CheckDuplicatePasswords(records)

	if len(results) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(results))
	}
	if len(results[0].Services) != 2 {
		t.Errorf("expected 2 services in group, got %d", len(results[0].Services))
	}
}

func TestCheckReusedLogins(t *testing.T) {
	records := map[string]domain.Record{
		"github": {
			Service: []byte("github"),
			Login:   []byte("user@test.com"),
		},
		"gitlab": {
			Service: []byte("gitlab"),
			Login:   []byte("user@test.com"),
		},
		"bank": {
			Service: []byte("bank"),
			Login:   []byte("other@test.com"),
		},
	}

	results := CheckReusedLogins(records)

	if len(results) != 1 {
		t.Fatalf("expected 1 reused login, got %d", len(results))
	}
	if results[0].Login != "user@test.com" {
		t.Errorf("expected user@test.com, got %s", results[0].Login)
	}
	if len(results[0].Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(results[0].Services))
	}
}

func TestCheckBreached(t *testing.T) {
	count, err := CheckBreached([]byte("password123"))
	if err != nil {
		t.Skipf("hibp api unavailable, skipping: %v", err)
	}
	if count == 0 {
		t.Error("expected 'password123' to be breached, got count 0")
	}
	t.Logf("'password123' found in %d breaches", count)

	count, err = CheckBreached([]byte("xv9#mK2!pL5nQ8@rT4-wS3$bN6"))
	if err != nil {
		t.Skipf("hibp api unavailable, skipping: %v", err)
	}
	if count > 0 {
		t.Errorf("expected unique password to not be breached, got %d", count)
	}
}
