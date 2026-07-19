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

// Add inserts a new Record into the Vault. Service names are stored in
// lowercase, and duplicates return common.ErrDuplicate.
func Add(vault Vault, record Record) (Vault, error) {
	key := strings.ToLower(string(record.Service))

	if vault.Records == nil {
		vault.Records = make(map[string]Record)
	}

	if _, exists := vault.Records[key]; exists {
		return Vault{}, common.ErrDuplicate
	}

	vault.Records[key] = cloneRecord(record)
	return vault, nil
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

// Delete removes the record for the specified service from the vault
// and zeroes the memory used by the record fields before deletion.
func Delete(vault Vault, service []byte) (Vault, error) {
	key := strings.ToLower(string(service))

	rec, exists := vault.Records[key]
	if !exists {
		return Vault{}, common.ErrNotFound
	}

	common.ZeroBytes(rec.Service)
	common.ZeroBytes(rec.Login)
	common.ZeroBytes(rec.Password)
	common.ZeroBytes(rec.Note)

	delete(vault.Records, key)
	return vault, nil
}

// Edit updates an existing record identified by service. If the service key
// changes (rename), it ensures the new key does not collide with an existing
// record. The old record bytes are zeroed before being replaced.
func Edit(vault Vault, service []byte, newRecord Record) (Vault, error) {
	oldKey := strings.ToLower(string(service))
	newKey := strings.ToLower(string(newRecord.Service))

	oldRec, exists := vault.Records[oldKey]
	if !exists {
		return Vault{}, common.ErrNotFound
	}

	if oldKey != newKey {
		if _, exists := vault.Records[newKey]; exists {
			return Vault{}, common.ErrDuplicate
		}
	}

	common.ZeroBytes(oldRec.Service)
	common.ZeroBytes(oldRec.Login)
	common.ZeroBytes(oldRec.Password)
	common.ZeroBytes(oldRec.Note)

	if oldKey != newKey {
		delete(vault.Records, oldKey)
	}

	vault.Records[newKey] = cloneRecord(newRecord)
	return vault, nil
}
