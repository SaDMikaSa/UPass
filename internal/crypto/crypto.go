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

func DeriveKey(password []byte, salt []byte) ([]byte, error) {
	key := argon2.IDKey(password, salt, common.Argon2Time, common.Argon2Memory, common.Argon2Threads, 32)

	if err := common.LockMemory(key); err != nil {
		common.ZeroBytes(key)
		return nil, fmt.Errorf("could not lock memory for derived key: %w", err)
	}

	return key, nil
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

	key, err := DeriveKey(password, salt)
	defer common.ZeroBytes(key)
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrCreateAES, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrCreateGCM, err)
	}

	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrGenerateNonce, err)
	}
	defer common.ZeroBytes(nonce)

	if err := common.LockMemory(nonce); err != nil {
		fmt.Printf("Warning: could not lock memory for nonce: %v\n", err)
	}

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

	// TODO: MIGRATION

	// type Migration func(data []byte) ([]byte, error)
	// var migrations = map[byte]Migration{ 1: migrateV1toV2 }
	if version != common.VaultVersion {
		return nil, nil, fmt.Errorf("unsupported vault version: %d", version)
	}
	fileTime := fullData[5]
	fileMemory := binary.BigEndian.Uint32(fullData[6:10])
	empLen := binary.BigEndian.Uint16(fullData[10:12])

	if uint32(fileTime) < common.Argon2Time {
		return nil, nil, fmt.Errorf("security check failed: KDF time parameter (%d) is below minimum required (%d). File may be tampered", fileTime, common.Argon2Time)
	}
	if fileMemory < common.Argon2Memory {
		return nil, nil, fmt.Errorf("security check failed: KDF memory parameter (%d) is below minimum required (%d). File may be tampered", fileMemory, common.Argon2Memory)
	}

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

	key, err := DeriveKeyWithParams(password, salt, uint32(fileTime), uint32(fileMemory), common.Argon2Threads)
	defer common.ZeroBytes(key)
	if err != nil {
		return nil, nil, fmt.Errorf("derive key: %w", err)
	}

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

func DeriveKeyWithParams(password []byte, salt []byte, time uint32, memory uint32, threads uint8) ([]byte, error) {
	key := argon2.IDKey(password, salt, time, memory, threads, 32)

	if err := common.LockMemory(key); err != nil {
		fmt.Printf("Warning: could not lock memory for derived key: %v\n", err)
	}

	return key, nil
}
