package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"time"

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

func FormatTime(rt string) string {
	t, err := time.Parse(time.RFC3339, rt)
	if err != nil {
		return "unparsable time: " + err.Error()
	}
	t = t.Local()

	if sameDate(t, time.Now()) { // today
		return fmt.Sprintf("Today at %02d:%02d %s (local time)", t.Hour()%12, t.Minute(), t.Format("PM"))
	}
	if sameDate(t, time.Now().AddDate(0, 0, -1)) { // yesterday
		return fmt.Sprintf("Yesterday at %02d:%02d %s (local time)", t.Hour()%12, t.Minute(), t.Format("PM"))
	}

	return fmt.Sprintf("%s %d, %d at %02d:%02d %s (local time)", t.Month().String(), t.Day(), t.Year(), t.Hour()%12, t.Minute(), t.Format("PM"))
}

func sameDate(t1, t2 time.Time) bool {
	yr1, mo1, day1 := t1.Date()
	yr2, mo2, day2 := t2.Date()
	return yr1 == yr2 && mo1 == mo2 && day1 == day2
}
