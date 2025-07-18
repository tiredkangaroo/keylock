package vault

import (
	"context"
	"errors"
	"log/slog"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/tiredkangaroo/keylock/config"
)

// fields in vault:
// keylock/redis - username
// keylock/redis - password
// keylock/psql - username
// keylock/psql - password
// keylock/encryption - key (enc_key) as hex encoded string 32 bytes/64 chars

var (
	ErrWrongType      = errors.New("wrong type for key in vault")
	ErrSubkeyNotFound = errors.New("subkey not found in vault")
)

type Vault struct {
	client *vault.Client
}

var v *Vault

func Init() {
	cfg := vault.DefaultConfig()
	cfg.Address = config.DefaultConfig.Vault.Address
	cfg.Timeout = time.Duration(config.DefaultConfig.Vault.Timeout) * time.Second
	cfg.MinRetryWait = time.Duration(config.DefaultConfig.Vault.RetryWaitMin) * time.Millisecond
	cfg.MaxRetryWait = time.Duration(config.DefaultConfig.Vault.RetryWaitMin+1500) * time.Millisecond // range of 1.5 seconds
	cfg.MaxRetries = config.DefaultConfig.Vault.RetryMax

	client, err := vault.NewClient(cfg)
	if err != nil {
		panic("new vault client: " + err.Error())
	}

	v = &Vault{
		client: client,
	}

	token := config.DefaultConfig.Vault.Token
	client.SetToken(token)
}

func getSecretField[T any](path, key, subkey string) (T, error) {
	var zero T // zero value of type T
	secret, err := v.client.KVv2(path).Get(context.Background(), key)
	if err != nil {
		return zero, err
	}
	v, ok := secret.Data[subkey]
	if !ok {
		return zero, ErrSubkeyNotFound
	}

	vt, ok := v.(T)
	if !ok {
		return zero, ErrWrongType
	}
	return vt, nil
}

func mustGetSecretField[T any](path, key, subkey string) T {
	v, err := getSecretField[T](path, key, subkey)
	if err != nil {
		slog.Warn("must failed (using zero value)", "err", err.Error(), "path", path, "key", key, "subkey", subkey)
		var zero T
		return zero
	}
	return v
}

func GetRedisUsername() string {
	return mustGetSecretField[string](config.DefaultConfig.Vault.Path, "redis", "username")
}
func GetRedisPassword() string {
	return mustGetSecretField[string](config.DefaultConfig.Vault.Path, "redis", "password")
}
func GetPostgresUsername() string {
	return mustGetSecretField[string](config.DefaultConfig.Vault.Path, "psql", "username")
}
func GetPostgresPassword() string {
	return mustGetSecretField[string](config.DefaultConfig.Vault.Path, "psql", "password")
}
func GetEncryptionKey() string {
	return mustGetSecretField[string](config.DefaultConfig.Vault.Path, "encryption", "key")
}
