package main

import (
	"flag"
	"os/user"
)

const service = "keylock-cli"

type Command uint8

const (
	CommandUnknown Command = iota
	CommandSignup
	CommandMe
	CommandSavePassword
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
	case "save-password":
		cmd = CommandSavePassword
	default:
		cmd = CommandUnknown
	}
}

type KeyringData struct {
	UserID      int64  `json:"user_id"`
	SessionCode string `json:"session_code"`
}

func main() {
	switch cmd {
	case CommandSignup:
		if err := signup(); err != nil {
			println("Error:", err.Error())
			return
		}
	case CommandMe:
		if err := me(); err != nil {
			println("Error:", err.Error())
			return
		}
	case CommandSavePassword:
		if err := savePassword(); err != nil {
			println("Error:", err.Error())
			return
		}
	default:
		println("Unknown command. Available commands: signup, me")
	}
}
