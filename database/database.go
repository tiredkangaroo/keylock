package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tiredkangaroo/keylock/utils"
)

var enc_key = os.Getenv("DATABASE_ENCRYPTION_KEY")

type DB struct {
	sql *sql.DB
}

func (db *DB) Close() error {
	if db.sql != nil {
		return db.sql.Close()
	}
	return nil
}

// encryption flow: decrypt the database user key with the enc_key, concat that decrypted key with the keylock key passed by the user, encrypt the password with that key
// decryption flow: decrypt the database user key with the enc_key, concat that decrypted key with the keylock key passed by the user, decrypt the password with that key
// algorithims:
// - database user specific key / aka key1 (encrypted with the enc_key): aes-256-gcm
// - keylock key / aka key 2 (passed by the user): pbkdf2 with sha256, 45000 iterations, 32 byte key length, salt in database (user-specific)
// - key for the password (encrypted with key1 + key2 concat): aes-256-gcm
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	Key1      []byte `json:"-"`
	Key1Nonce []byte `json:"-"`
	Key2Salt  []byte `json:"-"`

	CreatedAt string `json:"created_at"`
}

type Password struct {
	ID     int    `json:"id"`
	UserID int    `json:"user_id"`
	Name   string `json:"name"`

	Value     []byte `json:"-"`
	Key2Nonce []byte `json:"-"`

	CreatedAt string `json:"created_at"`
}

func Database() (*DB, error) {
	db := &DB{}
	sql, err := sql.Open("sqlite3", utils.ConfigFile("keylock.db"))
	if err != nil {
		return nil, err
	}
	db.sql = sql

	userTableCreate := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		key1 BLOB NOT NULL,
		key1_nonce BLOB NOT NULL,
		key2_salt BLOB NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	passwordTableCreate := `CREATE TABLE IF NOT EXISTS passwords (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		value BLOB NOT NULL,
		key2_nonce BLOB NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, name)
	)`
	_, err = db.sql.Exec(userTableCreate)
	if err != nil {
		return nil, fmt.Errorf("stmt 1: %w", err)
	}
	_, err = db.sql.Exec(passwordTableCreate)
	if err != nil {
		return nil, fmt.Errorf("stmt 2: %w", err)
	}
	// create index for faster lookups
	_, err = db.sql.Exec("CREATE INDEX IF NOT EXISTS idx_user_name ON users(name)")
	if err != nil {
		return nil, fmt.Errorf("stmt 3: %w", err)
	}
	_, err = db.sql.Exec("CREATE INDEX IF NOT EXISTS idx_password_user_id ON passwords(user_id)")
	if err != nil {
		return nil, fmt.Errorf("stmt 4: %w", err)
	}
	_, err = db.sql.Exec("CREATE INDEX IF NOT EXISTS idx_password_name ON passwords(name)")
	if err != nil {
		return nil, fmt.Errorf("stmt 5: %w", err)
	}

	return db, nil
}
