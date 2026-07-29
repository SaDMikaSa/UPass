package service

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
	"github.com/SaDMikaSa/UPass/internal/crypto"
	"github.com/SaDMikaSa/UPass/internal/domain"
	"github.com/SaDMikaSa/UPass/internal/store"
)

type VaultService struct {
	filename string
	vault    domain.Vault
	unlocked bool
}

// NewVaultService creates a VaultService bound to the given filename.
// The service provides higher-level operations on the vault file (unlocking,
// CRUD for records, backups and recovery).
func NewVaultService(filename string) *VaultService {
	return &VaultService{
		filename: filename,
	}
}

// Unlock loads and decrypts the vault stored at the service's filename
// using the provided master password. On success the service becomes unlocked
// and ready for record operations.
func (s *VaultService) Unlock(password []byte) error {
	vault, err := store.Load(s.filename, password)
	if err != nil {
		return fmt.Errorf("unlock failed: %w", err)
	}
	s.vault = vault
	s.unlocked = true
	return nil
}

// AddRecord adds a new password record to the unlocked vault and persists
// the change to disk using the provided master password. Returns an error
// if the service is locked or save fails.
func (s *VaultService) AddRecord(record domain.Record, password []byte) error {
	if !s.unlocked {
		return common.ErrLocked
	}

	newVault, err := domain.Add(s.vault, record)
	if err != nil {
		return fmt.Errorf("add record: %w", err)
	}

	if err = store.Save(s.filename, newVault, password); err != nil {
		return fmt.Errorf("save after add: %w", err)
	}

	s.vault = newVault
	return nil
}

// GetRecord retrieves a record by service name from the unlocked vault.
// If the vault is locked, ErrLocked is returned.
func (s *VaultService) GetRecord(service string) (domain.Record, error) {
	if !s.unlocked {
		return domain.Record{}, common.ErrLocked
	}

	return domain.Search(s.vault, []byte(service))
}

// ListServices returns a list of service names stored in the unlocked vault.
// Services that contain a note are annotated with a trailing " *".
func (s *VaultService) ListServices() []string {
	if !s.unlocked {
		return nil
	}
	services := make([]string, 0, len(s.vault.Records))
	for key := range s.vault.Records {
		entry := key
		services = append(services, entry)
	}

	sort.Strings(services)
	return services
}

// DeleteRecord removes the record for the given service from the vault and
// writes the updated vault to disk using the provided master password.
func (s *VaultService) DeleteRecord(service string, password []byte) error {
	if !s.unlocked {
		return common.ErrLocked
	}

	newVault, err := domain.Delete(s.vault, []byte(service))
	if err != nil {
		return fmt.Errorf("delete record: %w", err)
	}

	if err = store.Save(s.filename, newVault, password); err != nil {
		return fmt.Errorf("save after delete: %w", err)
	}
	s.vault = newVault
	return nil
}

// EditRecord modifies an existing record (possibly renaming the service
// key) and persists the change using the provided master password.
func (s *VaultService) EditRecord(service string, newRecord domain.Record, password []byte) error {
	if !s.unlocked {
		return common.ErrLocked
	}

	newVault, err := domain.Edit(s.vault, []byte(service), newRecord)
	if err != nil {
		return fmt.Errorf("edit record: %w", err)
	}

	if err = store.Save(s.filename, newVault, password); err != nil {
		return fmt.Errorf("save after edit: %w", err)
	}

	s.vault = newVault
	return nil
}

// Init creates a new vault file at the service's filename using the
// provided master password. Returns ErrFileExists if the target file
// already exists.
func (s *VaultService) Init(password []byte) error {
	_, err := os.Stat(s.filename)
	if err == nil {
		return common.ErrFileExists
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("check vault file: %w", err)
	}
	return store.Save(s.filename, domain.Vault{}, password)
}

// InitWithRecovery initializes a new vault and also generates a recovery key
// that can be used to decrypt the master password if the user forgets it.
// The returned string is the encoded recovery key and must be stored by the user.
func (s *VaultService) InitWithRecovery(password []byte) (string, error) {
	_, err := os.Stat(s.filename)
	if err == nil {
		return "", common.ErrFileExists
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("check vault file: %w", err)
	}

	recoveryKey, err := crypto.GenerateRecoveryKey()
	if err != nil {
		return "", err
	}

	encryptedMaster, err := crypto.EncryptMasterPassword(password, recoveryKey)
	if err != nil {
		return "", err
	}

	vault := domain.Vault{
		Records:             make(map[string]domain.Record),
		EncryptedMasterPass: encryptedMaster,
	}

	if err := store.Save(s.filename, vault, password); err != nil {
		return "", err
	}

	return crypto.EncodeRecoveryKey(recoveryKey), nil
}

