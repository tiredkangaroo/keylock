package main

import (
	"flag"
	"fmt"
	"os/user"
)

const service = "keylock-cli"

type Command uint8

const (
	CommandUnknown Command = iota
	CommandSignup
	CommandMe
	CommandSavePassword
	CommandRetrievePassword
	CommandListPasswords
	CommandDebugDump
)

var cmd Command
var currentUser, _ = user.Current()

func init() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		println("Usage: keylock <command> [<args>]")
		return
	}

	switch args[0] {
	case "signup":
		cmd = CommandSignup
	case "me":
		cmd = CommandMe
	case "set-password":
		cmd = CommandSavePassword
	case "get-password":
		cmd = CommandRetrievePassword
	case "list-passwords":
		cmd = CommandListPasswords
	case "debug-dump":
		cmd = CommandDebugDump
	default:
		cmd = CommandUnknown
	}
}

type KeyringData struct {
	UserID       int64  `json:"user_id"`
	SessionToken string `json:"session_token"`
	SessionCode  string `json:"session_code"`
}

func main() {
	switch cmd {
	case CommandSignup:
		if err := signup(); err != nil {
			println("\nError:", err.Error())
		}
	case CommandMe:
		if err := me(); err != nil {
			println("\nError:", err.Error())
		}
	case CommandSavePassword:
		if err := savePassword(); err != nil {
			println("\nError:", err.Error())
		}
	case CommandRetrievePassword:
		if err := retrievePassword(); err != nil {
			println("\nError:", err.Error())
		}
	case CommandListPasswords:
		if err := listPasswords(); err != nil {
			println("\nError: ", err.Error())
		}
	case CommandDebugDump:
		// this command just dumps information
		krdata, err := getKeyringData()
		if err != nil {
			fmt.Println("retrieving keyring data failed:", err)
			return
		} else {
			fmt.Printf("user id: %d\n", krdata.UserID)
			fmt.Printf("session token: %s\n", krdata.SessionToken)
			fmt.Printf("session code: %s\n", krdata.SessionCode)
		}
		fmt.Println("listing passwords.")
		if err := listPasswords(); err != nil {
			fmt.Println("listing passwords failed:", err)
			return
		}
	default:
		println("Unknown command. Available commands: signup, me, set-password, get-password, list-passwords")
	}
}
