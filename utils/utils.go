package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/sha256"
	"path/filepath"

	"github.com/tiredkangaroo/keylock/config"
)

func ConfigFile(a ...string) string {
	return filepath.Join(append([]string{config.DefaultConfig.Dirname()}, a...)...)
}

func Encrypt(key, nonce, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nil, nonce, plaintext, nil), nil
}

func Decrypt(key, nonce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func KeyFromKeys(key1, key2 []byte) []byte {
	data, err := hkdf.Key(sha256.New, append(key1, key2...), nil, "encrypting-passwords", 32) // 32 for aes 256
	if err != nil {
		panic(err)
	}
	return data
}
