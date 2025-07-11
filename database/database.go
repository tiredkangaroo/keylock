package database

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tiredkangaroo/keylock/utils"
)

var enc_key = make([]byte, 32) // 32 bytes for aes-256-gcm

func init() {
	base32.StdEncoding.Decode(enc_key, []byte(os.Getenv("KEY1_ENCRYPTION_KEY")))
}

type DB struct {
	sql *sql.DB
}

func (db *DB) Close() error {
	if db.sql != nil {
		return db.sql.Close()
	}
	return nil
}

// encryption flow: decrypt the database user key with the enc_key, hkdf that decrypted key with the keylock key passed by the user, encrypt the password with that key
// decryption flow: decrypt the database user key with the enc_key, hkdf that decrypted key with the keylock key passed by the user, decrypt the password with that key
// algorithims:
// - database user specific key / aka key1 (encrypted with the enc_key): aes-256-gcm
// - keylock key / aka key 2 (passed by the user): pbkdf2 with sha256, 100000 iterations, 32 byte key length, salt in database (user-specific)
// - key for the password (encrypted with key1 + key2 hkdf'd): aes-256-gcm
type User struct {
	ID   int64  `json:"id"`
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
		value_nonce BLOB NOT NULL,
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

// SaveUser saves a user to the database (oh great explanation, i know).
func (db *DB) SaveUser(u *User) error {
	// name is literally the only thing we're taking from the passed in struct
	// we'll generate the key1, key1_nonce, and key2_salt here
	randoms := make([]byte, 16+12+16) // 16 for key1, 12 for key1_nonce, 16 for key2_salt

	_, err := rand.Read(randoms) // err is never returned, program "crashes irrecoverably" on error ?? 💔
	if err != nil {
		return fmt.Errorf("generating randoms: %w", err)
	}

	key1_raw := randoms[:16]
	key1_nonce := randoms[16:28]
	key2_salt := randoms[28:]

	key_1, err := utils.Encrypt(enc_key, key1_nonce, key1_raw)
	if err != nil {
		return fmt.Errorf("encrypting key1: %w", err) // NOTE: this should NEVER leak ANY of the keys (i hope), imagine if it did and my app js has a vulnerability wide open and im writing the fact that i know about this vulnerability in the code comments, that would be great wouldn't it? :D
	}

	// id and created_at are defaulted by the database, so we don't need to set them
	stmt := `INSERT INTO users (name, key1, key1_nonce, key2_salt) VALUES (?, ?, ?, ?)`
	result, err := db.sql.Exec(stmt, u.Name, key_1, key1_nonce, key2_salt)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	var err error
	u.ID, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	return nil
}

func (db *DB) SavePassword(code string, pwd *Password) error {
	// steps:
	// - we need to get the user by id
	// - from the user, we need to get the key1, decrypt it with enc_key and key1_nonce, and get the key2_salt
	// - pbkdf2 the code with the key2_salt and the right configuration to get the key2
	// - encrypt the password with key1 + key2 hkdf'd
	// - save
	stmt := `SELECT key1, key1_nonce, key2_salt FROM users WHERE id = ?`
	var key1_raw, key1_nonce, key2_salt []byte
	err := db.sql.QueryRow(stmt, pwd.UserID).Scan(&key1_raw, &key1_nonce, &key2_salt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with id %d not found", pwd.UserID)
		}
		return fmt.Errorf("querying user: %w", err)
	}

	key1, err := utils.Decrypt(enc_key, key1_nonce, key1_raw) // decrypt key1
	if err != nil {
		return fmt.Errorf("decrypting key1: %w", err)
	}
	key2, err := pbkdf2.Key(sha256.New, code, key2_salt, 1e5, 32)
	if err != nil {
		return fmt.Errorf("pbkdf2 key: %w", err)
	}
	// encrypt the password with key1 + key2 hkdf'd
	key := utils.KeyFromKeys(key1, key2)

	nonce := make([]byte, 12) // 12 bytes for aes-256-gcm nonce
	_, err = rand.Read(nonce)
	if err != nil {
		return fmt.Errorf("generating nonce: %w", err)
	}

	encryptedValue, err := utils.Encrypt(key, nonce, pwd.Value)
	if err != nil {
		return fmt.Errorf("encrypting password: %w", err)
	}
	stmt = `INSERT INTO passwords (user_id, name, value, value_nonce) VALUES (?, ?, ?, ?)`
	_, err = db.sql.Exec(stmt, pwd.UserID, pwd.Name, encryptedValue, nonce)
}

func (db *DB) RetrievePassword(pwd *Password) error {

}
