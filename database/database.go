package database

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tiredkangaroo/keylock/utils"
)

// BIG NOTE: code should be the same for every password for a user! we do not want to store the code in the db.
// my first idea is that we check for other passwords for the user, if they have the same code, we use that code to decrypt the password.
// but that's super inefficient and its 3am rn.
//
// note for ui:
// - session code + code -> key2 should be done in this manner
// - decode code (uint16) -> bytes -> hex string
// - session code hex (30 bytes or 60 chars) + code hex (4 chars) = key2 hex (64 chars)

var enc_key = make([]byte, 32) // 32 bytes for aes-256-gcm

func init() {
	enc_key_str := os.Getenv("ENCRYPTION_KEY")
	if enc_key_str == "" {
		panic("ENCRYPTION_KEY environment variable is not set")
	}
	if len(enc_key_str) != 64 {
		panic(fmt.Errorf("ENCRYPTION_KEY must be 64 characters long (32 bytes encoded), got %d characters", len(enc_key_str)))
	}
	if _, err := hex.Decode(enc_key, []byte(enc_key_str)); err != nil {
		panic(fmt.Errorf("decoding ENCRYPTION_KEY: %w", err))
	}
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

// the whole thing:
// user signs in how ever the hell they want (sms code, email code, password, 2fa, yubikey, biometrics, etc.): this has nothing to do with the secrets
// user provides a code to encrypt/decrypt the secrets
//
// a master password is pbkdfd + key2_salt'd into 32 bytes -> 30 bytes will become a session code (e.g session storage), 2 bytes will become a uint16 number
// the user must remember the uint16 number (when they're in a session)
// combining the 32 bytes will give us key2
// if not in a session (or the code is forgotten), the user must provide the master password
// any signed in session + the code can reset the master password
//
// a user in a session can encrypt new secrets with the code as well as decrypt existing secrets
//
// this master password system provides very unique security properties specifically for edge cases.
// e.g. phone is stolen and the thief knows the code (shoulder-surfing). if the session ends (e.g device restart) - it protects the secrets.
// weird case but very possible. did u know iphones that have been left locked for a 3 days auto-restart to clear out hot memory? pretty nice measure taken.
// source for that iphone statement: that one seytonic video and this reddit post i found to confirm while writing ts: https://www.reddit.com/r/cybersecurity/comments/1gs6ngg/new_apple_security_feature_reboots_iphones_after/
//
// onion layer: secret -> encrypted by key2 (layer 1) -> encrypted by key1 (layer 2)
// this allows for unilateral key1 rotation.

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`

	Key1      []byte `json:"-"`
	Key1Nonce []byte `json:"-"`
	Key2Salt  []byte `json:"-"`

	CreatedAt string `json:"created_at"`
}

type Password struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
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
		value_layer1_nonce BLOB NOT NULL,
		value_layer2_nonce BLOB NOT NULL,
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
// Expected fields:
// - Name
// - Master Password (not stored, used to provide session code and code)
func (db *DB) SaveUser(name, masterPassword string) (id int64, sessionCode string, code string, err error) {
	// name is literally the only thing we're taking from the passed in struct
	// we'll generate the key1, key1_nonce, and key2_salt here
	randoms := make([]byte, 16+12+16) // 16 for key1, 12 for key1_nonce, 16 for key2_salt

	_, err = rand.Read(randoms) // err is never returned, program "crashes irrecoverably" on error ?? 💔
	if err != nil {
		err = fmt.Errorf("generating randoms: %w", err)
		return
	}

	key1_raw := randoms[:16]
	key1_nonce := randoms[16:28]
	key2_salt := randoms[28:]

	key_1, err := utils.Encrypt(enc_key, key1_nonce, key1_raw)
	if err != nil {
		err = fmt.Errorf("encrypting key1: %w", err)
		return
	}

	// id and created_at are defaulted by the database, so we don't need to set them
	stmt := `INSERT INTO users (name, key1, key1_nonce, key2_salt) VALUES (?, ?, ?, ?)`
	result, err := db.sql.Exec(stmt, name, key_1, key1_nonce, key2_salt)
	if err != nil {
		err = fmt.Errorf("inserting user: %w", err)
		return
	}

	id, err = result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("getting last insert id: %w", err)
		return
	}

	key2, err := pbkdf2.Key(sha256.New, masterPassword, key2_salt, 1e6, 32)
	if err != nil {
		err = fmt.Errorf("pbkdf2 key: %w", err)
		return
	}
	sessionCode = hex.EncodeToString(key2[:30])     // 30 bytes for session code
	rawCode := binary.BigEndian.Uint16(key2[30:32]) // 2 bytes for code, uint16
	code = fmt.Sprintf("%05d", rawCode)             // 5 digits
	return
}

// SavePassword saves a password to the database.
// expected fields:
// - UserID
// - Value
// - Name
func (db *DB) SavePassword(userid int64, name, key2, value string) error {
	key2_decoded, err := hex.DecodeString(key2)
	if err != nil {
		return fmt.Errorf("decoding key2 with hex: %w", err)
	}

	stmt := `SELECT key1, key1_nonce FROM users WHERE id = ?` // step 1: get the user to retrieve key1
	var key1_raw, key1_nonce []byte
	err = db.sql.QueryRow(stmt, userid).Scan(&key1_raw, &key1_nonce)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with id %d not found", userid)
		}
		return fmt.Errorf("querying user: %w", err)
	}

	key1, err := utils.Decrypt(enc_key, key1_nonce, key1_raw) // step 2: decrypt key1
	if err != nil {
		return fmt.Errorf("decrypting key1: %w", err)
	}

	// step 3: generate nonces for layer1 and layer2 encryption
	nonces := make([]byte, 24)  // 12 bytes for layer1 nonce, 12 bytes for layer2 nonce
	rand.Read(nonces)           // generate random nonces
	layer1_nonce := nonces[:12] // first 12 bytes for layer1 nonce
	layer2_nonce := nonces[12:] // last 12 bytes for layer2 nonce

	// step 4: encrypt the value with key2 to make layer 1
	layer1, err := utils.Encrypt(key2_decoded, layer1_nonce, []byte(value))
	if err != nil {
		return fmt.Errorf("encrypting layer 1: %w", err)
	}
	// step 5: encrypt the layer1 with key1 to make layer 2 (final value)
	layer2, err := utils.Encrypt(key1, layer2_nonce, layer1)
	if err != nil {
		return fmt.Errorf("encrypting layer 2: %w", err)
	}

	// step 6: insert the password into the database
	stmt = `INSERT INTO passwords (user_id, name, value, value_layer1_nonce, value_layer2_nonce) VALUES (?, ?, ?, ?, ?)`
	_, err = db.sql.Exec(stmt, userid, name, layer2, layer1_nonce, layer2_nonce)
	if err != nil {
		return fmt.Errorf("inserting password: %w", err)
	}
	return nil
}

// password will be set into the value field
// expected fields:
// - user id
// - name
func (db *DB) RetrievePassword(userid int64, name, key2 string) (pwd []byte, err error) {
	// steps:
	// - decode key2 from hex to bytes
	// - get password by name and user id (value, value_layer1_nonce and value_layer2_nonce)
	// - get the user by id (key1, key1_nonce)
	// - decrypt key1 with enc_key + key1_nonce
	// - decrypt layer 2 with key1 + value_layer2_nonce
	// - decrypt layer 1 with key2 + value_layer1_nonce (secret)

	// step 0: decode key2 from hex to bytes
	key2_decoded, err := hex.DecodeString(key2)
	if err != nil {
		err = fmt.Errorf("decoding key2 with hex: %w", err)
		return
	}

	// step 1: get password and extract value, value_layer1_nonce, value_layer2_nonce
	stmt := `SELECT value, value_layer1_nonce, value_layer2_nonce FROM passwords WHERE user_id = ? AND name = ?`
	var value, value_layer1_nonce, value_layer2_nonce []byte
	err = db.sql.QueryRow(stmt, userid, name).Scan(&value, &value_layer1_nonce, &value_layer2_nonce)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("password with name %s for user id %d not found", name, userid)
		} else {
			err = fmt.Errorf("querying password: %w", err)
		}
		return
	}

	// step 2: get the user to retrieve key1 and key1_nonce
	stmt = `SELECT key1, key1_nonce FROM users WHERE id = ?`
	var key1_raw, key1_nonce []byte
	err = db.sql.QueryRow(stmt, userid).Scan(&key1_raw, &key1_nonce)
	if err != nil {
		if err == sql.ErrNoRows { // unlikely but handle it
			err = fmt.Errorf("user with id %d not found", userid)
		} else {
			err = fmt.Errorf("querying user: %w", err)
		}
		return
	}

	// step 3: decrypt key1 with enc_key and key1_nonce
	key1, err := utils.Decrypt(enc_key, key1_nonce, key1_raw) // decrypt key1
	if err != nil {
		err = fmt.Errorf("decrypting key1: %w", err)
		return
	}

	// step 4: decrypt layer 2 with key1 and value_layer2_nonce
	layer1, err := utils.Decrypt(key1, value_layer2_nonce, value) // decrypt the layer 2 to get layer 1
	if err != nil {
		err = fmt.Errorf("decrypting layer 2: %w", err)
		return
	}
	// step 5: decrypt layer 1 with key2 and value_layer1_nonce
	pwd, err = utils.Decrypt(key2_decoded, value_layer1_nonce, layer1) // decrypt the layer 1 to get the password
	if err != nil {
		err = fmt.Errorf("decrypting layer 1: %w", err)
		return
	}
	return
}
