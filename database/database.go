package database

import (
	"bytes"
	"crypto/hkdf"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/tiredkangaroo/keylock/config"
	"github.com/tiredkangaroo/keylock/utils"
	"github.com/tiredkangaroo/keylock/vault"
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

func Init() {
	enc_key_str := vault.GetEncryptionKey()
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
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type Password struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func Database() (*DB, error) {
	db := &DB{}

	sslmode := "disable"
	if config.DefaultConfig.Postgres.SSL {
		sslmode = "require"
	}
	sql, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		vault.GetPostgresUsername(),
		vault.GetPostgresPassword(),
		config.DefaultConfig.Postgres.Host,
		config.DefaultConfig.Postgres.Port,
		config.DefaultConfig.Postgres.Database,
		sslmode,
	))
	if err != nil {
		return nil, err
	}
	db.sql = sql

	userTableCreate := `CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		key1 BYTEA NOT NULL,
		key1_nonce BYTEA NOT NULL,
		key2_salt BYTEA NOT NULL,
		key2_verifier BYTEA NOT NULL,
		created_at timestamp DEFAULT CURRENT_TIMESTAMP
	)`
	passwordTableCreate := `CREATE TABLE IF NOT EXISTS passwords (
		id BIGSERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		value BYTEA NOT NULL,
		value_layer1_nonce BYTEA NOT NULL,
		value_layer2_nonce BYTEA NOT NULL,
		created_at timestamp DEFAULT CURRENT_TIMESTAMP,
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

func (db *DB) GetUserByID(id int64) (*User, error) {
	stmt := `SELECT id, name, created_at FROM users WHERE id = $1;`
	var user User
	err := db.sql.QueryRow(stmt, id).Scan(&user.ID, &user.Name, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("querying user: %w", err)
	}
	return &user, nil
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

	key2, err := pbkdf2.Key(sha256.New, masterPassword, key2_salt, 1e6, 32)
	if err != nil {
		err = fmt.Errorf("pbkdf2 key: %w", err)
		return
	}
	key2_verifier, err := hkdf.Key(sha256.New, append(key2, enc_key...), nil, "key2-verifier", 32)
	if err != nil {
		err = fmt.Errorf("hkdf key2 verifier: %w", err)
		return
	}

	// id and created_at are defaulted by the database, so we don't need to set them
	stmt := `INSERT INTO users (name, key1, key1_nonce, key2_salt, key2_verifier) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	err = db.sql.QueryRow(stmt, name, key_1, key1_nonce, key2_salt, key2_verifier).Scan(&id)
	if err != nil {
		err = fmt.Errorf("inserting user: %w", err)
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

	stmt := `SELECT key1, key1_nonce, key2_verifier FROM users WHERE id = $1;` // step 1: get the user to retrieve key1
	var key1_raw, key1_nonce, key2_verifier []byte
	err = db.sql.QueryRow(stmt, userid).Scan(&key1_raw, &key1_nonce, &key2_verifier)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with id %d not found", userid)
		}
		return fmt.Errorf("querying user: %w", err)
	}

	// step 2: verify key2
	hkdf_key2, err := hkdf.Key(sha256.New, append(key2_decoded, enc_key...), nil, "key2-verifier", 32)
	if err != nil {
		return fmt.Errorf("hkdf key2 verifier: %w", err)
	}
	if !bytes.Equal(hkdf_key2, key2_verifier) {
		return fmt.Errorf("key2 verification failed")
	}

	key1, err := utils.Decrypt(enc_key, key1_nonce, key1_raw) // step 3: decrypt key1
	if err != nil {
		return fmt.Errorf("decrypting key1: %w", err)
	}

	// step 4: generate nonces for layer1 and layer2 encryption
	nonces := make([]byte, 24)  // 12 bytes for layer1 nonce, 12 bytes for layer2 nonce
	rand.Read(nonces)           // generate random nonces
	layer1_nonce := nonces[:12] // first 12 bytes for layer1 nonce
	layer2_nonce := nonces[12:] // last 12 bytes for layer2 nonce

	// step 5: encrypt the value with key2 to make layer 1
	layer1, err := utils.Encrypt(key2_decoded, layer1_nonce, []byte(value))
	if err != nil {
		return fmt.Errorf("encrypting layer 1: %w", err)
	}
	// step 6: encrypt the layer1 with key1 to make layer 2 (final value)
	layer2, err := utils.Encrypt(key1, layer2_nonce, layer1)
	if err != nil {
		return fmt.Errorf("encrypting layer 2: %w", err)
	}

	// step 7: insert the password into the database
	stmt = `INSERT INTO passwords (user_id, name, value, value_layer1_nonce, value_layer2_nonce) VALUES ($1, $2, $3, $4, $5);`
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
	// i dont think we need to verify key2 here since if u try to decrypt it with the wrong key2, it will just return an error (gcm)
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
	stmt := `SELECT value, value_layer1_nonce, value_layer2_nonce FROM passwords WHERE user_id = $1 AND name = $2;`
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
	stmt = `SELECT key1, key1_nonce FROM users WHERE id = $1;`
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

func (db *DB) ListPasswords(userID int64) ([]Password, error) {
	stmt := `SELECT id, name, created_at FROM passwords WHERE user_id = $1;`
	rows, err := db.sql.Query(stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("querying passwords: %w", err)
	}
	defer rows.Close()
	var passwords []Password
	for rows.Next() {
		pwd, err := scanPassword(rows, userID)
		if err != nil {
			return nil, fmt.Errorf("scanning password: %w", err)
		}
		passwords = append(passwords, pwd)
	}
	return passwords, nil
}

// expects id, name, created_at
func scanPassword(rows *sql.Rows, userID int64) (Password, error) {
	pwd := Password{
		UserID: userID,
	}
	if err := rows.Scan(&pwd.ID, &pwd.Name, &pwd.CreatedAt); err != nil {
		return Password{}, fmt.Errorf("scanning password row: %w", err)
	}
	return pwd, nil
}
