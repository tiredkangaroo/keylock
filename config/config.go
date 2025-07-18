package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const CONFIG_FILE = "keylock.toml"

var (
	ErrUnknownKeys = fmt.Errorf("unknown keys in config file")
	ErrMissingKeys = fmt.Errorf("missing required keys in config file")
)

type Config struct {
	Addr  string `toml:"addr"`
	Debug bool   `toml:"debug"`

	Redis struct {
		Network  string `toml:"network"`
		Hostport string `toml:"hostport"`
		DB       int    `toml:"db"`
		Timeout  int64  `toml:"timeout"` // in seconds
	} `toml:"redis"`

	Postgres struct {
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		SSL      bool   `toml:"ssl"`
		Database string `toml:"database"`
	} `toml:"postgres"`

	Vault struct {
		Address      string `toml:"address"`
		Timeout      int64  `toml:"timeout"`        // in seconds
		RetryWaitMin int64  `toml:"retry_wait_min"` // in ms
		RetryMax     int    `toml:"retry_max"`      // max num retries
		Token        string `toml:"token"`          // vault token
		Path         string `toml:"path"`           // the path: usually "keylock"
	} `toml:"vault"`

	dirname string // lowercase to avoid toml, is working directory ("./.keylock") or "/home/.keylock" or similar
}

func (c *Config) Dirname() string {
	return c.dirname
}

var DefaultConfig *Config = &Config{
	Addr:    ":0",
	Debug:   false,
	dirname: ".",
}

func Init() {
	file, err := getConfigFile()
	if err != nil {
		slog.Error("failed to get config file", "error", err)
		return
	}
	defer file.Close()
	if err := os.MkdirAll(DefaultConfig.dirname, 0755); err != nil {
		slog.Error("failed to create config directory (exit)", "dirname", DefaultConfig.dirname, "error", err)
		os.Exit(1)
	}

	_, err = toml.NewDecoder(file).Decode(DefaultConfig)
	if err != nil {
		slog.Error("failed to decode config file", "error", err)
		return
	}
	slog.Info("config loaded", "addr", DefaultConfig.Addr, "debug", DefaultConfig.Debug, "dirname", DefaultConfig.dirname)
}

func getConfigFile() (*os.File, error) {
	if file, err := os.Open(CONFIG_FILE); err == nil {
		slog.Info("using config file from current directory")
		DefaultConfig.dirname = "."
		return file, nil
	} else {
		slog.Error("opening current dir config", "error", err)
	}

	execfile, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failed to get executable path (config init failed)", "error", err)
		return nil, fmt.Errorf("getting executable path: %w", err)
	}
	execdir := filepath.Join(filepath.Dir(execfile), ".keylock")
	execdirConfig := filepath.Join(execdir, CONFIG_FILE)

	if file, err := os.Open(execdirConfig); err == nil {
		slog.Info("using config file from executable directory", "path", execdirConfig)
		DefaultConfig.dirname = execdir
		return file, nil
	} else {
		slog.Error("opening exec dir config", "error", err)
	}
	slog.Info("using default config")
	return nil, nil
}
