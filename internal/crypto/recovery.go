package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
)

// GenerateRecoveryKey returns a new cryptographically secure random 32‑byte
// recovery key. The encoded form of this key is shown to the user at init
// and can be used to recover the master password.
func GenerateRecoveryKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("generate recovery key: %w", err)
	}
	return key, nil
}

// EncryptMasterPassword encrypts the raw master password with a key derived
// from the provided recoveryKey and returns a blob containing salt|nonce|ciphertext.
func EncryptMasterPassword(masterPassword []byte, recoveryKey []byte) ([]byte, error) {

	salt, err := GenerateSalt()
	if err != nil {
		return nil, err
	}

	derivedKey, err := DeriveKey(recoveryKey, salt)
	defer common.ZeroBytes(derivedKey)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrCreateAES, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrCreateGCM, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrGenerateNonce, err)
	}

	ciphertext := gcm.Seal(nil, nonce, masterPassword, nil)

	result := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// DecryptMasterPassword decrypts the blob produced by EncryptMasterPassword
// and returns the original master password if the provided recovery key is valid.
func DecryptMasterPassword(encrypted []byte, recoveryKey []byte) ([]byte, error) {

	if len(encrypted) < common.SaltSize+12 {
		return nil, fmt.Errorf("encrypted data too short")
	}

	salt := encrypted[:common.SaltSize]
	nonce := encrypted[common.SaltSize : common.SaltSize+12]
	ciphertext := encrypted[common.SaltSize+12:]

	derivedKey, err := DeriveKey(recoveryKey, salt)
	defer common.ZeroBytes(derivedKey)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrCreateAES, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrCreateGCM, err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid recovery key")
	}

	return plaintext, nil
}

// EncodeRecoveryKey returns a base64 string representation of the recovery key
// suitable for copying and storing by the user.
func EncodeRecoveryKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

func DecodeRecoveryKey(encoded []byte) ([]byte, error) {
	encoded = bytes.TrimSpace(encoded)

	key, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return nil, fmt.Errorf("invalid recovery key format")
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("recovery key must be 32 bytes, got %d", len(key))
	}
	return key, nil
}
