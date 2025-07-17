package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"

	"github.com/tiredkangaroo/keylock/api"
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

	resp, err := api.PerformRequest[*api.NewAccountResponse](SERVER, &api.NewAccountRequest{
		Body: api.NewAccountRequestBody{
			Name:           username,
			MasterPassword: mp,
		},
	})
	if err != nil {
		return fmt.Errorf("issue with create account: %w", err)
	}

	fmt.Printf("\nAccount created successfully! Your user ID is %d.\n", resp.Body.UserID)

	krdata := KeyringData{
		UserID:       resp.Body.UserID,
		SessionToken: resp.Cookies.Session,
		SessionCode:  resp.Body.SessionCode,
	}
	err = setKeyringData(krdata)
	if err != nil {
		return fmt.Errorf("failed to save session code to keyring: %w", err)
	}

	fmt.Printf("Please remember this code in order to use this session: %s\n", resp.Body.Code)

	return nil
}

func me() error {
	fmt.Printf("Hello, %s.\n", currentUser.Username)
	krdata, err := getKeyringData()
	if err != nil {
		return fmt.Errorf("failed to get keyring data (hint: make sure you're signed in): %w", err)
	}
	fmt.Printf("Your user ID is %d.\n", krdata.UserID)
	return nil
}

func savePassword() error {
	krdata, err := getKeyringData()
	if err != nil {
		return fmt.Errorf("failed to get keyring data (hint: make sure you're signed in): %w", err)
	}
	if krdata.SessionCode == "" {
		return fmt.Errorf("session code is empty, please sign up or log in again")
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

	_, err = api.PerformRequest[*api.NewPasswordResponse](SERVER, &api.NewPasswordRequest{
		Cookies: api.NewPasswordRequestCookies{
			Session: krdata.SessionToken,
		},
		Body: api.NewPasswordRequestBody{
			Name:  name,
			Key2:  key2,
			Value: pwd,
		},
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
	if krdata.SessionCode == "" {
		return fmt.Errorf("session code is empty, please sign up or log in again")
	}
	key2, err := getKey2(krdata)
	if err != nil {
		return fmt.Errorf("failed to get key2: %w", err)
	}

	name, err := promptRequiredText("name of password (usually a website or service plus the account name, e.g. 'google-ajinest6'): ")
	if err != nil {
		return fmt.Errorf("failed to get name: %w", err)
	}

	resp, err := api.PerformRequest[*api.RetrievePasswordResponse](SERVER, &api.RetrievePasswordRequest{
		Cookies: api.RetrievePasswordRequestCookies{
			Session: krdata.SessionToken,
		},
		Body: api.RetrievePasswordRequestBody{
			UserID: krdata.UserID,
			Name:   name,
			Key2:   key2,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve password: %w", err)
	}

	fmt.Printf("Password for '%s': %s\n", name, resp.Body.Value)
	return nil
}

func listPasswords() error {
	krdata, err := getKeyringData()
	if err != nil {
		return fmt.Errorf("failed to get keyring data (hint: make sure you're signed in): %w", err)
	}
	if krdata.SessionToken == "" {
		return fmt.Errorf("session code is empty, please sign up or log in again")
	}

	resp, err := api.PerformRequest[*api.ListPasswordsResponse](SERVER, &api.ListPasswordsRequest{
		Cookies: api.ListPasswordsRequestCookies{
			Session: krdata.SessionToken,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to list passwords: %w", err)
	}
	fmt.Println("Your passwords:")
	for _, pwd := range resp.Body.Passwords {
		fmt.Printf("- %s (id: %d, created on: %s)\n", pwd.Name, pwd.ID, formatTime(pwd.CreatedAt))
	}
	return nil
}
