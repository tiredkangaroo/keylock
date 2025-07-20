package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

func promptText(prompt string) (string, error) {
	fmt.Printf("%s", prompt)
	text, err := stdinReader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	text = text[:len(text)-1] // remove the newline character
	return text, nil
}

func promptRequiredText(prompt string) (string, error) {
	text, err := promptText(prompt)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("input cannot be empty")
	}
	return text, nil
}

func promptPassword(prompt string) (string, error) {
	fmt.Printf("%s", prompt)
	pwd, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(pwd), nil
}

func promptRequiredPassword(prompt string) (string, error) {
	pwd, err := promptPassword(prompt)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(pwd) == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	return pwd, nil
}

func getKeyringData() (KeyringData, error) {
	dataStr, err := keyring.Get(service, currentUser.Username)
	if err != nil {
		return KeyringData{}, fmt.Errorf("failed to retrieve keyring data: %w", err)
	}
	var data KeyringData
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return KeyringData{}, fmt.Errorf("failed to unmarshal keyring data: %w", err)
	}
	return data, nil
}
func setKeyringData(krdata KeyringData) error {
	dataBytes, err := json.Marshal(krdata)
	if err != nil {
		return fmt.Errorf("failed to marshal keyring data: %w", err)
	}
	err = keyring.Set(service, currentUser.Username, string(dataBytes))
	if err != nil {
		return fmt.Errorf("failed to save keyring data: %w", err)
	}
	return nil
}

func getKey2(krdata KeyringData) (string, error) {
	code, err := promptRequiredText("5-digit code: ")
	if err != nil {
		return "", fmt.Errorf("failed to get code: %w", err)
	}
	if len(code) != 5 {
		return "", fmt.Errorf("code must be exactly 5 characters long")
	}
	codeuint, err := strconv.ParseUint(code, 10, 16)
	if err != nil {
		return "", fmt.Errorf("failed to parse code (hint: the code is a 5-digit NUMBER): %w", err)
	}
	codeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(codeBytes, uint16(codeuint))
	key2 := fmt.Sprintf("%s%x", krdata.SessionCode, codeBytes)
	return key2, nil
}
