package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/tiredkangaroo/keylock/api"
	"github.com/zalando/go-keyring"
)

const SERVER = "http://localhost:8755"

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
	fmt.Println()

	resp := &api.NewAccountResponse{}
	err = api.PerformRequest(SERVER, &api.NewAccountRequest{
		Body: api.NewAccountRequestBody{
			Name:           username,
			MasterPassword: mp,
		},
	}, resp)

	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	fmt.Printf("\nAccount created successfully! Your user ID is %d.\n", resp.Body.UserID)

	krdata, _ := json.Marshal(KeyringData{
		UserID:      resp.Body.UserID,
		SessionCode: resp.Body.SessionCode,
	})
	err = keyring.Set(service, currentUser.Username, string(krdata))
	if err != nil {
		return fmt.Errorf("failed to save session code to keyring: %w", err)
	}

	fmt.Printf("Please remember this code in order to use this session: %s\n", resp.Body.Code)

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

	resp := &api.NewPasswordResponse{}
	err = api.PerformRequest(SERVER, &api.NewPasswordRequest{
		Body: api.NewPasswordRequestBody{
			UserID: krdata.UserID,
			Name:   name,
			Key2:   key2,
			Value:  pwd,
		},
	}, resp)

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

	resp := &api.RetrievePasswordResponse{}
	err = api.PerformRequest(SERVER, &api.RetrievePasswordRequest{
		Body: api.RetrievePasswordRequestBody{
			UserID: krdata.UserID,
			Name:   name,
			Key2:   key2,
		},
	}, resp)

	if err != nil {
		return fmt.Errorf("failed to retrieve password: %w", err)
	}
	fmt.Printf("Password for '%s': %s\n", name, resp.Body.Value)
	return nil
}