func (s *VaultService) RecoverVault(encodedKey []byte) ([]byte, error) {
	recoveryKey, err := crypto.DecodeRecoveryKey(encodedKey)
	defer common.ZeroBytes(recoveryKey)
	if err != nil {
		return nil, err
	}

	fullData, err := os.ReadFile(s.filename)
	defer common.ZeroBytes(fullData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrReadVaultFile, err)
	}

	if len(fullData) < common.BaseHeaderSize {
		return nil, fmt.Errorf("file too short")
	}
	if !bytes.Equal(fullData[:4], common.VaultMagic) {
		return nil, common.ErrInvalidVaultFile
	}

	empLen := binary.BigEndian.Uint16(fullData[10:12])
	if empLen == 0 {
		return nil, fmt.Errorf("no recovery key set for this vault")
	}

	offset := common.BaseHeaderSize
	if len(fullData) < offset+int(empLen) {
		return nil, common.ErrFileTooShortForEMP
	}

	encryptedMaster := fullData[offset : offset+int(empLen)]

	masterPassword, err := crypto.DecryptMasterPassword(encryptedMaster, recoveryKey)
	defer common.ZeroBytes(masterPassword)
	if err != nil {
		return nil, err
	}

	return masterPassword, nil
}

// ChangePassword replaces the vault master password. It updates the
// encrypted master password (EMP) using the provided encoded recovery key and
// persists the vault encrypted with the new password.
// It ensures transactional integrity: memory is only updated if disk write succeeds.
func (s *VaultService) ChangePassword(oldPassword []byte, newPassword []byte, encodedRecoveryKey []byte) error {
	if !s.unlocked {
		return common.ErrLocked
	}

	if _, err := store.Load(s.filename, oldPassword); err != nil {
		return fmt.Errorf("old password verification failed: %w", err)
	}

	recoveryKey, err := crypto.DecodeRecoveryKey(encodedRecoveryKey)
	defer common.ZeroBytes(recoveryKey)
	if err != nil {
		return err
	}

	newEncryptedMaster, err := crypto.EncryptMasterPassword(newPassword, recoveryKey)
	if err != nil {
		return err
	}

	vaultToSave := s.vault
	vaultToSave.EncryptedMasterPass = newEncryptedMaster

	if err := store.Save(s.filename, vaultToSave, newPassword); err != nil {
		return fmt.Errorf("save after password change: %w", err)
	}

	s.vault.EncryptedMasterPass = newEncryptedMaster

	return nil
}

type SearchResult struct {
	Service   string
	Login     string
	MatchedIn string // "service", "login", "note"
}

// SearchAll performs a case-insensitive search across services and,
// optionally, logins, notes and tags. It returns a slice of SearchResult describing
// where matches were found.
func (s *VaultService) SearchAll(query string, inLogin, inNote, inTag bool) []SearchResult {
	if !s.unlocked {
		return nil
	}

	var results []SearchResult
	query = strings.ToLower(query)
	queryBytes := []byte(query)

	for _, rec := range s.vault.Records {
		serviceLower := strings.ToLower(string(rec.Service))
		loginLower := bytes.ToLower(rec.Login)
		noteLower := bytes.ToLower(rec.Note)

		serviceName := string(rec.Service)
		loginName := string(rec.Login)

		if serviceLower == string(queryBytes) {
			results = append(results, SearchResult{Service: serviceName, Login: loginName, MatchedIn: "service"})
			continue
		}

		if inLogin && bytes.Contains(loginLower, queryBytes) {
			results = append(results, SearchResult{Service: serviceName, Login: loginName, MatchedIn: "login"})
			continue
		}

		if inNote && bytes.Contains(noteLower, queryBytes) {
			results = append(results, SearchResult{Service: serviceName, Login: loginName, MatchedIn: "note"})
			continue
		}

		if inTag {
			parts := strings.SplitN(serviceName, ":", 2)
			if len(parts) == 2 {
				if strings.Contains(strings.ToLower(parts[1]), query) {
					results = append(results, SearchResult{Service: serviceName, Login: loginName, MatchedIn: "tag"})
					continue
				}
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return strings.ToLower(results[i].Service) < strings.ToLower(results[j].Service)
	})

	return results
}

// HasNote reports whether the given service has a non-empty note field.
func (s *VaultService) HasNote(service string) bool {
	key := strings.ToLower(service)
	rec, exists := s.vault.Records[key]
	if !exists {
		return false
	}
	return len(rec.Note) > 0
}

// Records returns the internal records map for the unlocked vault.
// Returns nil if the service is not unlocked.
func (s *VaultService) Records() map[string]domain.Record {
	if !s.unlocked {
		return nil
	}
	return s.vault.Records
}

// Filename returns the filepath of the vault the service manages.
func (s *VaultService) Filename() string {
	return s.filename
}

// DropRecoveryAndChangePassword changes the master password and disables the
// recovery mechanism by clearing the encrypted master password from the vault.
// After this operation, the vault can no longer be recovered via `upass recover`.
func (s *VaultService) DropRecoveryAndChangePassword(newPassword []byte) error {
	if !s.unlocked {
		return common.ErrLocked
	}

	if len(s.vault.EncryptedMasterPass) > 0 {
		common.ZeroBytes(s.vault.EncryptedMasterPass)
		s.vault.EncryptedMasterPass = nil
	}

	if err := store.Save(s.filename, s.vault, newPassword); err != nil {
		return fmt.Errorf("save after password change: %w", err)
	}

	return nil
}
