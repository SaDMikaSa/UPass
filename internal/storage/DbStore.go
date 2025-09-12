package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	maxPasswordLen = 70
	maxServiceLen  = 120
	maxLoginLen    = 70
)

// DbStore represents a SQLite3-based storage implementation.
type DbStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store instance.
func NewSQLiteStore(dbPath string) (*DbStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS passwords (
		service TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		password TEXT NOT NULL
	)`

	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &DbStore{db: db}, nil
}

// AddRecord stores a password record in the database.
func (s *DbStore) AddRecord(record PasswordRecord) error {

	if len(record.Service) > maxServiceLen ||
		len(record.Username) > maxLoginLen ||
		len(record.Password) > maxPasswordLen {
		return fmt.Errorf("invalid input length")
	}

	query := `INSERT OR REPLACE INTO passwords (service, username, password) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, record.Service, record.Username, record.Password)
	return err
}

// GetRecord retrieves a password record by service name.
func (s *DbStore) GetRecord(service string) (PasswordRecord, error) {
	if len(service) > maxServiceLen {
		return PasswordRecord{}, fmt.Errorf("invalid service name length")
	}

	query := `SELECT service, username, password FROM passwords WHERE service = ?`
	var record PasswordRecord
	err := s.db.QueryRow(query, service).Scan(&record.Service, &record.Username, &record.Password)
	return record, err
}

// DeleteRecord removes a password record by service name.
func (s *DbStore) DeleteRecord(service string) error {
	query := `DELETE FROM passwords WHERE service = ?`
	_, err := s.db.Exec(query, service)
	return err
}

// Close closes the underlying database connection.
func (s *DbStore) Close() error {
	return s.db.Close()
}
