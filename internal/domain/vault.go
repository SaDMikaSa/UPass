package domain

import (
	"strings"

	"github.com/SaDMikaSa/UPass/internal/common"
)

// cloneRecord makes a deep copy of a Record, allocating new byte slices
// for service, login, password and note to avoid accidental sharing of memory.
func cloneRecord(src Record) Record {
	dst := Record{}
	if src.Service != nil {
		dst.Service = make([]byte, len(src.Service))
		copy(dst.Service, src.Service)
	}
	if src.Login != nil {
		dst.Login = make([]byte, len(src.Login))
		copy(dst.Login, src.Login)
	}
	if src.Password != nil {
		dst.Password = make([]byte, len(src.Password))
		copy(dst.Password, src.Password)
	}
	if src.Note != nil {
		dst.Note = make([]byte, len(src.Note))
		copy(dst.Note, src.Note)
	}
	return dst
}

// Search looks up a Record by service name (case-insensitive) and returns a
// cloned copy to preserve the original data in the vault map.
func Search(vault Vault, service []byte) (Record, error) {
	key := strings.ToLower(string(service))

	rec, exists := vault.Records[key]
	if !exists {
		return Record{}, common.ErrNotFound
	}

	return cloneRecord(rec), nil
}

func Add(vault Vault, record Record) (Vault, error) {
	key := strings.ToLower(string(record.Service))
	if _, exists := vault.Records[key]; exists {
		return Vault{}, common.ErrDuplicate
	}

	newRecords := make(map[string]Record, len(vault.Records)+1)
	for k, v := range vault.Records {
		newRecords[k] = cloneRecord(v)
	}

	newRecords[key] = cloneRecord(record)

	return Vault{
		Records:             newRecords,
		EncryptedMasterPass: vault.EncryptedMasterPass,
	}, nil
}

func Delete(vault Vault, service []byte) (Vault, error) {
	key := strings.ToLower(string(service))
	_, exists := vault.Records[key]
	if !exists {
		return Vault{}, common.ErrNotFound
	}

	orig := vault.Records[key]
	common.ZeroBytes(orig.Service)
	common.ZeroBytes(orig.Login)
	common.ZeroBytes(orig.Password)
	common.ZeroBytes(orig.Note)

	vault.Records[key] = orig

	newRecords := make(map[string]Record, len(vault.Records)-1)
	for k, v := range vault.Records {
		if k != key {
			newRecords[k] = cloneRecord(v)
		}
	}

	return Vault{
		Records:             newRecords,
		EncryptedMasterPass: vault.EncryptedMasterPass,
	}, nil
}

func Edit(vault Vault, service []byte, newRecord Record) (Vault, error) {
	oldKey := strings.ToLower(string(service))
	newKey := strings.ToLower(string(newRecord.Service))

	_, exists := vault.Records[oldKey]
	if !exists {
		return Vault{}, common.ErrNotFound
	}

	if oldKey != newKey {
		if _, exists := vault.Records[newKey]; exists {
			return Vault{}, common.ErrDuplicate
		}
	}

	oldRec := vault.Records[oldKey]
	common.ZeroBytes(oldRec.Service)
	common.ZeroBytes(oldRec.Login)
	common.ZeroBytes(oldRec.Password)
	common.ZeroBytes(oldRec.Note)
	vault.Records[oldKey] = oldRec

	newRecords := make(map[string]Record, len(vault.Records))
	for k, v := range vault.Records {
		if k != oldKey {
			newRecords[k] = cloneRecord(v)
		}
	}

	newRecords[newKey] = cloneRecord(newRecord)

	return Vault{
		Records:             newRecords,
		EncryptedMasterPass: vault.EncryptedMasterPass,
	}, nil
}
