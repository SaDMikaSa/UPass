package domain

type Record struct {
	Service  []byte
	Login    []byte
	Password []byte
	Note     []byte
}

type Vault struct {
	Records             map[string]Record
	EncryptedMasterPass []byte `json:"-"`
}
