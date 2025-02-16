package main

import (
	"blink-liveview-websocket/account"
	"flag"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

var (
	email = flag.String("email", "", "Blink account email address")
)

func main() {
	flag.Usage = func() {
		fmt.Print("Usage: blink-liveview-websocket [options]\n\nOptions:\n")

		flag.PrintDefaults()
	}

	flag.Parse()

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		os.Exit(1)
	}
	pass := string(passwordBytes)
	fmt.Println()

	account.Run(email, &pass)
}
