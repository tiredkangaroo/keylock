package utils

import (
	"path/filepath"

	"github.com/tiredkangaroo/keylock/config"
)

func ConfigFile(a ...string) string {
	return filepath.Join(append([]string{config.DefaultConfig.Dirname()}, a...)...)
}
