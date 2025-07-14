package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall"

	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

func performRequest[T any](req *http.Request, method, path string, body any) (resp *http.Response, v T, err error) {
	req.Method = method
	if req.URL == nil {
		req.URL = &url.URL{Scheme: "http", Host: SERVER, Path: path}
	} else {
		req.URL.Path = path
	}
	var reqbody io.ReadCloser
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, v, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqbody = io.NopCloser(bytes.NewReader(data))
	}
	req.Body = reqbody

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, v, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, v, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, v, fmt.Errorf("error from server: %s", respbody)
	}

	err = json.Unmarshal(respbody, &v)
	if err != nil {
		return nil, v, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return
}

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
