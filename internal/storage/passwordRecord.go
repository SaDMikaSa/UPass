package storage

// PasswordRecord represents a stored password entry with associated service and username.
type PasswordRecord struct {
	Service  string `json:"service"`
	Username string `json:"username"`
	Password string `json:"password"`
}
