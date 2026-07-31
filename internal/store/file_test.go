//go:build !windows

package store

import (
	"os"
	"testing"

	"github.com/SaDMikaSa/UPass/internal/domain"
)

func TestStore(t *testing.T) {
	dir := t.TempDir()
	vaultPath := dir + "/test.json"
	password := []byte("masterpass")

	vault := domain.Vault{Records: make(map[string]domain.Record)}
	record := domain.Record{
		Service:  []byte("gmail"),
		Login:    []byte("test@gmail.com"),
		Password: []byte("test12345"),
	}

	vault, err := domain.Add(vault, record)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	err = Save(vaultPath, vault, password)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	_, err = os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}

	vault, err = Load(vaultPath, password)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	rec := vault.Records["gmail"]
	if string(rec.Service) != string(record.Service) {
		t.Errorf("expected service %s, got %s", record.Service, rec.Service)
	}
	if string(rec.Login) != string(record.Login) {
		t.Errorf("expected login %s, got %s", record.Login, rec.Login)
	}
	if string(rec.Password) != string(record.Password) {
		t.Errorf("expected password %s, got %s", record.Password, rec.Password)
	}
}
