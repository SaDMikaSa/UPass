package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/common"
	"golang.org/x/crypto/argon2"
)

// DeriveKey derives a 32-byte encryption key from the given password and salt
// using Argon2id with parameters taken from common configuration.
func DeriveKey(password []byte, salt []byte) []byte {
	return argon2.IDKey(password, salt, common.Argon2Time, common.Argon2Memory, 4, 32)
}

// GenerateSalt returns a securely generated random salt of size common.SaltSize.
// The salt is used as input to the Argon2 KDF.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, common.SaltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrGenerateSalt, err)
	}
	return salt, nil
}

// Encrypt encrypts the provided plaintext (JSON vault data) using a key
// derived from the provided password. The result is a custom binary format
// containing a header (magic, version, KDF params, EMP length and EMP), salt,
// nonce and the AES‑GCM ciphertext.
func Encrypt(plaintext []byte, password []byte, encryptedMasterPass []byte) ([]byte, error) {
	salt, err := GenerateSalt()
	if err != nil {
		return nil, err
	}

	key := DeriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrCreateAES, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrCreateGCM, err)
	}

	defer common.ZeroBytes(key)

	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrGenerateNonce, err)
	}

	defer common.ZeroBytes(nonce)

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	header := make([]byte, 0, 128)
	header = append(header, common.VaultMagic...)
	header = append(header, common.VaultVersion)

	header = append(header, byte(common.Argon2Time))

	memBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(memBytes, uint32(common.Argon2Memory))
	header = append(header, memBytes...)

	empLen := len(encryptedMasterPass)
	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, uint16(empLen))
	header = append(header, lenBytes...)

	header = append(header, encryptedMasterPass...)

	header = append(header, salt...)
	header = append(header, nonce...)
	header = append(header, ciphertext...)

	return header, nil
}

// Decrypt parses the custom vault binary format, derives the key using the
// stored KDF parameters, and returns the plaintext JSON and the encrypted
// master password (EMP) blob. It validates header magic/version and basic
// length constraints.
func Decrypt(fullData []byte, password []byte) (plaintext []byte, encryptedMasterPass []byte, err error) {
	minLen := common.BaseHeaderSize + common.SaltSize + 12
	if len(fullData) < minLen {
		return nil, nil, fmt.Errorf("file too short: got %d bytes, need at least %d", len(fullData), minLen)
	}

	if !bytes.Equal(fullData[:4], common.VaultMagic) {
		return nil, nil, common.ErrInvalidVaultFile
	}

	version := fullData[4]
	if version != common.VaultVersion {
		return nil, nil, fmt.Errorf("unsupported vault version: %d", version)
	}
	fileTime := fullData[5]
	fileMemory := binary.BigEndian.Uint32(fullData[6:10])
	empLen := binary.BigEndian.Uint16(fullData[10:12])

	var emp []byte
	offset := common.BaseHeaderSize
	if empLen > 0 {
		if len(fullData) < offset+int(empLen) {
			return nil, nil, common.ErrFileTooShortForEMP
		}
		emp = make([]byte, empLen)
		copy(emp, fullData[offset:offset+int(empLen)])
		offset += int(empLen)
	}

	if len(fullData) < offset+common.SaltSize+12 {
		return nil, nil, fmt.Errorf("file too short for encrypted data")
	}

	salt := fullData[offset : offset+common.SaltSize]
	nonce := fullData[offset+common.SaltSize : offset+common.SaltSize+12]
	ciphertext := fullData[offset+common.SaltSize+12:]

	key := DeriveKeyWithParams(password, salt, uint32(fileTime), uint32(fileMemory))
	defer common.ZeroBytes(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("create new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", common.ErrCreateGCM, err)
	}

	if len(nonce) != gcm.NonceSize() {
		return nil, nil, fmt.Errorf("nonce size mismatch")
	}

	plaintext, err = gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("wrong master password or corrupted data")
	}

	return plaintext, emp, nil
}

// DeriveKeyWithParams derives a key using explicit Argon2 parameters. This
// is used when decrypting older vaults that stored the KDF parameters in the
// file header.
func DeriveKeyWithParams(password []byte, salt []byte, time uint32, memory uint32) []byte {
	return argon2.IDKey(password, salt, time, memory, 4, 32)
}
