package domain

import (
	"errors"
	"testing"

	"github.com/SaDMikaSa/UPass/internal/common"
)

func TestAdd(t *testing.T) {
	vault := Vault{Records: make(map[string]Record)}
	record := Record{
		Service:  []byte("github"),
		Login:    []byte("user@example.com"),
		Password: []byte("secret"),
	}

	result, err := Add(vault, record)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(result.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(result.Records))
	}

	rec := result.Records["github"]
	if string(rec.Service) != "github" {
		t.Errorf("expected github, got: %q", string(rec.Service))
	}
	if string(rec.Login) != "user@example.com" {
		t.Errorf("expected user@example.com, got: %q", string(rec.Login))
	}
	if string(rec.Password) != "secret" {
		t.Errorf("expected secret, got: %q", string(rec.Password))
	}
}

func TestAddDuplicate(t *testing.T) {
	vault := Vault{Records: make(map[string]Record)}
	record := Record{
		Service:  []byte("github"),
		Login:    []byte("user@example.com"),
		Password: []byte("secret"),
	}

	vault, err := Add(vault, record)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = Add(vault, record)
	if !errors.Is(err, common.ErrDuplicate) {
		t.Errorf("expected ErrDuplicate, got: %v", err)
	}
}

func TestSearch_Found(t *testing.T) {
	vault := Vault{Records: make(map[string]Record)}
	record := Record{
		Service:  []byte("github"),
		Login:    []byte("user@example.com"),
		Password: []byte("secret"),
	}

	vault, err := Add(vault, record)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	rec, err := Search(vault, []byte("github"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if string(rec.Service) != "github" {
		t.Errorf("expected github, got %q", string(rec.Service))
	}
}

func TestSearch_NotFound(t *testing.T) {
	vault := Vault{Records: make(map[string]Record)}

	_, err := Search(vault, []byte("gmail"))
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestDeleteEmpty(t *testing.T) {
	vault := Vault{Records: make(map[string]Record)}

	_, err := Delete(vault, []byte("github"))
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestDelete(t *testing.T) {
	vault := Vault{Records: make(map[string]Record)}
	record := Record{
		Service:  []byte("github"),
		Login:    []byte("user@example.com"),
		Password: []byte("secret"),
	}

	vault, err := Add(vault, record)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	vault, err = Delete(vault, []byte("github"))
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(vault.Records) != 0 {
		t.Errorf("expected empty vault, got: %v", vault.Records)
	}
}

func TestEdit(t *testing.T) {
	vault := Vault{Records: make(map[string]Record)}
	record := Record{
		Service:  []byte("github"),
		Login:    []byte("user@example.com"),
		Password: []byte("secret"),
	}
	newRecord := Record{
		Service:  []byte("github"),
		Login:    []byte("user@gmail.com"),
		Password: []byte("secret"),
	}

	vault, err := Add(vault, record)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	vault, err = Edit(vault, []byte("github"), newRecord)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	rec, err := Search(vault, []byte("github"))
	if err != nil {
		t.Fatalf("expected to find record, got: %v", err)
	}
	if string(rec.Login) != "user@gmail.com" {
		t.Errorf("expected login user@gmail.com, got %q", string(rec.Login))
	}
}

func TestEditNotFound(t *testing.T) {
	vault := Vault{Records: make(map[string]Record)}
	record := Record{
		Service:  []byte("github"),
		Login:    []byte("user@example.com"),
		Password: []byte("secret"),
	}

	vault, err := Add(vault, record)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = Edit(vault, []byte("gmail"), record)
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestEditRenameService(t *testing.T) {
	vault := Vault{Records: make(map[string]Record)}
	record := Record{
		Service:  []byte("github"),
		Login:    []byte("user@example.com"),
		Password: []byte("secret"),
	}
	newRecord := Record{
		Service:  []byte("gitlab"),
		Login:    []byte("user@example.com"),
		Password: []byte("secret"),
	}

	vault, err := Add(vault, record)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	vault, err = Edit(vault, []byte("github"), newRecord)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = Search(vault, []byte("github"))
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("expected ErrNotFound for old key, got: %v", err)
	}

	rec, err := Search(vault, []byte("gitlab"))
	if err != nil {
		t.Fatalf("expected to find record for new key, got: %v", err)
	}
	if string(rec.Service) != "gitlab" {
		t.Errorf("expected gitlab, got %q", string(rec.Service))
	}
}

func TestCaseInsensitive(t *testing.T) {
	vault := Vault{Records: make(map[string]Record)}
	record := Record{
		Service:  []byte("GitHub"),
		Login:    []byte("user@example.com"),
		Password: []byte("secret"),
	}

	vault, err := Add(vault, record)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	rec, err := Search(vault, []byte("gItHuB"))
	if err != nil {
		t.Fatalf("expected to find record, got: %v", err)
	}
	if string(rec.Service) != "GitHub" {
		t.Errorf("expected original case GitHub, got %q", string(rec.Service))
	}

	dupRecord := Record{
		Service:  []byte("github"),
		Login:    []byte("other@example.com"),
		Password: []byte("secret"),
	}
	_, err = Add(vault, dupRecord)
	if !errors.Is(err, common.ErrDuplicate) {
		t.Errorf("expected ErrDuplicate for case-insensitive duplicate, got: %v", err)
	}
}
