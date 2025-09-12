package storage

// StorageManager manages password records using a database store.
type StorageManager struct {
	store *DbStore
}

// NewStorageManager creates a new StorageManager instance with the specified SQLite database path.
// It initializes the underlying database store and returns an error if initialization fails.
func NewStorageManager(dbPath string) (*StorageManager, error) {
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		return nil, err
	}
	return &StorageManager{store: store}, nil
}

// AddPassword stores a new password record for the specified service and username.
func (sm *StorageManager) AddPassword(service, username, password string) error {
	record := PasswordRecord{
		Service:  service,
		Username: username,
		Password: password,
	}
	return sm.store.AddRecord(record)
}

// GetPassword retrieves a password record by service name.
func (sm *StorageManager) GetPassword(service string) (PasswordRecord, error) {
	return sm.store.GetRecord(service)
}

// DeletePassword removes a password record by service name.
func (sm *StorageManager) DeletePassword(service string) error {
	return sm.store.DeleteRecord(service)
}

// Close terminates the underlying database connection.
func (sm *StorageManager) Close() error {
	return sm.store.Close()
}
