package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/tiredkangaroo/keylock/api"
	"github.com/zalando/go-keyring"
)

const SERVER = "localhost:8755"

var stdinReader = bufio.NewReader(os.Stdin)

func signup() error {
	username, err := promptRequiredText("choose a username: ")
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}

	mp, err := promptRequiredPassword("choose a master password: ")
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	_, resp, err := performRequest[api.NewAccountResponse](new(http.Request), http.MethodPost, "/api/accounts/new", api.NewAccountRequest{
		Name:           username,
		MasterPassword: mp,
	})
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	fmt.Printf("\nAccount created successfully! Your user ID is %d.\n", resp.UserID)

	krdata, _ := json.Marshal(KeyringData{
		UserID:      resp.UserID,
		SessionCode: resp.SessionCode,
	})
	err = keyring.Set(service, currentUser.Username, string(krdata))
	if err != nil {
		return fmt.Errorf("failed to save session code to keyring: %w", err)
	}

	fmt.Printf("Please remember this code in order to use this session: %s\n", resp.Code)

	return nil
}

func me() error {
	fmt.Printf("Hello, %s.\n", currentUser.Username)
	keyringData, err := keyring.Get(service, currentUser.Username)
	if err != nil {
		return fmt.Errorf("failed to retrieve keyring data: %w", err)
	}
	var data KeyringData
	if err := json.Unmarshal([]byte(keyringData), &data); err != nil {
		return fmt.Errorf("failed to unmarshal keyring data: %w", err)
	}
	fmt.Printf("Your user ID is %d.\n", data.UserID)
	println("Signup successful!")
	return nil
}

func savePassword() error {
	krdata, err := getKeyringData()
	if err != nil {
		return fmt.Errorf("failed to get keyring data (hint: make sure you're signed in): %w", err)
	}
	code, err := promptRequiredText("5-digit code: ")
	if err != nil {
		return fmt.Errorf("failed to get code: %w", err)
	}
	if len(code) != 5 {
		return fmt.Errorf("code must be exactly 5 characters long")
	}
	codeuint, err := strconv.ParseUint(code, 10, 16)
	if err != nil {
		return fmt.Errorf("failed to parse code (hint: the code is a 5-digit NUMBER): %w", err)
	}
	codeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(codeBytes, uint16(codeuint))
	key2 := fmt.Sprintf("%s%x", krdata.SessionCode, codeBytes)

	name, err := promptRequiredText("name of password (usually a website or service plus the account name, e.g. 'google-ajinest6': ")
	if err != nil {
		return fmt.Errorf("failed to get name: %w", err)
	}
	pwd, err := promptRequiredPassword("password: ")
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	_, _, err = performRequest[api.NewPasswordResponse](new(http.Request), http.MethodPost, "/api/passwords/new", api.NewPasswordRequest{
		UserID: krdata.UserID,
		Name:   name,
		Key2:   key2,
		Value:  pwd,
	})
	if err != nil {
		return fmt.Errorf("failed to save password: %w", err)
	}
	fmt.Printf("\nPassword for '%s' saved successfully!\n", name)
	return nil
}

func retrievePassword() error {
	krdata, err := getKeyringData()
	if err != nil {
		return fmt.Errorf("failed to get keyring data (hint: make sure you're signed in): %w", err)
	}
	key2, err := getKey2(krdata)
	if err != nil {
		return fmt.Errorf("failed to get key2: %w", err)
	}

	name, err := promptRequiredText("name of password (usually a website or service plus the account name, e.g. 'google-ajinest6'): ")
	if err != nil {
		return fmt.Errorf("failed to get name: %w", err)
	}

	_, resp, err := performRequest[api.RetrievePasswordResponse](new(http.Request), http.MethodPost, "/api/passwords/retrieve", api.RetrievePasswordRequest{
		UserID: krdata.UserID,
		Name:   name,
		Key2:   key2,
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve password: %w", err)
	}
	fmt.Printf("Password for '%s': %s\n", name, resp.Value)
	return nil
}
