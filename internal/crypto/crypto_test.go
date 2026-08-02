package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	pt := []byte(`{"records":{}}`)
	pw := []byte("securePassword123")
	emp := []byte("encryptedMasterPassData")

	ct, err := Encrypt(pt, pw, emp)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	gotPt, gotEmp, err := Decrypt(ct, pw)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(gotPt, pt) {
		t.Errorf("Plaintext mismatch: got %s, want %s", gotPt, pt)
	}
	if !bytes.Equal(gotEmp, emp) {
		t.Errorf("EMP mismatch: got %v, want %v", gotEmp, emp)
	}
}

func TestDecryptWrongPassword(t *testing.T) {
	pt := []byte(`{"records":{"test":{}}}`)
	pw := []byte("correctPassword")
	wrongPw := []byte("wrongPassword")

	ct, _ := Encrypt(pt, pw, nil)
	_, _, err := Decrypt(ct, wrongPw)

	if err == nil {
		t.Fatal("Expected error for wrong password, got nil")
	}
}

func TestDecryptCorruptedData(t *testing.T) {
	pt := []byte(`{"records":{}}`)
	pw := []byte("password")

	ct, _ := Encrypt(pt, pw, nil)
	ct[len(ct)-5] ^= 0xFF

	_, _, err := Decrypt(ct, pw)
	if err == nil {
		t.Fatal("Expected error for corrupted data, got nil")
	}
}

func TestDecryptEmptyPlaintext(t *testing.T) {
	pt := []byte(``)
	pw := []byte("password")

	ct, err := Encrypt(pt, pw, nil)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	got, _, err := Decrypt(ct, pw)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Expected empty plaintext, got length %d", len(got))
	}
}
