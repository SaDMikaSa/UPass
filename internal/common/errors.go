package common

import "errors"

var (
	ErrNotFound     = errors.New("record not found")
	ErrDuplicate    = errors.New("record with this service already exists")
	ErrCanceled     = errors.New("canceled")
	ErrFileExists   = errors.New("vault file already exists")
	ErrNotEmpty     = errors.New("entered data must not be empty")
	ErrLocked       = errors.New("vault is locked")
	ErrNotMatched   = errors.New("passwords do not match")
	ErrCharsetEmpty = errors.New("charset is empty")

	ErrVaultLockedByAnother = errors.New("vault is locked by another instance")
	ErrFileTooShortForEMP   = errors.New("file too short for encrypted master password")
	ErrInvalidVaultFile     = errors.New("not a valid UPass vault file")
	ErrReadVaultFile        = errors.New("read vault file")

	ErrCreateAES     = errors.New("create aes")
	ErrCreateGCM     = errors.New("create gcm")
	ErrGenerateNonce = errors.New("generate nonce")
	ErrGenerateSalt  = errors.New("generate salt")
)
